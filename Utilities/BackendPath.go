package Utilities

import (
	"os"
	"path/filepath"
	"runtime"
)

func BackendExecutable(root string) string {
	name := "audioWhisper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(root, "audioWhisper", name)
}

func BackendInstalled(root string) bool {
	for _, path := range []string{BackendExecutable(root), filepath.Join(root, "audioWhisper.py")} {
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	return false
}
