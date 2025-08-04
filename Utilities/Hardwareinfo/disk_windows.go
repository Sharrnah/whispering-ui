//go:build windows
// +build windows

package Hardwareinfo

import (
	"golang.org/x/sys/windows"
)

func GetFreeSpace(path string) (uint64, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(p, &freeBytesAvailable, &totalBytes, &totalFreeBytes)
	if err != nil {
		return 0, err
	}
	return freeBytesAvailable, nil
}
