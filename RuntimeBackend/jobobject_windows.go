//go:build windows

package RuntimeBackend

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

func (c *WhisperProcessConfig) assignProcessToJobObject(pid int) error {
	// Create Job if we haven’t already.
	if c.jobObjectHandle == 0 {
		h, err := windows.CreateJobObject(nil, nil)
		if err != nil {
			return err
		}
		// Set KILL_ON_JOB_CLOSE so all children die when we close this handle.
		var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
		info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
		if _, err := windows.SetInformationJobObject(
			h,
			windows.JobObjectExtendedLimitInformation,
			uintptr(unsafe.Pointer(&info)),
			uint32(unsafe.Sizeof(info)),
		); err != nil {
			windows.CloseHandle(h)
			return err
		}
		c.jobObjectHandle = uintptr(h)
	}

	// Open a HANDLE for the process ID.
	ph, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(ph)

	// Assign the process to the Job.
	return windows.AssignProcessToJobObject(windows.Handle(c.jobObjectHandle), ph)
}

func closeJobObject(c *WhisperProcessConfig) {
	if c.jobObjectHandle != 0 {
		windows.CloseHandle(windows.Handle(c.jobObjectHandle))
		c.jobObjectHandle = 0
	}
}

// On Windows this is a no-op (we use Job Objects instead)
func setNewProcessGroup(_ *exec.Cmd) {}
