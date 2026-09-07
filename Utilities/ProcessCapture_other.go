//go:build !windows

package Utilities

import "fmt"

func StartApplicationAudioCapture(executable string, pid uint32, emit func([]byte)) (func(), error) {
	return nil, fmt.Errorf("per-application capture is not implemented on this platform; select an audio monitor input")
}
