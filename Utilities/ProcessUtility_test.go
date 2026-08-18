package Utilities

import (
	"os"
	"testing"
	"time"
)

func TestIsProcessRunningChecksTheProcessRatherThanOnlyThePort(t *testing.T) {
	if !isProcessRunning(os.Getpid()) {
		t.Fatal("current test process reported as stopped")
	}
	if isProcessRunning(0) {
		t.Fatal("PID 0 reported as a running backend process")
	}
}

func TestWaitForProcessTerminationDoesNotTrustAClosedPort(t *testing.T) {
	if waitForProcessTermination("127.0.0.1:0", os.Getpid(), 250*time.Millisecond) {
		t.Fatal("closed port was treated as termination while the process was still running")
	}
}
