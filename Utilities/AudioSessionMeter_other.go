//go:build !windows

package Utilities

import "errors"

func StartApplicationAudioMeter(_ uint32, _ string, _ func(float32)) (*ApplicationAudioMeter, error) {
	return nil, errors.New("application audio metering is only available on Windows")
}
