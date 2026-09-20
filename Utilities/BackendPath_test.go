package Utilities

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBackendInstalledUsesNativeExecutable(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "audioWhisper"), 0755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "linux" {
		if err := os.WriteFile(filepath.Join(root, "audioWhisper", "audioWhisper.exe"), []byte("windows"), 0755); err != nil {
			t.Fatal(err)
		}
		if BackendInstalled(root) {
			t.Fatal("Windows executable counted as Linux backend")
		}
	}
	if err := os.WriteFile(BackendExecutable(root), []byte("native"), 0755); err != nil {
		t.Fatal(err)
	}
	if !BackendInstalled(root) {
		t.Fatal("native backend not detected")
	}
}
