//go:build windows

package Utilities

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)

type applicationProcessEntry struct {
	PID        uint32
	ParentPID  uint32
	Executable string
}

func applicationProcessSnapshot() ([]applicationProcessEntry, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)

	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}
	entries := make([]applicationProcessEntry, 0, 128)
	for {
		entries = append(entries, applicationProcessEntry{
			PID:        entry.ProcessID,
			ParentPID:  entry.ParentProcessID,
			Executable: windows.UTF16ToString(entry.ExeFile[:]),
		})
		entry.Size = uint32(unsafe.Sizeof(entry))
		if err := windows.Process32Next(snapshot, &entry); err != nil {
			break
		}
	}
	return entries, nil
}

func normalizedApplicationExecutable(executable string) string {
	return strings.ToLower(filepath.Base(strings.TrimSpace(executable)))
}

// resolveApplicationProcessID treats a PID as a hint and promotes a helper
// window to the root of a same-executable process family. If the PID is stale,
// the executable is accepted only when it identifies one independent family.
func resolveApplicationProcessID(executable string, preferredPID uint32) (uint32, bool) {
	entries, err := applicationProcessSnapshot()
	if err != nil {
		return 0, false
	}
	requested := normalizedApplicationExecutable(executable)
	byPID := make(map[uint32]applicationProcessEntry, len(entries))
	for _, entry := range entries {
		byPID[entry.PID] = entry
	}

	if preferred, found := byPID[preferredPID]; found && (requested == "" || normalizedApplicationExecutable(preferred.Executable) == requested) {
		target := preferred
		visited := make(map[uint32]struct{})
		for {
			if _, seen := visited[target.PID]; seen {
				break
			}
			visited[target.PID] = struct{}{}
			parent, found := byPID[target.ParentPID]
			if !found || normalizedApplicationExecutable(parent.Executable) != normalizedApplicationExecutable(target.Executable) {
				break
			}
			target = parent
		}
		return target.PID, true
	}
	if requested == "" {
		return 0, false
	}

	currentPID := uint32(os.Getpid())
	matches := make([]applicationProcessEntry, 0, 1)
	matchingPIDs := make(map[uint32]struct{})
	for _, entry := range entries {
		if entry.PID != currentPID && normalizedApplicationExecutable(entry.Executable) == requested {
			matches = append(matches, entry)
			matchingPIDs[entry.PID] = struct{}{}
		}
	}
	roots := make([]applicationProcessEntry, 0, len(matches))
	for _, entry := range matches {
		if _, parentMatches := matchingPIDs[entry.ParentPID]; !parentMatches {
			roots = append(roots, entry)
		}
	}
	if len(roots) != 1 {
		return 0, false
	}
	return roots[0].PID, true
}

func applicationProcessTreeIDs(rootPID uint32) map[uint32]struct{} {
	entries, err := applicationProcessSnapshot()
	if err != nil {
		return nil
	}
	result := make(map[uint32]struct{})
	for _, entry := range entries {
		if entry.PID == rootPID {
			result[rootPID] = struct{}{}
			break
		}
	}
	if len(result) == 0 {
		return nil
	}
	for changed := true; changed; {
		changed = false
		for _, entry := range entries {
			if _, included := result[entry.PID]; included {
				continue
			}
			if _, parentIncluded := result[entry.ParentPID]; parentIncluded {
				result[entry.PID] = struct{}{}
				changed = true
			}
		}
	}
	return result
}

// GetApplicationProcesses returns visible top-level Windows applications.
// Targeting the top-level process (rather than an arbitrary helper process) is
// important because WASAPI can include its complete child process tree.
func GetApplicationProcesses() []ApplicationProcess {
	currentPID := uint32(os.Getpid())
	windowApplications := make(map[uint32]ApplicationProcess)

	callback := syscall.NewCallback(func(window uintptr, _ uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(window)
		if visible == 0 {
			return 1
		}
		textLength, _, _ := procGetWindowTextLengthW.Call(window)
		if textLength == 0 {
			return 1
		}

		var pid uint32
		procGetWindowThreadProcessId.Call(window, uintptr(unsafe.Pointer(&pid)))
		if pid == 0 || pid == currentPID {
			return 1
		}

		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			return 1
		}
		defer windows.CloseHandle(handle)

		imageBuffer := make([]uint16, 32768)
		imageLength := uint32(len(imageBuffer))
		if err := windows.QueryFullProcessImageName(handle, 0, &imageBuffer[0], &imageLength); err != nil {
			return 1
		}
		executable := filepath.Base(windows.UTF16ToString(imageBuffer[:imageLength]))
		if executable == "" {
			return 1
		}

		titleBuffer := make([]uint16, int(textLength)+1)
		copied, _, _ := procGetWindowTextW.Call(
			window,
			uintptr(unsafe.Pointer(&titleBuffer[0])),
			uintptr(len(titleBuffer)),
		)
		windowTitle := ""
		if copied > 0 {
			windowTitle = strings.TrimSpace(windows.UTF16ToString(titleBuffer[:copied]))
		}

		if existing, found := windowApplications[pid]; !found || len(windowTitle) < len(existing.WindowTitle) {
			windowApplications[pid] = ApplicationProcess{
				PID:         pid,
				Executable:  executable,
				WindowTitle: windowTitle,
			}
		}
		return 1
	})
	procEnumWindows.Call(callback, 0)

	applications := make(map[uint32]ApplicationProcess)
	for _, application := range windowApplications {
		if rootPID, ok := resolveApplicationProcessID(application.Executable, application.PID); ok {
			application.PID = rootPID
		}
		if existing, found := applications[application.PID]; !found || len(application.WindowTitle) < len(existing.WindowTitle) {
			applications[application.PID] = application
		}
	}

	result := make([]ApplicationProcess, 0, len(applications))
	for _, application := range applications {
		result = append(result, application)
	}
	sort.Slice(result, func(i, j int) bool {
		left := strings.ToLower(result[i].Executable)
		right := strings.ToLower(result[j].Executable)
		if left != right {
			return left < right
		}
		if result[i].WindowTitle != result[j].WindowTitle {
			return result[i].WindowTitle < result[j].WindowTitle
		}
		return result[i].PID < result[j].PID
	})
	return result
}
