//go:build windows

package RuntimeBackend

import "os/exec"

func killPluginProcessTree(cmd *exec.Cmd) { _ = cmd.Process.Kill() }
