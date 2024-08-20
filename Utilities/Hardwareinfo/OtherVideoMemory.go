package Hardwareinfo

import (
	"fmt"
	"github.com/jaypipes/ghw"
	"strings"
)

//
//func GetWinGPUMemory(gpuVendor string) (memoryUsed int64, memoryTotal int64) {
//	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM,AdapterCompatibility", "/Format:Textvaluelist")
//
//	// Hide command line window
//	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
//
//	output, err := cmd.Output()
//	if err != nil {
//		fmt.Printf("Error running wmic: %v\n", err)
//		return 0, 0
//	}
//
//	// output []byte to string
//	outputString := string(output[:])
//	outputString = strings.TrimSpace(outputString)
//
//	fmt.Println("GPU output")
//	fmt.Println(outputString)
//
//	// Split the output into records based on the field names
//	records := strings.Split(outputString, "AdapterCompatibility=")
//
//	for _, record := range records {
//		if record == "" {
//			continue
//		}
//
//		// Split the record into fields
//		fields := strings.Split(record, "AdapterRAM=")
//		if len(fields) < 2 {
//			continue
//		}
//		adapterCompatibility := strings.TrimSpace(fields[0])
//
//		ramAndName := strings.Split(fields[1], "Name=")
//		if len(ramAndName) < 2 {
//			continue
//		}
//		adapterRAM := strings.TrimSpace(ramAndName[0])
//		name := strings.TrimSpace(ramAndName[1])
//
//		if adapterRAM == "" {
//			fmt.Println("AdapterRAM field is empty")
//			continue
//		}
//
//		adapterRAMInt, err := strconv.ParseInt(adapterRAM, 10, 64)
//		if err != nil {
//			fmt.Println("Error converting AdapterRAM to integer:", err)
//			continue
//		}
//		//adapterRAMInt = adapterRAMInt // to megabytes
//
//		if strings.Contains(strings.ToLower(adapterCompatibility), strings.TrimSpace(strings.ToLower(gpuVendor))) {
//			return 0, adapterRAMInt
//		} else if gpuVendor == "" {
//			// Try to find AMD first, if not available, find nvidia, lastly try to find "Intel Corporation"
//			if strings.Contains(strings.ToLower(adapterCompatibility), "amd") {
//				return 0, adapterRAMInt
//			}
//			if strings.Contains(strings.ToLower(adapterCompatibility), "nvidia") {
//				return 0, adapterRAMInt
//			}
//			if strings.Contains(strings.ToLower(adapterCompatibility), "intel") {
//				return 0, adapterRAMInt
//			}
//		}
//
//		// Print the extracted information
//		fmt.Printf("AdapterCompatibility: %s, AdapterRAM: %s, Name: %s\n", adapterCompatibility, adapterRAM, name)
//	}
//	return 0, 0
//}

func GetWinGPUMemory(gpuVendor string) (memoryUsed int64, memoryTotal int64) {
	gpu, err := ghw.GPU()
	if err != nil {
		fmt.Printf("Error getting GPU info: %v", err)
		return 0, 0
	}

	fmt.Printf("GPU: %v\n", gpu)
	if gpu != nil {
		for _, card := range gpu.GraphicsCards {
			fmt.Printf(" %v\n", card)
			if card.Node == nil {
				continue
			}
			if strings.Contains(strings.ToLower(card.DeviceInfo.Vendor.Name), strings.TrimSpace(strings.ToLower(gpuVendor))) {
				return 0, card.Node.Memory.TotalUsableBytes / 1024 / 1024
			} else if gpuVendor == "" {
				if strings.Contains(strings.ToLower(card.DeviceInfo.Vendor.Name), strings.ToLower("NVIDIA")) {
					fmt.Printf("NVIDIA Card found.\n")
					return 0, card.Node.Memory.TotalUsableBytes / 1024 / 1024
				}
				if strings.Contains(strings.ToLower(card.DeviceInfo.Vendor.Name), strings.ToLower("AMD")) {
					fmt.Printf("AMD Card found.\n")
					return 0, card.Node.Memory.TotalUsableBytes / 1024 / 1024
				}
				if strings.Contains(strings.ToLower(card.DeviceInfo.Vendor.Name), strings.ToLower("Intel")) {
					fmt.Printf("Intel Card found.\n")
					return 0, card.Node.Memory.TotalUsableBytes / 1024 / 1024
				}
			}
		}
	}
	return 0, 0
}
