package Hardwareinfo

import (
	"encoding/xml"
	"fmt"
	"github.com/jaypipes/ghw"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type nvidiaSmiLog struct {
	GPUs []gpu `xml:"gpu"`
}

type gpu struct {
	MemoryUsed  string `xml:"fb_memory_usage>used"`
	MemoryTotal string `xml:"fb_memory_usage>total"`
}

func HasNVIDIACard() bool {
	gpu, err := ghw.GPU()
	if err != nil {
		fmt.Printf("Error getting GPU info: %v", err)
		return false
	}

	fmt.Printf("GPU: %v\n", gpu)
	if gpu != nil {
		for _, card := range gpu.GraphicsCards {
			fmt.Printf(" %v\n", card)
			if strings.ToLower(card.DeviceInfo.Vendor.Name) == strings.ToLower("NVIDIA") {
				fmt.Printf("NVIDIA Card found.\n")
				return true
			}
		}
	}
	return false
}

func haveExe(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func GetGPUMemory() (memoryUsed int64, memoryTotal int64) {
	if haveExe("nvidia-smi") {
		cmd := exec.Command("nvidia-smi", "-q", "-x")

		// Hide command line window
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error running nvidia-smi: %v\n", err)
			return 0, 0
		}

		var nlog nvidiaSmiLog
		if err := xml.Unmarshal(output, &nlog); err != nil {
			fmt.Printf("Error parsing nvidia-smi output: %v\n", err)
			return 0, 0
		}

		for i, gpu := range nlog.GPUs {
			memoryUsed := strings.TrimSpace(gpu.MemoryUsed)
			memoryTotal := strings.TrimSpace(gpu.MemoryTotal)
			fmt.Printf("GPU %d: Memory used: %s, Memory total: %s\n", i, memoryUsed, memoryTotal)

			if strings.HasSuffix(memoryUsed, "MiB") {
				memoryUsed = strings.TrimSpace(memoryUsed[:len(memoryUsed)-3])
			}
			if strings.HasSuffix(memoryTotal, "MiB") {
				memoryTotal = strings.TrimSpace(memoryTotal[:len(memoryTotal)-3])
			}

			// convert memoryUsed and memoryTotal to int64
			memoryUsedInt, _ := strconv.ParseInt(memoryUsed, 10, 64)
			memoryTotalInt, _ := strconv.ParseInt(memoryTotal, 10, 64)

			return memoryUsedInt, memoryTotalInt
		}
	} else {
		fmt.Printf("nvidia-smi not found\n")
	}
	return 0, 0
}

func GetGPUComputeCapability() (computeCapabilityVersion float32) {
	if haveExe("nvidia-smi") {
		cmd := exec.Command("nvidia-smi", "--query-gpu=compute_cap", "--format=csv,noheader")

		// Hide command line window
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error running nvidia-smi: %v\n", err)
			return 0.0
		}

		// output []byte to string
		outputString := string(output[:])
		outputString = strings.TrimSpace(outputString)

		// convert outputString to float32
		computeCapabilityVersion, _ := strconv.ParseFloat(outputString, 32)
		return float32(computeCapabilityVersion)
	} else {
		fmt.Printf("nvidia-smi not found\n")
	}
	return 0.0
}
