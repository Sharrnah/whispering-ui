//go:build !windows

package RuntimeBackend

import (
	"os/exec"
	"syscall"
)

func (c *WhisperProcessConfig) assignProcessToJobObject(_ int) error {
	// Non-Windows: nothing to do.
	return nil
}

func closeJobObject(_ *WhisperProcessConfig) {}

func setNewProcessGroup(cmd *exec.Cmd) {
	// Start the Python process in its own process group so we can kill the group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
