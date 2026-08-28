//go:build windows

package Utilities

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	comMultithreaded     = 0
	clsctxInprocServer   = 1
	eRender              = 0
	deviceStateActive    = 1
	applicationMeterTick = 50 * time.Millisecond
)

var (
	ole32Meter                = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeExMeter   = ole32Meter.NewProc("CoInitializeEx")
	procCoUninitializeMeter   = ole32Meter.NewProc("CoUninitialize")
	procCoCreateInstanceMeter = ole32Meter.NewProc("CoCreateInstance")
	clsidMMDeviceEnumerator   = mustWindowsGUID("BCDE0395-E52F-467C-8E3D-C4579291692E")
	iidIMMDeviceEnumerator    = mustWindowsGUID("A95664D2-9614-4F35-A746-DE8DB63617E6")
	iidIAudioSessionManager2  = mustWindowsGUID("77AA99A0-1BD6-484F-8BC7-2C654C9A9B6F")
	iidIAudioSessionControl2  = mustWindowsGUID("BFB7FF88-7239-4FC9-8FA2-07C950BE9C6D")
	iidIAudioMeterInformation = mustWindowsGUID("C02216F6-8C67-4B5B-9D00-D008E73E0064")
)

func mustWindowsGUID(value string) windows.GUID {
	if !strings.HasPrefix(value, "{") {
		value = "{" + value + "}"
	}
	guid, err := windows.GUIDFromString(value)
	if err != nil {
		panic(err)
	}
	return guid
}

func failedHRESULT(result uintptr) bool {
	return int32(uint32(result)) < 0
}

func hresultError(operation string, result uintptr) error {
	return fmt.Errorf("%s failed with HRESULT 0x%08X", operation, uint32(result))
}

func comCall(object unsafe.Pointer, methodIndex uintptr, arguments ...uintptr) uintptr {
	vtable := *(*unsafe.Pointer)(object)
	method := (*[32]uintptr)(vtable)[methodIndex]
	returnArguments := make([]uintptr, 0, len(arguments)+1)
	returnArguments = append(returnArguments, uintptr(object))
	returnArguments = append(returnArguments, arguments...)
	result, _, _ := syscall.SyscallN(method, returnArguments...)
	return result
}

func releaseCOM(object unsafe.Pointer) {
	if object != nil {
		comCall(object, 2)
	}
}

func queryCOMInterface(object unsafe.Pointer, iid *windows.GUID) (unsafe.Pointer, bool) {
	var result unsafe.Pointer
	hr := comCall(
		object,
		0,
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&result)),
	)
	return result, !failedHRESULT(hr) && result != nil
}

// StartApplicationAudioMeter monitors every active render endpoint and reports
// the loudest audio session belonging to the selected process or any child.
func StartApplicationAudioMeter(processID uint32, executable string, onPeak func(float32)) (*ApplicationAudioMeter, error) {
	if strings.TrimSpace(executable) == "" {
		return nil, fmt.Errorf("an application executable must be selected")
	}
	if onPeak == nil {
		return nil, fmt.Errorf("an application audio meter callback is required")
	}

	meter := newApplicationAudioMeter()
	ready := make(chan error, 1)
	go runApplicationAudioMeter(meter, processID, executable, onPeak, ready)
	if err := <-ready; err != nil {
		meter.Stop()
		return nil, err
	}
	return meter, nil
}

func runApplicationAudioMeter(meter *ApplicationAudioMeter, processID uint32, executable string, onPeak func(float32), ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(meter.doneCh)

	hr, _, _ := procCoInitializeExMeter.Call(0, comMultithreaded)
	if failedHRESULT(hr) {
		ready <- hresultError("CoInitializeEx", hr)
		return
	}
	defer procCoUninitializeMeter.Call()

	var deviceEnumerator unsafe.Pointer
	hr, _, _ = procCoCreateInstanceMeter.Call(
		uintptr(unsafe.Pointer(&clsidMMDeviceEnumerator)),
		0,
		clsctxInprocServer,
		uintptr(unsafe.Pointer(&iidIMMDeviceEnumerator)),
		uintptr(unsafe.Pointer(&deviceEnumerator)),
	)
	if failedHRESULT(hr) || deviceEnumerator == nil {
		ready <- hresultError("CoCreateInstance(MMDeviceEnumerator)", hr)
		return
	}
	defer releaseCOM(deviceEnumerator)
	ready <- nil

	ticker := time.NewTicker(applicationMeterTick)
	defer ticker.Stop()
	resolvedPID := processID
	lastResolution := time.Time{}
	processIDs := map[uint32]struct{}{processID: {}}

	for {
		select {
		case <-meter.stopCh:
			return
		case now := <-ticker.C:
			if now.Sub(lastResolution) >= 500*time.Millisecond {
				if currentPID, ok := resolveApplicationProcessID(executable, resolvedPID); ok {
					resolvedPID = currentPID
					processIDs = applicationProcessTreeIDs(currentPID)
				} else {
					processIDs = nil
				}
				lastResolution = now
			}

			peak := float32(0)
			if len(processIDs) > 0 {
				peak = applicationSessionPeak(deviceEnumerator, processIDs)
			}
			onPeak(peak)
		}
	}
}

func applicationSessionPeak(deviceEnumerator unsafe.Pointer, processIDs map[uint32]struct{}) float32 {
	var devices unsafe.Pointer
	hr := comCall(
		deviceEnumerator,
		3,
		eRender,
		deviceStateActive,
		uintptr(unsafe.Pointer(&devices)),
	)
	if failedHRESULT(hr) || devices == nil {
		return 0
	}
	defer releaseCOM(devices)

	var deviceCount uint32
	if failedHRESULT(comCall(devices, 3, uintptr(unsafe.Pointer(&deviceCount)))) {
		return 0
	}
	peak := float32(0)
	for index := uint32(0); index < deviceCount; index++ {
		var device unsafe.Pointer
		if failedHRESULT(comCall(devices, 4, uintptr(index), uintptr(unsafe.Pointer(&device)))) || device == nil {
			continue
		}
		devicePeak := applicationEndpointSessionPeak(device, processIDs)
		releaseCOM(device)
		if devicePeak > peak {
			peak = devicePeak
		}
	}
	return peak
}

func applicationEndpointSessionPeak(device unsafe.Pointer, processIDs map[uint32]struct{}) float32 {
	var manager unsafe.Pointer
	hr := comCall(
		device,
		3,
		uintptr(unsafe.Pointer(&iidIAudioSessionManager2)),
		clsctxInprocServer,
		0,
		uintptr(unsafe.Pointer(&manager)),
	)
	if failedHRESULT(hr) || manager == nil {
		return 0
	}
	defer releaseCOM(manager)

	var sessions unsafe.Pointer
	if failedHRESULT(comCall(manager, 5, uintptr(unsafe.Pointer(&sessions)))) || sessions == nil {
		return 0
	}
	defer releaseCOM(sessions)

	var sessionCount int32
	if failedHRESULT(comCall(sessions, 3, uintptr(unsafe.Pointer(&sessionCount)))) {
		return 0
	}
	peak := float32(0)
	for index := int32(0); index < sessionCount; index++ {
		var session unsafe.Pointer
		if failedHRESULT(comCall(sessions, 4, uintptr(index), uintptr(unsafe.Pointer(&session)))) || session == nil {
			continue
		}

		control, ok := queryCOMInterface(session, &iidIAudioSessionControl2)
		if !ok {
			releaseCOM(session)
			continue
		}
		var sessionPID uint32
		processResult := comCall(control, 14, uintptr(unsafe.Pointer(&sessionPID)))
		releaseCOM(control)
		if failedHRESULT(processResult) {
			releaseCOM(session)
			continue
		}
		if _, selected := processIDs[sessionPID]; !selected {
			releaseCOM(session)
			continue
		}

		meter, ok := queryCOMInterface(session, &iidIAudioMeterInformation)
		releaseCOM(session)
		if !ok {
			continue
		}
		var sessionPeak float32
		peakResult := comCall(meter, 3, uintptr(unsafe.Pointer(&sessionPeak)))
		releaseCOM(meter)
		if !failedHRESULT(peakResult) && sessionPeak > peak {
			peak = sessionPeak
		}
	}
	return peak
}
