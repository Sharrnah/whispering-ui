package RuntimeBackend

import (
	"os"
	"os/exec"
	"testing"
)

func TestStopProcessUsesCapturedProcessAfterProgramIsCleared(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start short-lived process: %v", err)
	}
	process := cmd.Process
	if err := cmd.Wait(); err != nil {
		t.Fatalf("wait for short-lived process: %v", err)
	}

	config := &WhisperProcessConfig{}
	config.stopProcess(cmd, process, 0)

	if config.Program != nil || config.running {
		t.Fatal("stopped config unexpectedly became active")
	}
}

func TestClearProcessIfCurrentPreservesReplacement(t *testing.T) {
	oldProcess := &exec.Cmd{}
	replacement := &exec.Cmd{}
	config := &WhisperProcessConfig{
		Program: replacement,
		running: true,
	}

	config.clearProcessIfCurrent(oldProcess)

	if config.Program != replacement || !config.running {
		t.Fatal("cleanup from an old process cleared its replacement")
	}
}
