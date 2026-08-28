package Utilities

import (
	"strconv"
	"strings"
)

const (
	audioProcessOptionPrefix    = "wasapi-process:"
	AudioApplicationOptionValue = "wasapi-application-audio"
)

// ApplicationProcess is a visible top-level application that can be targeted
// by Windows WASAPI process-loopback capture.
type ApplicationProcess struct {
	PID         uint32
	Executable  string
	WindowTitle string
}

func FormatAudioProcessOptionValue(pid uint32, executable string) string {
	return audioProcessOptionPrefix + strconv.FormatUint(uint64(pid), 10) + "|" + executable
}

func ParseAudioProcessOptionValue(value string) (uint32, string, bool) {
	if !strings.HasPrefix(value, audioProcessOptionPrefix) {
		return 0, "", false
	}
	parts := strings.SplitN(strings.TrimPrefix(value, audioProcessOptionPrefix), "|", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return 0, "", false
	}
	pid, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, "", false
	}
	return uint32(pid), parts[1], true
}

func IsAudioApplicationOptionValue(value string) bool {
	return value == AudioApplicationOptionValue
}
