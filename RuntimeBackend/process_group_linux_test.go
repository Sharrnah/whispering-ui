//go:build linux

package RuntimeBackend

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestForcedShutdownStopsBackendChildren(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60 & echo $!; wait")
	setNewProcessGroup(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { killBackendProcessGroup(cmd.Process.Pid); _ = cmd.Wait() })
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("backend child did not start")
	}
	child, err := strconv.Atoi(scanner.Text())
	if err != nil {
		t.Fatal(err)
	}
	killBackendProcessGroup(cmd.Process.Pid)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		stat, err := os.ReadFile("/proc/" + strconv.Itoa(child) + "/stat")
		if os.IsNotExist(err) {
			return
		}
		// A zombie has terminated; its reaping belongs to the container's init.
		if err == nil && strings.Contains(string(stat), ") Z ") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("child remained alive after backend group termination")
}
