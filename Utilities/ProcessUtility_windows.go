//go:build windows

package Utilities

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func ProcessHideWindowAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

// isProcessRunning checks the process handle itself. On Windows,
// os.FindProcess succeeds even for a PID that no longer exists, so it cannot
// be used to confirm that a backend actually stopped.
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		// Access denied means the process exists but cannot be inspected. Treat
		// all errors except an invalid PID conservatively as still running.
		return err != windows.ERROR_INVALID_PARAMETER
	}
	defer windows.CloseHandle(handle)

	status, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return true
	}
	return status == uint32(windows.WAIT_TIMEOUT)
}
