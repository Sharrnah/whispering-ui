//go:build windows

package Utilities

import (
	"os/exec"
	"syscall"
)

func ProcessHideWindowAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
