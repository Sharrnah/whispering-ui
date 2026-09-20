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

func killBackendProcessGroup(pid int) {
	// Only called for the captured backend started with Setpgid. Its children
	// may still hold output pipes open after the parent has exited.
	if pid > 0 {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
}

func setNewProcessGroup(cmd *exec.Cmd) {
	// Start the Python process in its own process group so we can kill the group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
