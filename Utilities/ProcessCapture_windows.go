//go:build windows

package Utilities

// Native equivalent of the backend's WASAPI process-loopback adapter. All COM
// calls are confined to one MTA thread; packets are mono PCM16 at 16 kHz.
import (
	"fmt"
	"golang.org/x/sys/windows"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

var activateProcessAudio = windows.NewLazySystemDLL("mmdevapi.dll").NewProc("ActivateAudioInterfaceAsync")
var processAudioIID = mustWindowsGUID("1CB9AD4C-DBFA-4C32-B178-C2F568A703B2")
var processCaptureIID = mustWindowsGUID("C8ADBD64-E71E-48A0-A4DE-185C395CD317")
var processCompletionIID = mustWindowsGUID("41D949AB-9862-444A-80F6-C261334DA5EB")
var processAgileIID = mustWindowsGUID("94EA2B94-E9CC-49E0-C0FF-EE64CA8F5B90")
var processUnknownIID = mustWindowsGUID("00000000-0000-0000-C000-000000000046")
var processComOwners sync.Map

type processActivationResult struct {
	client unsafe.Pointer
	err    error
}
type processCompletion struct {
	vtable *[4]uintptr
	refs   int32
	done   chan processActivationResult
}

var processCompletionVTable = [4]uintptr{
	windows.NewCallback(func(this, iid, out uintptr) uintptr {
		if out == 0 {
			return 0x80004002
		}
		wanted := *(*windows.GUID)(unsafe.Pointer(iid))
		if wanted != processCompletionIID && wanted != processAgileIID && wanted != processUnknownIID {
			*(*uintptr)(unsafe.Pointer(out)) = 0
			return 0x80004002
		}
		*(*uintptr)(unsafe.Pointer(out)) = this
		atomic.AddInt32(&(*processCompletion)(unsafe.Pointer(this)).refs, 1)
		return 0
	}),
	windows.NewCallback(func(this uintptr) uintptr {
		return uintptr(atomic.AddInt32(&(*processCompletion)(unsafe.Pointer(this)).refs, 1))
	}),
	windows.NewCallback(func(this uintptr) uintptr {
		owner := (*processCompletion)(unsafe.Pointer(this))
		refs := atomic.AddInt32(&owner.refs, -1)
		if refs == 0 {
			processComOwners.Delete(owner)
		}
		return uintptr(refs)
	}),
	windows.NewCallback(func(this, operation uintptr) uintptr {
		owner := (*processCompletion)(unsafe.Pointer(this))
		var hr int32
		var client unsafe.Pointer
		result := comCall(unsafe.Pointer(operation), 3, uintptr(unsafe.Pointer(&hr)), uintptr(unsafe.Pointer(&client)))
		var err error
		if failedHRESULT(result) {
			err = hresultError("GetActivateResult", result)
		} else if hr < 0 {
			err = hresultError("Process-loopback activation", uintptr(uint32(hr)))
		}
		owner.done <- processActivationResult{client, err}
		return 0
	}),
}

type processNativeCapture struct {
	client, capture, operation unsafe.Pointer
	event                      windows.Handle
}

func (c *processNativeCapture) close() {
	if c.client != nil {
		comCall(c.client, 11)
	}
	releaseCOM(c.capture)
	releaseCOM(c.client)
	releaseCOM(c.operation)
	if c.event != 0 {
		windows.CloseHandle(c.event)
	}
}

func openProcessCapture(pid uint32) (*processNativeCapture, error) {
	if err := activateProcessAudio.Find(); err != nil {
		return nil, err
	}
	activation := struct{ kind, pid, mode uint32 }{1, pid, 0}
	variant := struct {
		vt       uint16
		reserved [3]uint16
		size     uint32
		data     uintptr
	}{vt: 65, size: 12, data: uintptr(unsafe.Pointer(&activation))}
	owner := &processCompletion{vtable: &processCompletionVTable, refs: 1, done: make(chan processActivationResult, 1)}
	processComOwners.Store(owner, true)
	defer func() {
		if atomic.AddInt32(&owner.refs, -1) == 0 {
			processComOwners.Delete(owner)
		}
	}()
	capture := &processNativeCapture{}
	name, _ := windows.UTF16PtrFromString("VAD\\Process_Loopback")
	hr, _, _ := activateProcessAudio.Call(uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&processAudioIID)), uintptr(unsafe.Pointer(&variant)), uintptr(unsafe.Pointer(owner)), uintptr(unsafe.Pointer(&capture.operation)))
	runtime.KeepAlive(activation)
	runtime.KeepAlive(variant)
	if failedHRESULT(hr) {
		capture.close()
		return nil, hresultError("ActivateAudioInterfaceAsync", hr)
	}
	select {
	case result := <-owner.done:
		capture.client = result.client
		if result.err != nil {
			capture.close()
			return nil, result.err
		}
	case <-time.After(10 * time.Second):
		// Keep the callback alive until Windows finishes activation, even on timeout.
		go func() { result := <-owner.done; releaseCOM(result.client); releaseCOM(capture.operation) }()
		return nil, fmt.Errorf("process audio activation timed out")
	}
	if capture.client == nil {
		capture.close()
		return nil, fmt.Errorf("process audio returned no client")
	}
	format := struct {
		kind, channels         uint16
		rate, bytesPerSecond   uint32
		alignment, bits, extra uint16
	}{1, 1, 16000, 32000, 2, 16, 0}
	hr = comCall(capture.client, 3, 0, 0x88060000, 0, 0, uintptr(unsafe.Pointer(&format)), 0)
	if failedHRESULT(hr) {
		capture.close()
		return nil, hresultError("Process audio Initialize", hr)
	}
	event, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		capture.close()
		return nil, err
	}
	capture.event = event
	hr = comCall(capture.client, 13, uintptr(event))
	if !failedHRESULT(hr) {
		hr = comCall(capture.client, 14, uintptr(unsafe.Pointer(&processCaptureIID)), uintptr(unsafe.Pointer(&capture.capture)))
	}
	if !failedHRESULT(hr) {
		hr = comCall(capture.client, 10)
	}
	if failedHRESULT(hr) {
		capture.close()
		return nil, hresultError("Process audio Start", hr)
	}
	return capture, nil
}

func (c *processNativeCapture) read(emit func([]byte)) (bool, error) {
	received := false
	for {
		var frames uint32
		hr := comCall(c.capture, 5, uintptr(unsafe.Pointer(&frames)))
		if failedHRESULT(hr) {
			return received, hresultError("Process audio packet size", hr)
		}
		if frames == 0 {
			return received, nil
		}
		var data unsafe.Pointer
		var flags uint32
		hr = comCall(c.capture, 3, uintptr(unsafe.Pointer(&data)), uintptr(unsafe.Pointer(&frames)), uintptr(unsafe.Pointer(&flags)), 0, 0)
		if failedHRESULT(hr) {
			return received, hresultError("Process audio GetBuffer", hr)
		}
		packet := make([]byte, int(frames)*2)
		if flags&2 == 0 && data != nil {
			copy(packet, unsafe.Slice((*byte)(data), len(packet)))
		}
		hr = comCall(c.capture, 4, uintptr(frames))
		if failedHRESULT(hr) {
			return received, hresultError("Process audio ReleaseBuffer", hr)
		}
		emit(packet)
		received = true
	}
}

// StartApplicationAudioCapture tracks a saved executable/PID across restarts.
// An absent application emits silence instead of switching to system audio.
func StartApplicationAudioCapture(executable string, pid uint32, emit func([]byte)) (func(), error) {
	if windows.RtlGetVersion().BuildNumber < 20348 {
		return nil, fmt.Errorf("application audio capture requires Windows build 20348 or newer")
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	ready := make(chan error, 1)
	var once sync.Once
	go func() {
		defer close(done)
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		hr, _, _ := procCoInitializeExMeter.Call(0, 0)
		if failedHRESULT(hr) {
			ready <- hresultError("CoInitializeEx", hr)
			return
		}
		defer procCoUninitializeMeter.Call()
		var capture *processNativeCapture
		defer func() {
			if capture != nil {
				capture.close()
			}
		}()
		resolved, found := resolveApplicationProcessID(executable, pid)
		if found {
			var err error
			capture, err = openProcessCapture(resolved)
			if err != nil {
				ready <- err
				return
			}
		}
		ready <- nil
		lastAudio, lastEmit, lastResolve := time.Now(), time.Now(), time.Now()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case now := <-ticker.C:
				if now.Sub(lastResolve) >= time.Second {
					next, exists := resolveApplicationProcessID(executable, resolved)
					lastResolve = now
					if capture != nil && (!exists || next != resolved) {
						capture.close()
						capture = nil
					}
					if capture == nil && exists {
						capture, _ = openProcessCapture(next)
						resolved = next
					}
				}
				if capture != nil {
					received, err := capture.read(func(pcm []byte) { emit(pcm); lastEmit = time.Now() })
					if err != nil {
						capture.close()
						capture = nil
					} else if received {
						lastAudio = now
					}
				}
				if now.Sub(lastAudio) >= 64*time.Millisecond && now.Sub(lastEmit) >= 32*time.Millisecond {
					emit(make([]byte, 1024))
					lastEmit = now
				}
			}
		}
	}()
	err := <-ready
	if err != nil {
		<-done
		return nil, err
	}
	return func() { once.Do(func() { close(stop) }); <-done }, nil
}
