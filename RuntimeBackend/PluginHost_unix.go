//go:build !windows

package RuntimeBackend

import (
	"os/exec"
	"syscall"
)

func killPluginProcessTree(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	_ = cmd.Process.Kill()
}
