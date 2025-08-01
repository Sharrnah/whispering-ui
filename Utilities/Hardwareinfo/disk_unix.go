//go:build !windows
// +build !windows

package Hardwareinfo

import (
	"syscall"
)

// getFreeSpace returns available disk space in bytes for the given path (non-Windows).
func GetFreeSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	// Available blocks * block size
	return stat.Bavail * uint64(stat.Bsize), nil
}
