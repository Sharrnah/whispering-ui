//go:build linux

package Utilities

import (
	"os/exec"
	"syscall"
)

func ProcessHideWindowAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
