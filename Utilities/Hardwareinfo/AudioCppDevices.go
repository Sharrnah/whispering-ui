package Hardwareinfo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"whispering-tiger-ui/Utilities"
)

// AudioCppDeviceInfo is one adapter reported by audiocpp_server --list-devices.
// Indexes are backend-specific: Vulkan:1 and CUDA:1 need not be the same card.
type AudioCppDeviceInfo struct {
	Backend string
	Index   int
	Name    string
	Kind    string
}

var audioCppDevicePattern = regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z0-9_/-]*):(\d+)\s+"([^"]+)"(?:\s+\[([^\]]+)\])?\s*$`)

func ParseAudioCppDevices(output string) []AudioCppDeviceInfo {
	devices := make([]AudioCppDeviceInfo, 0)
	for _, match := range audioCppDevicePattern.FindAllStringSubmatch(output, -1) {
		index, err := strconv.Atoi(match[2])
		if err != nil || index < 0 {
			continue
		}
		backend := strings.ToLower(strings.TrimSpace(match[1]))
		if backend == "rocm" {
			backend = "hip"
		}
		devices = append(devices, AudioCppDeviceInfo{
			Backend: backend,
			Index:   index,
			Name:    strings.TrimSpace(match[3]),
			Kind:    strings.TrimSpace(match[4]),
		})
	}
	return devices
}

func audioCppServerFilename() string {
	if runtime.GOOS == "windows" {
		return "audiocpp_server.exe"
	}
	return "audiocpp_server"
}

func addAudioCppServerPath(paths *[]string, seen map[string]bool, path string) {
	if path == "" {
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}
	absolute, err := filepath.Abs(path)
	if err == nil {
		path = absolute
	}
	key := strings.ToLower(filepath.Clean(path))
	if !seen[key] {
		seen[key] = true
		*paths = append(*paths, path)
	}
}

func configuredAudioCppServer() (string, bool, error) {
	for _, variable := range []string{
		"WHISPERING_TIGER_AUDIOCPP_SERVER",
		"AUDIOCPP_SERVER_PATH",
	} {
		path := strings.TrimSpace(os.Getenv(variable))
		if path == "" {
			continue
		}
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			path = filepath.Join(path, audioCppServerFilename())
		}
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			return "", true, fmt.Errorf("%s does not point to audiocpp_server", variable)
		}
		return path, true, nil
	}
	return "", false, nil
}

func cachedAudioCppServers() ([]string, error) {
	if configured, present, err := configuredAudioCppServer(); present {
		if err != nil {
			return nil, err
		}
		return []string{configured}, nil
	}

	paths := make([]string, 0)
	seen := make(map[string]bool)
	workingDirectory, _ := os.Getwd()
	executable, _ := os.Executable()
	roots := []string{
		filepath.Join(workingDirectory, ".cache", "audio.cpp", "runtime"),
		filepath.Join(workingDirectory, "audioWhisper", ".cache", "audio.cpp", "runtime"),
		filepath.Join(filepath.Dir(executable), ".cache", "audio.cpp", "runtime"),
		filepath.Join(filepath.Dir(executable), "audioWhisper", ".cache", "audio.cpp", "runtime"),
	}
	filename := strings.ToLower(audioCppServerFilename())
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry == nil || entry.IsDir() {
				return nil
			}
			if strings.ToLower(entry.Name()) == filename {
				addAudioCppServerPath(&paths, seen, path)
			}
			return nil
		})
	}
	if path, err := exec.LookPath(audioCppServerFilename()); err == nil {
		addAudioCppServerPath(&paths, seen, path)
	} else if path, err := exec.LookPath("audiocpp_server"); err == nil {
		addAudioCppServerPath(&paths, seen, path)
	}
	return paths, nil
}

// GetAudioCppDevices asks every locally available runtime build for its exact
// adapter mapping. It never downloads a runtime merely to populate the UI.
func GetAudioCppDevices() ([]AudioCppDeviceInfo, error) {
	servers, err := cachedAudioCppServers()
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, fmt.Errorf("no local audiocpp_server was found")
	}

	devices := make([]AudioCppDeviceInfo, 0)
	seen := make(map[string]bool)
	for _, server := range servers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, server, "--list-devices")
		Utilities.ProcessHideWindowAttr(cmd)
		output, commandErr := cmd.CombinedOutput()
		cancel()
		parsed := ParseAudioCppDevices(string(output))
		if commandErr != nil && len(parsed) == 0 {
			continue
		}
		for _, device := range parsed {
			key := fmt.Sprintf("%s:%d", device.Backend, device.Index)
			if seen[key] {
				continue
			}
			seen[key] = true
			devices = append(devices, device)
		}
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("local audiocpp_server did not report any devices")
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Backend == devices[j].Backend {
			return devices[i].Index < devices[j].Index
		}
		return devices[i].Backend < devices[j].Backend
	})
	return devices, nil
}

// GetGraphicsDevices provides named OS adapters as a clearly marked fallback
// before an audio.cpp runtime has been installed. The runtime's own mapping is
// preferred because Vulkan/HIP indices are defined by that runtime.
func GetGraphicsDevices() []GPUInfo {
	info, err := ghw.GPU()
	if err != nil || info == nil {
		return nil
	}
	devices := make([]GPUInfo, 0, len(info.GraphicsCards))
	seen := make(map[string]bool)
	for _, card := range info.GraphicsCards {
		if card == nil || card.DeviceInfo == nil {
			continue
		}
		name := ""
		vendor := ""
		if card.DeviceInfo.Product != nil {
			name = strings.TrimSpace(card.DeviceInfo.Product.Name)
		}
		if card.DeviceInfo.Vendor != nil {
			vendor = strings.TrimSpace(card.DeviceInfo.Vendor.Name)
		}
		if name == "" {
			name = vendor
		}
		key := strings.ToLower(vendor + "\x00" + name)
		if name == "" || seen[key] {
			continue
		}
		seen[key] = true
		devices = append(devices, GPUInfo{AdapterName: name, VendorName: vendor})
	}
	return devices
}
