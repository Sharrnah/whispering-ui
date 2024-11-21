package Hardwareinfo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"whispering-tiger-ui/Utilities"
)

type GPUInfo struct {
	AdapterName string
	VendorName  string
	MemoryMB    int64
}

var cachedGPUs []GPUInfo
var cacheInitialized bool

func GetWinGPUs() ([]GPUInfo, error) {
	if cacheInitialized {
		return cachedGPUs, nil
	}

	cmd := exec.Command("powershell", "-Command",
		"Get-ItemProperty -Path 'HKLM:\\SYSTEM\\ControlSet001\\Control\\Class\\{4d36e968-e325-11ce-bfc1-08002be10318}\\0*' | "+
			"ForEach-Object { "+
			"$adapterString = $_.'HardwareInformation.AdapterString'; "+
			"$providerName = $_.ProviderName; "+
			"$memorySize = $_.'HardwareInformation.qwMemorySize'; "+
			"\"$adapterString`n$providerName`n$memorySize\" "+
			"}",
	)
	Utilities.ProcessHideWindowAttr(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error reading from registry: %v", err)
	}

	outputString := string(output)
	lines := strings.Split(strings.TrimSpace(outputString), "\n")
	var gpus []GPUInfo

	for i := 0; i < len(lines); i += 3 {
		if i+2 < len(lines) {
			adapterName := strings.TrimSpace(lines[i])
			vendorName := strings.TrimSpace(lines[i+1])
			memorySizeStr := strings.TrimSpace(lines[i+2])
			memorySize, err := strconv.ParseInt(memorySizeStr, 10, 64)
			if err != nil {
				fmt.Printf("Error parsing memory size: %v\n", err)
				memorySize = 0
			}
			if memorySize > 0 {
				// Convert bytes to megabytes
				memorySize = memorySize / Utilities.MiB
			}
			gpus = append(gpus, GPUInfo{
				AdapterName: adapterName,
				VendorName:  vendorName,
				MemoryMB:    memorySize,
			})
		}
	}

	return gpus, nil
}

func GetGPUByVendor(partialVendorName string) ([]GPUInfo, error) {
	gpus, err := GetWinGPUs()
	if err != nil {
		return nil, err
	}

	var matchedGPUs []GPUInfo
	partialVendorName = strings.TrimSpace(strings.ToLower(partialVendorName))

	if partialVendorName == "" {
		return gpus, nil
	}

	for _, gpu := range gpus {
		if strings.Contains(strings.ToLower(gpu.VendorName), partialVendorName) {
			matchedGPUs = append(matchedGPUs, gpu)
		}
	}

	return matchedGPUs, nil
}

func FindDedicatedGPUByVendor(partialVendorNames []string) ([]GPUInfo, error) {
	for _, vendor := range partialVendorNames {
		gpus, err := GetGPUByVendor(vendor)
		if err == nil && len(gpus) > 0 {
			return gpus, nil
		}
	}

	// If no GPUs were found for any of the specified vendors, return all available GPUs
	gpus, err := GetGPUByVendor("")
	if err == nil && len(gpus) > 0 {
		return gpus, nil
	}

	return nil, fmt.Errorf("no dedicated GPUs found for vendors: %s", strings.Join(partialVendorNames, ", "))
}
