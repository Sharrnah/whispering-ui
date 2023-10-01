package Hardwareinfo

import (
	"fmt"
	"github.com/jaypipes/ghw"
)

type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func GetCPUMemory() int64 {
	memory, err := ghw.Memory()
	if err != nil {
		fmt.Printf("Error getting memory info: %v", err)
		return 0
	}

	if memory != nil {
		fmt.Printf("Memory: %v\n", memory)
		return memory.TotalUsableBytes / 1024 / 1024
	}
	return 0
}
