//go:build !windows

package Utilities

// GetApplicationProcesses intentionally returns no entries outside Windows.
// The profile and regular audio device code remain fully portable.
func GetApplicationProcesses() []ApplicationProcess {
	return nil
}
