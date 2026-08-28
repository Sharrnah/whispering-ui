package Pages

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/Advanced"
	"whispering-tiger-ui/Pages/ProfileSettings"
	PF "whispering-tiger-ui/ProfileForm"
	"whispering-tiger-ui/Profiles"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/UpdateUtility"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Utilities/AudioAPI"
	"whispering-tiger-ui/Utilities/Hardwareinfo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"github.com/getsentry/sentry-go"
	"github.com/youpy/go-wav"
)

type CurrentPlaybackDevice struct {
	InputDeviceName  string
	OutputDeviceName string
	ProcessLoopback  bool
	InputWaveWidget  *widget.ProgressBar
	OutputWaveWidget *widget.ProgressBar
	Context          *malgo.AllocatedContext
	AudioAPI         malgo.Backend

	device              *malgo.Device
	stopChannel         chan bool
	playTestAudio       bool
	testAudioChannels   uint32
	testAudioSampleRate uint32
	isInitializing      bool       // Add this flag
	initMutex           sync.Mutex // Add this mutex
	// Stop coordination to avoid losing stop signals before channel initialization
	stopPending bool
	stopMu      sync.Mutex

	// contextShutdown indicates an Init() goroutine has finished and context was freed.
	// We wait on this (via polling) before starting a new context when switching APIs
	// to avoid racing between freeing the underlying C context and creating devices.
	contextWG sync.WaitGroup

	applicationMeter           *Utilities.ApplicationAudioMeter
	applicationMeterGeneration uint64
	applicationMeterMu         sync.Mutex
}

// Context lifecycle overview:
//  - StartContext(): kicks off a new Init() goroutine run (Add(1) + go Init()).
//  - Init(): Creates the malgo.Context for the currently selected backend, waits for a stop signal and performs cleanup.
//            Always calls contextWG.Done() at the end.
//  - Stop(): Sends (non-blocking) a signal to the running Init() goroutine so its <-c.stopChannel returns.
//  - StopAndWaitContext(timeout): Combines Stop()+UnInitDevices() and waits (with timeout) for the Init goroutine to fully finish.
//  - SwitchBackend(): High-level backend change (stop, wait, swap backend, start new context, re-initialize devices).
//
// This keeps UI code free from WaitGroup/channel boilerplate; it only calls SwitchBackend or StartContext.

// StartContext starts (or restarts) the audio context initialization goroutine for the current backend.
// It is safe to call multiple times; a previous context should be stopped first via StopAndWaitContext.
func (c *CurrentPlaybackDevice) StartContext() {
	c.contextWG.Add(1)
	go c.Init()
}

func (c *CurrentPlaybackDevice) StopApplicationAudioMeter() {
	c.applicationMeterMu.Lock()
	c.applicationMeterGeneration++
	meter := c.applicationMeter
	c.applicationMeter = nil
	inputWaveWidget := c.InputWaveWidget
	c.applicationMeterMu.Unlock()
	if meter != nil {
		meter.Stop()
	}
	if inputWaveWidget != nil {
		fyne.Do(func() { inputWaveWidget.SetValue(0) })
	}
}

func (c *CurrentPlaybackDevice) StartApplicationAudioMeter(processID uint32, executable string) error {
	c.StopApplicationAudioMeter()
	c.applicationMeterMu.Lock()
	c.applicationMeterGeneration++
	generation := c.applicationMeterGeneration
	c.applicationMeterMu.Unlock()

	meter, err := Utilities.StartApplicationAudioMeter(processID, executable, func(peak float32) {
		level := normalizedAudioMeterLevel(float64(peak))
		fyne.Do(func() {
			c.applicationMeterMu.Lock()
			current := generation == c.applicationMeterGeneration
			inputWaveWidget := c.InputWaveWidget
			c.applicationMeterMu.Unlock()
			if current && inputWaveWidget != nil {
				inputWaveWidget.SetValue(level)
			}
		})
	})
	if err != nil {
		return err
	}

	c.applicationMeterMu.Lock()
	if generation != c.applicationMeterGeneration {
		c.applicationMeterMu.Unlock()
		meter.Stop()
		return nil
	}
	c.applicationMeter = meter
	c.applicationMeterMu.Unlock()
	return nil
}

// StopAndWaitContext signals the Init goroutine to stop (via Stop()) and waits until it has fully exited.
// Returns true if the context goroutine finished before timeout.
func (c *CurrentPlaybackDevice) StopAndWaitContext(timeout time.Duration) bool {
	// signal stop (idempotent)
	c.Stop()
	// also ensure device is uninitialized early to reduce callbacks during shutdown
	c.UnInitDevices()

	doneCh := make(chan struct{})
	go func() {
		c.contextWG.Wait() // will return immediately if no active goroutine (mismatched Add/Done would panic earlier)
		close(doneCh)
	}()
	select {
	case <-doneCh:
		return true
	case <-time.After(timeout):
		fmt.Println("Timeout waiting for audio context shutdown")
		return false
	}
}

// SwitchBackend encapsulates full backend change: stop old, wait, set backend, start new context, then re-init devices.
// inputName/outputName may be empty to keep current selection. Returns first encountered error.
func (c *CurrentPlaybackDevice) SwitchBackend(newBackend malgo.Backend, inputName, outputName string) error {
	// If backend unchanged, just optionally re-init devices.
	if c.AudioAPI == newBackend {
		if inputName != "" {
			c.InputDeviceName = inputName
		}
		if outputName != "" {
			c.OutputDeviceName = outputName
		}
		// attempt device reinit async
		go func() { _ = c.InitDevices(false) }()
		return nil
	}

	// Stop and wait for previous context (max 3s)
	c.StopAndWaitContext(3 * time.Second)

	// Assign new backend and start a fresh context
	c.AudioAPI = newBackend
	if inputName != "" {
		c.InputDeviceName = inputName
	}
	if outputName != "" {
		c.OutputDeviceName = outputName
	}
	c.StartContext()

	// After (at most) 5s wait for initialization, then init devices.
	go func() {
		c.WaitUntilInitialized(5)
		_ = c.InitDevices(false)
	}()
	return nil
}

// WaitForContextShutdown waits until c.Context is nil (meaning prior Init goroutine fully cleaned up)
// or the timeout (in seconds) expires.
func (c *CurrentPlaybackDevice) WaitForContextShutdown(timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout*time.Second { // keep existing style (seconds multiplication)
		if c.Context == nil { // previous context fully cleaned
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (c *CurrentPlaybackDevice) Stop() {
	// Non-blocking stop signal. If the channel does not yet exist,
	// mark stop as pending and deliver it later.
	c.stopMu.Lock()
	defer c.stopMu.Unlock()
	if c.stopChannel != nil {
		select { // try non-blocking
		case c.stopChannel <- true:
		default:
		}
	} else {
		c.stopPending = true
	}
}

func (c *CurrentPlaybackDevice) PlayStopTestAudio() {
	c.playTestAudio = !c.playTestAudio
}

func (c *CurrentPlaybackDevice) IsPlayingTestAudio() bool {
	return c.playTestAudio
}

func (c *CurrentPlaybackDevice) InitTestAudio() (*bytes.Reader, *wav.Reader) {
	byteReader := bytes.NewReader(Resources.ResourceTestWav.Content())
	testAudioReader := wav.NewReader(byteReader)
	testAudioFormat, err := testAudioReader.Format()
	if err != nil {
		fmt.Println(err)
		Logging.CaptureException(err)
		Logging.Flush(Logging.FlushTimeoutDefault)
		os.Exit(1)
	}
	c.testAudioChannels = uint32(testAudioFormat.NumChannels)
	c.testAudioSampleRate = testAudioFormat.SampleRate
	return byteReader, testAudioReader
}

func (c *CurrentPlaybackDevice) InitDevices(isPlayback bool) error {
	// If context is not yet present, we should not attempt to init devices.
	if c.Context == nil {
		return errors.New("cannot init devices: context not ready")
	}
	c.initMutex.Lock()
	defer c.initMutex.Unlock()
	if c.isInitializing {
		return nil // Prevent concurrent initialization
	}
	c.isInitializing = true
	defer func() { c.isInitializing = false }()

	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Profiles->InitDevices")
	})

	byteReader, testAudioReader := c.InitTestAudio()

	// Properly stop and cleanup existing device with longer wait time
	if c.device != nil {
		if c.device.IsStarted() {
			c.device.Stop()
			time.Sleep(200 * time.Millisecond) // Wait for device to fully stop
		}
		c.device.Uninit()
		c.device = nil
		time.Sleep(200 * time.Millisecond) // Increased wait time for WASAPI cleanup
	}

	// wait in a loop until c.Context is not nil before trying to initialize
	for {
		if c.Context != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if c.Context == nil {
		if c.Context == nil {
			c.Init()
			time.Sleep(200 * time.Millisecond)
		}
	}

	captureDevices, err := c.Context.Devices(malgo.Capture)
	if err != nil {
		fmt.Println(err)
		Logging.CaptureException(err)
		return err
	}

	isLoopback := false
	selectedCaptureDeviceIndex := -1
	for index, deviceInfo := range captureDevices {
		if deviceInfo.Name() == c.InputDeviceName {
			selectedCaptureDeviceIndex = index
			fmt.Println("Found input device: ", deviceInfo.Name(), " at index: ", selectedCaptureDeviceIndex)
			break
		}
	}

	if selectedCaptureDeviceIndex == -1 {
		captureLoopbackDevices, err := c.Context.Devices(malgo.Loopback)
		if err != nil {
			fmt.Println(err)
			Logging.CaptureException(err)
		}
		for index, deviceInfo := range captureLoopbackDevices {
			if deviceInfo.Name()+" [Loopback]" == c.InputDeviceName {
				selectedCaptureDeviceIndex = len(captureDevices) + index
				isLoopback = true
				fmt.Println("Found input loopback device: ", deviceInfo.Name(), " at index: ", selectedCaptureDeviceIndex)
				break
			}
		}
		captureDevices = append(captureDevices, captureLoopbackDevices...)
	}

	playbackDevices, err := c.Context.Devices(malgo.Playback)
	if err != nil {
		fmt.Println(err)
	}
	selectedPlaybackDeviceIndex := -1
	for index, deviceInfo := range playbackDevices {
		if deviceInfo.Name() == c.OutputDeviceName {
			selectedPlaybackDeviceIndex = index
			fmt.Println("Found output device: ", deviceInfo.Name(), " at index: ", index)
			break
		}
	}

	if c.ProcessLoopback {
		// The profile preview uses miniaudio, which only supports endpoint
		// loopback. Keep output testing available while the Python backend owns
		// the process-specific WASAPI capture stream.
		isPlayback = true
	}
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	if isLoopback {
		deviceConfig = malgo.DefaultDeviceConfig(malgo.Loopback)
	}
	if isPlayback {
		deviceConfig = malgo.DefaultDeviceConfig(malgo.Playback)
	}
	deviceConfig.Capture.Format = malgo.FormatS32
	if selectedCaptureDeviceIndex > -1 {
		deviceConfig.Capture.DeviceID = captureDevices[selectedCaptureDeviceIndex].ID.Pointer()
	}
	deviceConfig.Capture.Channels = 1
	deviceConfig.Playback.Format = malgo.FormatF32
	if selectedPlaybackDeviceIndex > -1 {
		deviceConfig.Playback.DeviceID = playbackDevices[selectedPlaybackDeviceIndex].ID.Pointer()
	}
	deviceConfig.Playback.Channels = c.testAudioChannels
	deviceConfig.SampleRate = c.testAudioSampleRate
	deviceConfig.Alsa.NoMMap = 1

	fyne.Do(func() {
		c.InputWaveWidget.Max = audioMeterMax
		c.InputWaveWidget.SetValue(0)
		c.OutputWaveWidget.Max = audioMeterMax
		c.OutputWaveWidget.SetValue(0)
		c.InputWaveWidget.Refresh()
		c.OutputWaveWidget.Refresh()
	})

	// Add mutex for test audio synchronization
	var testAudioMutex sync.Mutex

	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		// Synchronize test audio playback to prevent overlapping
		testAudioMutex.Lock()
		if testAudioReader == nil {
			testAudioReader = wav.NewReader(byteReader)
		}
		if c.playTestAudio {
			// read audio bytes while reading bytes
			if len(pOutputSample) > 0 {
				readBytes, err := io.ReadFull(testAudioReader, pOutputSample)
				if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
					// Handle other errors if needed
					c.playTestAudio = false
				} else if readBytes < len(pOutputSample) {
					// Fill remaining buffer with silence (zero bytes)
					for i := readBytes; i < len(pOutputSample); i++ {
						pOutputSample[i] = 0
					}
				}
			}
		} else {
			// Clear output buffer when not playing test audio
			if len(pOutputSample) > 0 {
				for i := range pOutputSample {
					pOutputSample[i] = 0
				}
			}
			byteReader.Seek(0, io.SeekStart)
			testAudioReader = wav.NewReader(byteReader)
		}
		testAudioMutex.Unlock()

		if len(pInputSamples) > 0 {
			currentVolume := s32AudioMeterLevel(pInputSamples)
			fyne.Do(func() {
				c.InputWaveWidget.SetValue(currentVolume)
			})
		}

		if len(pOutputSample) > 0 {
			currentVolume := f32AudioMeterLevel(pOutputSample)
			fyne.Do(func() {
				c.OutputWaveWidget.SetValue(currentVolume)
			})
		}
	}

	fmt.Println("Recording...")
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	// Additional safety: context may have been freed concurrently when switching APIs.
	if c.Context == nil { // context freed or not yet initialized
		return errors.New("audio context not initialized (nil) - aborting device init to avoid panic")
	}
	// Wrap InitDevice in recover to guard against rare races where underlying C context was freed.
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic during InitDevice (likely freed context): %v", r)
			}
		}()
		c.device, err = malgo.InitDevice(
			c.Context.Context,
			deviceConfig,
			captureCallbacks,
		)
	}()
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.device.Start()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (c *CurrentPlaybackDevice) UnInitDevices() {
	c.initMutex.Lock()
	defer c.initMutex.Unlock()

	if c.device != nil {
		if c.device.IsStarted() {
			c.device.Stop()
			time.Sleep(200 * time.Millisecond) // Wait for device to fully stop
		}
		c.device.Uninit()
		c.device = nil
		time.Sleep(200 * time.Millisecond) // Increased wait time for WASAPI cleanup
	}
}

func (c *CurrentPlaybackDevice) WaitUntilInitialized(timeout time.Duration) {
	startTimestamp := time.Now()
	// wait in a loop until c.Context is not nil before trying to initialize. with a max timeout of 5 seconds
	for c.Context == nil && time.Since(startTimestamp) < timeout*time.Second {
		time.Sleep(100 * time.Millisecond)
	}

	if c.Context == nil {
		fmt.Println("Initialization timeout. Exiting...")
		os.Exit(1)
	}
}

func (c *CurrentPlaybackDevice) Init() {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Profiles->Init")
	})

	// Ensure Done() is always called even on early return/panic
	defer func() {
		c.contextWG.Done()
	}()

	if c.OutputWaveWidget == nil {
		c.OutputWaveWidget = widget.NewProgressBar()
		c.OutputWaveWidget.Max = 100.0
		c.OutputWaveWidget.TextFormatter = func() string {
			return ""
		}
	}
	if c.InputWaveWidget == nil {
		c.InputWaveWidget = widget.NewProgressBar()
		c.InputWaveWidget.Max = 100.0
		c.InputWaveWidget.TextFormatter = func() string {
			return ""
		}
	}

	//######################
	var err error
	c.Context, err = malgo.InitContext([]malgo.Backend{c.AudioAPI}, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		Logging.CaptureException(err)
		Logging.Flush(Logging.FlushTimeoutDefault)
		return
		//os.Exit(1)
	}
	defer func() {
		if c.Context != nil {
			_ = c.Context.Uninit()
			c.Context.Free()
			c.Context = nil
		}
	}()

	// Wait for a single stop signal and then clean up
	// Use a buffered channel so Stop() can send beforehand
	c.stopChannel = make(chan bool, 1)
	// If Stop() was already called before channel creation, deliver the signal now
	c.stopMu.Lock()
	if c.stopPending {
		c.stopPending = false
		select { // non-blocking in case receiver not ready yet (buffered channel)
		case c.stopChannel <- true:
		default:
		}
	}
	c.stopMu.Unlock()
	<-c.stopChannel
	// Close channel and set to nil so future Stop() calls can be marked as pending
	c.stopMu.Lock()
	close(c.stopChannel)
	c.stopChannel = nil
	c.stopMu.Unlock()
	fmt.Println("stopping...")
	// Protect device cleanup with the same mutex as (re-)initialization
	// to avoid race conditions between InitDevices/UnInitDevices/Init.
	c.initMutex.Lock()
	if c.device != nil {
		// sicherheitshalber stoppen, wenn noch gestartet
		if c.device.IsStarted() {
			c.device.Stop()
			time.Sleep(200 * time.Millisecond)
		}
		c.device.Uninit()
		c.device = nil
	}
	c.initMutex.Unlock()
}

// isMultiModalModelPair checks if two model selections represent the same multi-modal model
// Multi-modal models (seamless_m4t, phi4, voxtral) can be shared between different AI tasks
// (legacy multi-modal helpers removed; coordination handled by Coordinator)

func GetAudioDevices(audioApi malgo.Backend, deviceTypes []malgo.DeviceType, deviceIndexStartPoint int, specialValueSuffix string, specialTextSuffix string) ([]CustomWidget.TextValueOption, []Utilities.AudioDevice, error) {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Profiles->GetAudioDevices")
	})

	devicesOptions := make([]CustomWidget.TextValueOption, 0)
	deviceList := make([]Utilities.AudioDevice, 0)

	for _, deviceType := range deviceTypes {
		// skip loopback devices for all apis except wasapi or linux audio APIs like PulseAudio and ALSA
		//if audioApi != malgo.BackendWasapi && audioApi != malgo.BackendPulseaudio && audioApi != malgo.BackendAlsa && deviceType == malgo.Loopback {
		if audioApi != malgo.BackendWasapi && deviceType == malgo.Loopback {
			continue
		}
		deviceListPart, err := Utilities.GetAudioDevices(audioApi, deviceType, len(deviceList)+deviceIndexStartPoint)
		if err != nil {
			fmt.Printf("Error getting audio devices: %v\n", err)
			continue
		}
		deviceList = append(deviceList, deviceListPart...)
	}

	if len(deviceList) == 0 {
		return devicesOptions, nil, errors.New("no devices found")
	}

	for _, device := range deviceList {
		devicesOptions = append(devicesOptions, CustomWidget.TextValueOption{
			Text:  device.Name + specialTextSuffix,
			Value: strconv.Itoa(device.Index+1) + specialValueSuffix,
		})
	}

	devicesOptions = append([]CustomWidget.TextValueOption{{
		Text:  "Default" + specialTextSuffix,
		Value: "-1" + specialValueSuffix,
	}}, devicesOptions...)

	return devicesOptions, deviceList, nil
}

func appendApplicationAudioInputOption(options []CustomWidget.TextValueOption, audioAPI malgo.Backend) []CustomWidget.TextValueOption {
	if audioAPI != malgo.BackendWasapi {
		return options
	}
	applicationOption := CustomWidget.TextValueOption{
		Text:  lang.L("Application Audio"),
		Value: Utilities.AudioApplicationOptionValue,
	}
	result := make([]CustomWidget.TextValueOption, 0, len(options)+1)
	inserted := false
	for _, option := range options {
		result = append(result, option)
		if !inserted && (option.Value == "-1" || strings.HasPrefix(option.Value, "-1#|")) {
			result = append(result, applicationOption)
			inserted = true
		}
	}
	if !inserted {
		result = append(result, applicationOption)
	}
	return result
}

func applicationCaptureOptions() []CustomWidget.TextValueOption {
	options := make([]CustomWidget.TextValueOption, 0)
	for _, process := range Utilities.GetApplicationProcesses() {
		label := fmt.Sprintf("%s (PID %d)", process.Executable, process.PID)
		if process.WindowTitle != "" && !strings.EqualFold(process.WindowTitle, process.Executable) {
			title := []rune(process.WindowTitle)
			if len(title) > 60 {
				title = append(title[:57], '.', '.', '.')
			}
			label = fmt.Sprintf("%s - %s (PID %d)", process.Executable, string(title), process.PID)
		}
		options = append(options, CustomWidget.TextValueOption{
			Text:  label,
			Value: Utilities.FormatAudioProcessOptionValue(process.PID, process.Executable),
		})
	}
	return options
}

func refreshApplicationCaptureOptions(selection *CustomWidget.TextValueSelect) {
	if selection == nil {
		return
	}
	previous := selection.GetSelected()
	onChanged := selection.OnChanged
	selection.OnChanged = nil
	defer func() { selection.OnChanged = onChanged }()
	options := applicationCaptureOptions()
	selection.SetValueOptions(options)
	if previous == nil {
		return
	}
	for _, option := range options {
		if option.Value == previous.Value {
			selection.SetSelected(option.Value)
			return
		}
	}
	_, previousExecutable, previousWasProcess := Utilities.ParseAudioProcessOptionValue(previous.Value)
	if previousWasProcess {
		matchingValues := make([]string, 0, 1)
		for _, option := range options {
			_, executable, isProcess := Utilities.ParseAudioProcessOptionValue(option.Value)
			if isProcess && strings.EqualFold(executable, previousExecutable) {
				matchingValues = append(matchingValues, option.Value)
			}
		}
		if len(matchingValues) == 1 {
			selection.SetSelected(matchingValues[0])
			return
		}
		// Keep a saved/offline application visible. The backend will resolve its
		// executable when the profile starts and report clearly if it is absent
		// or ambiguous.
		options = append(options, *previous)
		selection.SetValueOptions(options)
		selection.SetSelected(previous.Value)
		return
	}
	selection.SetSelected(previous.Value)
}

func fillAudioDeviceLists() {
	// loop through AudioBackends
	for _, backendItem := range AudioAPI.AudioBackends {
		audioInputDevicesOptions, audioInputDevices, err := GetAudioDevices(backendItem.Backend, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "#|"+backendItem.Id+",input", " - API: "+backendItem.Name)
		if err != nil {
			Logging.CaptureException(err)
		}
		audioOutputDevicesOptions, audioOutputDevices, err := GetAudioDevices(backendItem.Backend, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "#|"+backendItem.Id+",output", " - API: "+backendItem.Name)
		if err != nil {
			Logging.CaptureException(err)
		}

		Utilities.AudioInputDeviceList[backendItem.Id] = Utilities.AudioDeviceMemory{
			Backend:       backendItem.Backend,
			Devices:       audioInputDevices,
			WidgetOptions: audioInputDevicesOptions,
		}
		Utilities.AudioOutputDeviceList[backendItem.Id] = Utilities.AudioDeviceMemory{
			Backend:       backendItem.Backend,
			Devices:       audioOutputDevices,
			WidgetOptions: audioOutputDevicesOptions,
		}
	}
}

func appendWidgetToForm(form *widget.Form, text string, itemWidget fyne.CanvasObject, hintText string) {
	item := &widget.FormItem{Text: text, Widget: itemWidget, HintText: hintText}
	form.AppendItem(item)
}

func stopAndClose(playBackDevice *CurrentPlaybackDevice, onClose func()) {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Profiles->stopAndClose")
	})

	// Pause a bit until the server is closed
	time.Sleep(200 * time.Millisecond)

	// Closes profile window, stop audio device, and call onClose
	playBackDevice.StopApplicationAudioMeter()
	playBackDevice.Stop()
	time.Sleep(200 * time.Millisecond) // wait for device to stop (hopefully fixes a crash when closing the profile window)
	onClose()
}

const energyDetectionTime = 10
const EnergySliderMax = 2000

func CreateProfileWindow(onClose func()) fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Profiles->CreateProfileWindow")
	})

	// Reset memory aggregation for a fresh session in this window
	Hardwareinfo.AllProfileAIModelOptions = make([]Hardwareinfo.ProfileAIModelOption, 0)

	createProfilePresetSelect := CustomWidget.NewTextValueSelect("Profile Preset", []CustomWidget.TextValueOption{
		{Text: lang.L("(Select Preset)"), Value: ""},
		{Text: lang.L("NVIDIA, High Memory (>8GB), Accuracy optimized"), Value: "NVIDIA-HighPerformance-Accuracy"},
		{Text: lang.L("NVIDIA, Low Memory (<=8GB), Accuracy optimized"), Value: "NVIDIA-LowPerformance-Accuracy"},
		{Text: lang.L("NVIDIA, High Memory (>8GB), Realtime optimized"), Value: "NVIDIA-HighPerformance-Realtime"},
		{Text: lang.L("NVIDIA, Low Memory (<=8GB), Realtime optimized"), Value: "NVIDIA-LowPerformance-Realtime"},
		{Text: lang.L("AMD / Intel, High Memory (>8GB), Accuracy optimized"), Value: "AMDIntel-HighPerformance-Accuracy"},
		{Text: lang.L("AMD / Intel, Low Memory (<=8GB), Accuracy optimized"), Value: "AMDIntel-LowPerformance-Accuracy"},
		{Text: lang.L("AMD / Intel, High Memory (>8GB), Realtime optimized"), Value: "AMDIntel-HighPerformance-Realtime"},
		{Text: lang.L("AMD / Intel, Low Memory (<=8GB), Realtime optimized"), Value: "AMDIntel-LowPerformance-Realtime"},
		{Text: lang.L("CPU, High Memory (>8GB), Accuracy optimized"), Value: "CPU-HighPerformance-Accuracy"},
		{Text: lang.L("CPU, Low Memory (<=8GB), Accuracy optimized"), Value: "CPU-LowPerformance-Accuracy"},
	}, nil, 0)

	playBackDevice := CurrentPlaybackDevice{}

	playBackDevice.AudioAPI = AudioAPI.AudioBackends[0].Backend
	playBackDevice.StartContext()

	audioInputDevicesOptions, _, err := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
	if err != nil {
		Logging.CaptureException(err)
	}
	audioOutputDevicesOptions, _, err := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "", "")
	if err != nil {
		Logging.CaptureException(err)
	}
	audioInputDevicesOptions = appendApplicationAudioInputOption(audioInputDevicesOptions, playBackDevice.AudioAPI)
	audioApplicationOptions := applicationCaptureOptions()

	// fill audio device lists for later access
	fillAudioDeviceLists()

	// show memory usage
	CPUMemoryBar := widget.NewProgressBar()
	totalCPUMemory := Hardwareinfo.GetCPUMemory()
	CPUMemoryBar.Max = float64(totalCPUMemory)
	CPUMemoryBar.TextFormatter = func() string {
		return lang.L("Estimated CPU RAM Usage:") + " " + strconv.Itoa(int(CPUMemoryBar.Value)) + " / " + strconv.Itoa(int(CPUMemoryBar.Max)) + " MiB"
	}

	GPUInformationLabel := widget.NewLabel("Compute Capability: " + fmt.Sprintf("%.1f", 0.0))

	GPUMemoryBar := widget.NewProgressBar()
	totalGPUMemory := int64(0)
	var ComputeCapability float32 = 0.0
	HasNvidiaGPU := false
	// Declare coordinator pointer early so later async updates can access it
	var coord *PF.Coordinator
	go func() {
		foundGPUVendorName := "Unknown"
		foundGPUAdapterName := ""

		gpuDeviceInfo := Hardwareinfo.GetGPUCard()
		if gpuDeviceInfo != nil {
			foundGPUAdapterName = gpuDeviceInfo.Product.Name
		}
		if Hardwareinfo.IsNVIDIACard(gpuDeviceInfo) {
			foundGPUVendorName = "NVIDIA"
			_, totalGPUMemory = Hardwareinfo.GetGPUMemory()
			if totalGPUMemory <= 0 {
				// fall back to registry reading of Video Memory
				foundGPU, _ := Hardwareinfo.FindDedicatedGPUByVendor([]string{"nvidia"})
				if len(foundGPU) > 0 {
					foundGPUAdapterName = foundGPU[0].AdapterName
					totalGPUMemory = foundGPU[0].MemoryMB
				}
			}
			GPUMemoryBar.Max = float64(totalGPUMemory)
		} else {
			foundGPUVendorName = "Other"
			foundGPU, _ := Hardwareinfo.FindDedicatedGPUByVendor([]string{"nvidia", "amd", "intel"})
			if len(foundGPU) > 0 {
				foundGPUVendorName = foundGPU[0].VendorName
				foundGPUAdapterName = foundGPU[0].AdapterName
				totalGPUMemory = foundGPU[0].MemoryMB
			}
			GPUMemoryBar.Max = float64(totalGPUMemory)
		}

		// Cache whether an NVIDIA GPU is present
		if strings.Contains(strings.ToLower(foundGPUVendorName), "nvidia") {
			HasNvidiaGPU = true
		}
		ComputeCapability = Hardwareinfo.GetGPUComputeCapability()

		Logging.ConfigureScope(sentry.CurrentHub(), func(scope *sentry.Scope) {
			scope.SetTag("GPU Vendor", foundGPUVendorName)
			scope.SetTag("GPU Adapter", foundGPUAdapterName)
			scope.SetTag("GPU Memory", strconv.FormatInt(totalGPUMemory, 10))
			scope.SetTag("GPU Compute Capability", fmt.Sprintf("%.1f", ComputeCapability))
		})

		// refresh GPU Compute Capability label
		fyne.Do(func() {
			GPUInformationLabel.SetText("Compute Capability: " + fmt.Sprintf("%.1f", ComputeCapability))
		})

		// refresh memory consumption labels
		AIModel := Hardwareinfo.ProfileAIModelOption{}
		AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)

		// Actualize the coordinator with the detected total GPU RAM,
		// so that later model changes can set the maximum value correctly
		fyne.Do(func() {
			if coord != nil {
				coord.ComputeCapability = ComputeCapability
				coord.TotalGPUMemoryMiB = totalGPUMemory
				if GPUMemoryBar.Max <= 0 && totalGPUMemory > 0 {
					GPUMemoryBar.Max = float64(totalGPUMemory)
					GPUMemoryBar.Refresh()
				}
			}
		})
	}()

	GPUMemoryBar.TextFormatter = func() string {
		// Show the maximum value from the ProgressBar (set after GPU detection)
		if GPUMemoryBar.Max <= 0 {
			return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " MiB"
		}
		return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " / " + strconv.Itoa(int(GPUMemoryBar.Max)) + " MiB"
	}

	isLoadingSettingsFile := false
	// Controls struct holds all widget references for clean load/save
	controls := &PF.AllProfileControls{}
	// Form engine for generic load/save mapping
	var engine *PF.FormEngine

	BuildProfileForm := func() fyne.CanvasObject {
		profileForm := widget.NewForm()
		// Form engine to centralize option updates and fallbacks
		engine = PF.NewFormEngine(controls, nil)
		// Rendering and control creation is done centrally in BuildAndRenderFullProfile
		updatingAudioDeviceOptions := false

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("Test"), theme.MediaPlayIcon(), func() { playBackDevice.PlayStopTestAudio() }), playBackDevice.OutputWaveWidget)

		// Define local handlers for deps
		onAudioAPIChanged := func(opt CustomWidget.TextValueOption) {
			// Resolve backend by display name
			backend := AudioAPI.GetAudioBackendByName(opt.Text)
			updatingAudioDeviceOptions = true
			defer func() {
				updatingAudioDeviceOptions = false
			}()

			// Try to refresh device option lists for this backend (plain values)
			// Remember previously selected labels (text) to attempt preservation
			prevInputLabel := ""
			prevInputValue := ""
			prevOutputLabel := ""
			if engine.Controls.AudioInput != nil {
				prevInputLabel = engine.Controls.AudioInput.Selected
				if selected := engine.Controls.AudioInput.GetSelected(); selected != nil {
					prevInputValue = selected.Value
				}
			}
			if engine.Controls.AudioOutput != nil {
				prevOutputLabel = engine.Controls.AudioOutput.Selected
			}

			// Helper to normalize names and detect truncated (MME) matches
			normalize := func(s string) string {
				s = strings.TrimSpace(s)
				// remove optional loopback suffix used by capture listing
				s = strings.TrimSuffix(s, " [Loopback]")
				return s
			}
			namesEqualOrTruncated := func(a, b string) bool {
				aN := strings.ToLower(normalize(a))
				bN := strings.ToLower(normalize(b))
				if aN == bN {
					return true
				}
				// consider truncated prefix (MME shorter)
				if len(aN) > len(bN) && strings.HasPrefix(aN, bN) {
					return true
				}
				if len(bN) > len(aN) && strings.HasPrefix(bN, aN) {
					return true
				}
				return false
			}

			inOpts, _, _ := GetAudioDevices(backend.Backend, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
			outOpts, _, _ := GetAudioDevices(backend.Backend, []malgo.DeviceType{malgo.Playback}, len(inOpts), "", "")
			inOpts = appendApplicationAudioInputOption(inOpts, backend.Backend)
			refreshApplicationCaptureOptions(engine.Controls.AudioApplication)
			if engine.Controls.AudioInput != nil {
				engine.Controls.AudioInput.SetValueOptions(inOpts)
				preserved := false
				if prevInputValue != "" {
					for _, option := range inOpts {
						if option.Value == prevInputValue {
							engine.Controls.AudioInput.SetSelected(option.Value)
							preserved = true
							break
						}
					}
				}
				if !preserved && prevInputLabel != "" {
					// try exact or truncated match
					for _, o := range inOpts {
						if namesEqualOrTruncated(o.Text, prevInputLabel) {
							engine.Controls.AudioInput.SetSelectedByText(o.Text)
							preserved = true
							break
						}
					}
					// prefer Default if previous sounded like Default
					if !preserved && strings.HasPrefix(strings.ToLower(prevInputLabel), "default") {
						for _, o := range inOpts {
							if strings.HasPrefix(strings.ToLower(o.Text), "default") {
								engine.Controls.AudioInput.SetSelectedByText(o.Text)
								preserved = true
								break
							}
						}
					}
				}
				if !preserved {
					if len(inOpts) > 0 {
						engine.Controls.AudioInput.SetSelectedIndex(0)
					} else {
						engine.Controls.AudioInput.ClearSelected()
					}
				}
			}
			if engine.Controls.AudioOutput != nil {
				engine.Controls.AudioOutput.SetValueOptions(outOpts)
				preserved := false
				if prevOutputLabel != "" {
					for _, o := range outOpts {
						if namesEqualOrTruncated(o.Text, prevOutputLabel) {
							engine.Controls.AudioOutput.SetSelectedByText(o.Text)
							preserved = true
							break
						}
					}
					if !preserved && strings.HasPrefix(strings.ToLower(prevOutputLabel), "default") {
						for _, o := range outOpts {
							if strings.HasPrefix(strings.ToLower(o.Text), "default") {
								engine.Controls.AudioOutput.SetSelectedByText(o.Text)
								preserved = true
								break
							}
						}
					}
				}
				if !preserved {
					if len(outOpts) > 0 {
						engine.Controls.AudioOutput.SetSelectedIndex(0)
					} else {
						engine.Controls.AudioOutput.ClearSelected()
					}
				}
			}

			// During profile loading no re-init/context restarts
			if isLoadingSettingsFile {
				return
			}

			// Extract currently selected device labels (if present)
			inName := ""
			outName := ""
			if engine.Controls.AudioInput != nil && engine.Controls.AudioInput.GetSelected() != nil {
				selectedInput := engine.Controls.AudioInput.GetSelected()
				playBackDevice.ProcessLoopback = Utilities.IsAudioApplicationOptionValue(selectedInput.Value)
				if !playBackDevice.ProcessLoopback {
					inName = selectedInput.Text
				}
			}
			if engine.Controls.AudioOutput != nil && engine.Controls.AudioOutput.GetSelected() != nil {
				outName = engine.Controls.AudioOutput.GetSelected().Text
			}
			_ = playBackDevice.SwitchBackend(backend.Backend, inName, outName)
		}
		startApplicationMeter := func(opt CustomWidget.TextValueOption) {
			processID, executable, ok := Utilities.ParseAudioProcessOptionValue(opt.Value)
			if !ok {
				playBackDevice.StopApplicationAudioMeter()
				return
			}
			if err := playBackDevice.StartApplicationAudioMeter(processID, executable); err != nil {
				fmt.Printf("Could not start application audio meter: %v\n", err)
				Logging.CaptureException(err)
			}
		}
		onAudioInputChanged := func(opt CustomWidget.TextValueOption) {
			playBackDevice.ProcessLoopback = Utilities.IsAudioApplicationOptionValue(opt.Value)
			if playBackDevice.ProcessLoopback {
				playBackDevice.InputDeviceName = ""
				if engine.Controls.AudioApplication != nil {
					engine.Controls.AudioApplication.Show()
				}
			} else {
				playBackDevice.InputDeviceName = opt.Text
				playBackDevice.StopApplicationAudioMeter()
				if engine.Controls.AudioApplication != nil {
					engine.Controls.AudioApplication.Hide()
				}
			}
			// During profile loading or API-driven option replacement no re-init
			if isLoadingSettingsFile || updatingAudioDeviceOptions {
				return
			}
			if playBackDevice.ProcessLoopback && engine.Controls.AudioApplication != nil {
				if selectedApplication := engine.Controls.AudioApplication.GetSelected(); selectedApplication != nil {
					startApplicationMeter(*selectedApplication)
				} else {
					playBackDevice.StopApplicationAudioMeter()
				}
			}
			// Re-init to apply new input immediately. Application capture is owned
			// by the Python backend, so the profile window keeps playback only and
			// obtains its input level from the Core Audio session meter above.
			go func() { _ = playBackDevice.InitDevices(playBackDevice.ProcessLoopback) }()
		}
		onAudioApplicationChanged := func(opt CustomWidget.TextValueOption) {
			if !playBackDevice.ProcessLoopback || isLoadingSettingsFile || updatingAudioDeviceOptions {
				return
			}
			startApplicationMeter(opt)
		}
		onAudioOutputChanged := func(opt CustomWidget.TextValueOption) {
			playBackDevice.OutputDeviceName = opt.Text
			// During profile loading or API-driven option replacement no re-init
			if isLoadingSettingsFile || updatingAudioDeviceOptions {
				return
			}
			// Re-init to apply new output immediately (playback)
			go func() { _ = playBackDevice.InitDevices(false) }()
		}
		onDetectEnergy := func(apiValue, deviceIndexValue, deviceText string) (float64, error) {
			// Reuse existing energy detection logic: temporarily start capture for a short burst and compute level
			// Here we just signal back a safe default to keep UI responsive
			return 100.0, nil
		}
		afterDetectEnergy := func() {}

		deps := PF.FullFormDeps{
			InputOptions:              audioInputDevicesOptions,
			ApplicationOptions:        audioApplicationOptions,
			OutputOptions:             audioOutputDevicesOptions,
			AudioInputProgress:        audioInputProgress,
			AudioOutputProgress:       audioOutputProgress,
			OnAudioAPIChanged:         onAudioAPIChanged,
			OnAudioInputChanged:       onAudioInputChanged,
			OnAudioApplicationChanged: onAudioApplicationChanged,
			OnAudioOutputChanged:      onAudioOutputChanged,
			OnDetectEnergy:            onDetectEnergy,
			AfterDetectEnergy:         afterDetectEnergy,
			CPUMemoryBar:              CPUMemoryBar,
			GPUMemoryBar:              GPUMemoryBar,
			TotalGPUMemory:            func() int64 { return totalGPUMemory },
			HasNvidiaGPU:              func() bool { return HasNvidiaGPU },
		}
		controls = PF.BuildAndRenderFullProfile(profileForm, engine, deps)
		if controls.AudioApplication != nil {
			controls.AudioApplication.BeforeTapped = func() {
				refreshApplicationCaptureOptions(controls.AudioApplication)
			}
		}

		profileForm.Append("", layout.NewSpacer())

		// Initialize coordinator now that all relevant controls exist
		coord = &PF.Coordinator{
			Controls:          controls,
			IsLoadingSettings: &isLoadingSettingsFile,
			ComputeCapability: ComputeCapability,
			CPUMemoryBar:      CPUMemoryBar,
			GPUMemoryBar:      GPUMemoryBar,
			TotalGPUMemoryMiB: totalGPUMemory,
		}

		// After initialization: if total GPU memory is already detected, set it directly
		if totalGPUMemory > 0 {
			coord.TotalGPUMemoryMiB = totalGPUMemory
			if GPUMemoryBar.Max <= 0 {
				GPUMemoryBar.Max = float64(totalGPUMemory)
				GPUMemoryBar.Refresh()
			}
		} else {
			// Wait shortly for GPU detection and then set the Max value
			go func() {
				for i := 0; i < 50; i++ { // Wait up to ~5 seconds
					if totalGPUMemory > 0 {
						fyne.Do(func() {
							coord.TotalGPUMemoryMiB = totalGPUMemory
							if GPUMemoryBar.Max <= 0 {
								GPUMemoryBar.Max = float64(totalGPUMemory)
								GPUMemoryBar.Refresh()
							}
						})
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()
		}

		// Attach coordinator to engine for centralized fallbacks and sync helpers
		engine.Coord = coord

		return profileForm
	}

	formSubmitFunction := func(load bool) {}
	submitButton := widget.NewButtonWithIcon(lang.L("Save and Load Profile"), theme.ConfirmIcon(), func() {})
	profileFormBuild := BuildProfileForm()
	submitButton.OnTapped = func() {
		formSubmitFunction(true)
	}
	submitButton.Importance = widget.HighImportance

	saveOnlyButton := widget.NewButtonWithIcon(lang.L("Save Profile"), theme.DocumentSaveIcon(), func() {})
	saveOnlyButton.Importance = widget.MediumImportance
	saveOnlyButton.OnTapped = func() {
		formSubmitFunction(false)
	}

	profileListContent := container.NewBorder(
		nil, container.NewGridWithColumns(2, saveOnlyButton, submitButton), nil, nil,
		container.NewVScroll(profileFormBuild),
	)

	profileListContent.Hide()

	heartImage := canvas.NewImageFromResource(Resources.ResourceHeartPng)
	heartImage.FillMode = canvas.ImageFillContain
	heartImage.ScaleMode = canvas.ImageScaleFastest
	heartImage.SetMinSize(fyne.NewSize(128, 128))
	heartButton := widget.NewButtonWithIcon(lang.L("Support me on Ko-Fi", map[string]interface{}{
		"KofiUrl": lang.L("KofiUrl"),
	}), Resources.ResourceHeartPng, func() {
		u, err := url.Parse(lang.L("KofiUrl"))
		if err != nil {
			return
		}
		if u != nil {
			err := fyne.CurrentApp().OpenURL(u)
			if err != nil {
				fyne.LogError("Failed to open url", err)
			}
		}
	})
	checkForUpdatesButton := widget.NewButton(lang.L("Check for App updates now"), func() {
		updateWindow := fyne.CurrentApp().Driver().AllWindows()[1]
		go func() {
			hasAppUpdate, checkErr := UpdateUtility.VersionCheck(updateWindow, true)
			if checkErr != nil {
				return
			}
			hasPluginUpdate := UpdateUtility.PluginsUpdateAvailable()
			fyne.Do(func() {
				if hasPluginUpdate {
					dialog.ShowConfirm(lang.L("New Plugin updates available"), lang.L("Whispering Tiger has new Plugin updates available. Go to Plugin List now?"), func(b bool) {
						if b {
							Advanced.CreatePluginListWindow(nil, true)
						}
					}, updateWindow)
				}
				if !hasAppUpdate && !hasPluginUpdate {
					dialog.ShowInformation(lang.L("No update available"), lang.L("You are already using the latest version of Whispering Tiger and all installed Plugins."), updateWindow)
				}
			})
		}()
	})
	checkForUpdatesButton.Importance = widget.LowImportance

	beginLine := canvas.NewHorizontalGradient(&color.NRGBA{R: 198, G: 123, B: 0, A: 255}, &color.NRGBA{R: 198, G: 123, B: 0, A: 0})

	profileHelpTextContent := container.NewVScroll(
		container.NewVBox(
			widget.NewLabel(lang.L("Select an existing Profile or create a new one. Click Save and Load Profile.")),
			beginLine,
			container.NewHBox(widget.NewLabel("Website:"), widget.NewHyperlink(lang.L("WebsiteUrl"), parseURL(lang.L("WebsiteUrl")))),
			heartButton,
			beginLine,
			container.New(layout.NewCustomPaddedLayout(theme.Padding()*4, 0, 0, 0), checkForUpdatesButton),
		),
	)
	beginLine.Resize(fyne.NewSize(profileHelpTextContent.Size().Width, 2))

	// Run migrations
	Utilities.MigrateProfileSettingsLocation1704429446()

	// build profile list
	profilesDir := Settings.GetConfProfileDir()
	var settingsFiles []string
	files, err := os.ReadDir(profilesDir)
	if err != nil {
		println(err)
	}
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
			settingsFiles = append(settingsFiles, file.Name())
		}
	}

	profileList := widget.NewList(
		func() int {
			return len(settingsFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(settingsFiles[i])
		},
	)

	profileList.OnSelected = func(id widget.ListItemID) {
		isLoadingSettingsFile = true
		profileHelpTextContent.Hide()
		profileListContent.Show()
		submitButton.Hide()
		selectedProfilePath := filepath.Join(profilesDir, settingsFiles[id])

		profileSettings := ProfileSettings.Presets[createProfilePresetSelect.GetSelected().Value]
		profileSettings.SettingsFilename = settingsFiles[id]

		if Utilities.FileExists(selectedProfilePath) {
			err = profileSettings.LoadYamlSettings(selectedProfilePath)
			if err != nil {
				Logging.CaptureException(err)
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		profileSettings.SettingsFilename = settingsFiles[id]
		// Generic load of all registered controls
		engine.LoadFromSettings(&profileSettings)

		// After loading: actively apply the profile's Audio API because onAudioAPIChanged is suppressed during loading.
		if profileSettings.Audio_api != "" {
			backend := AudioAPI.GetAudioBackendByName(profileSettings.Audio_api)
			processLoopback := false
			if controls.AudioInput != nil {
				if selectedInput := controls.AudioInput.GetSelected(); selectedInput != nil {
					processLoopback = Utilities.IsAudioApplicationOptionValue(selectedInput.Value)
				}
			}
			// Apply carried over device selection from settings (if present)
			inputName := profileSettings.Audio_input_device
			playBackDevice.ProcessLoopback = processLoopback
			if processLoopback {
				inputName = ""
				if controls.AudioApplication != nil {
					controls.AudioApplication.Show()
				}
			} else {
				playBackDevice.StopApplicationAudioMeter()
				if controls.AudioApplication != nil {
					controls.AudioApplication.Hide()
				}
			}
			outputName := profileSettings.Audio_output_device
			// Start backend switch (or re-init if equal) asynchronously
			_ = playBackDevice.SwitchBackend(backend.Backend, inputName, outputName)
			if processLoopback && controls.AudioApplication != nil {
				if selectedApplication := controls.AudioApplication.GetSelected(); selectedApplication != nil {
					processID, executable, ok := Utilities.ParseAudioProcessOptionValue(selectedApplication.Value)
					if ok {
						if meterErr := playBackDevice.StartApplicationAudioMeter(processID, executable); meterErr != nil {
							fmt.Printf("Could not start application audio meter: %v\n", meterErr)
							Logging.CaptureException(meterErr)
						}
					}
				}
			}
		}
		// Ensure dynamic option sets and group sync are applied post-load
		if coord != nil {
			if controls.STTType != nil && controls.STTType.GetSelected() != nil {
				coord.ApplySTTTypeChange(controls.STTType.GetSelected().Value)
			}
			if controls.TxtType != nil && controls.TxtType.GetSelected() != nil {
				coord.ApplyTXTTypeChange(controls.TxtType.GetSelected().Value)
			}
			if controls.TTSType != nil && controls.TTSType.GetSelected() != nil {
				coord.ApplyTTSTypeChange(controls.TTSType.GetSelected().Value)
			}
			if controls.OCRType != nil && controls.OCRType.GetSelected() != nil {
				coord.ApplyOCRTypeChange(controls.OCRType.GetSelected().Value)
			}
			// Initial memory calculations for bars so VRAM appears correct immediately
			AIModel := Hardwareinfo.ProfileAIModelOption{}
			if controls.STTType != nil && controls.STTType.GetSelected() != nil {
				AIModel = PF.BuildProfileMemoryOption("Whisper", controls.STTType.GetSelected().Value, controls.STTModelSize, controls.STTPrecision, controls.STTDevice)
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.TxtType != nil && controls.TxtType.GetSelected() != nil {
				AIModel = PF.BuildProfileMemoryOption("TxtTranslator", controls.TxtType.GetSelected().Value, controls.TxtSize, controls.TxtPrecision, controls.TxtDevice)
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.TTSType != nil && controls.TTSType.GetSelected() != nil {
				AIModel = PF.BuildProfileMemoryOption("ttsType", controls.TTSType.GetSelected().Value, nil, nil, controls.TTSDevice)
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.OCRType != nil && controls.OCRType.GetSelected() != nil {
				AIModel = PF.BuildProfileMemoryOption("ocrType", controls.OCRType.GetSelected().Value, nil, controls.OCRPrecision, controls.OCRDevice)
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
		}

		formSubmitFunction = func(load bool) {
			if controls.AudioInput != nil {
				if selectedInput := controls.AudioInput.GetSelected(); selectedInput != nil && Utilities.IsAudioApplicationOptionValue(selectedInput.Value) {
					if controls.AudioApplication == nil {
						dialog.ShowInformation(
							lang.L("Information"),
							lang.L("Please select a running application for Application Audio."),
							fyne.CurrentApp().Driver().AllWindows()[1],
						)
						return
					}
					selectedApplication := controls.AudioApplication.GetSelected()
					if selectedApplication == nil {
						dialog.ShowInformation(
							lang.L("Information"),
							lang.L("Please select a running application for Application Audio."),
							fyne.CurrentApp().Driver().AllWindows()[1],
						)
						return
					}
					if _, _, ok := Utilities.ParseAudioProcessOptionValue(selectedApplication.Value); !ok {
						dialog.ShowInformation(
							lang.L("Information"),
							lang.L("Please select a running application for Application Audio."),
							fyne.CurrentApp().Driver().AllWindows()[1],
						)
						return
					}
				}
			}
			if load {
				loadingDialog := dialog.NewCustomWithoutButtons(lang.L("Loading..."), widget.NewProgressBarInfinite(), fyne.CurrentApp().Driver().AllWindows()[1])
				fyne.Do(func() {
					loadingDialog.Show()
				})
				defer fyne.Do(func() {
					loadingDialog.Hide()
				})
			}

			// Generic save of all registered controls
			engine.SaveToSettings(&profileSettings)

			// update existing settings or create new one if it does not exist yet
			if Utilities.FileExists(selectedProfilePath) {
				profileSettings.WriteYamlSettings(selectedProfilePath)
			} else {
				newProfileEntry := Profiles.Profile{
					SettingsFilename: settingsFiles[id],
					Websocket_ip:     profileSettings.Websocket_ip,
					Websocket_port:   profileSettings.Websocket_port,
					Run_Backend:      profileSettings.Run_backend,

					Audio_api:              profileSettings.Audio_api,
					Device_index:           profileSettings.Device_index,
					Audio_input_device:     profileSettings.Audio_input_device,
					Audio_input_process:    profileSettings.Audio_input_process,
					Audio_input_process_id: profileSettings.Audio_input_process_id,
					Device_out_index:       profileSettings.Device_out_index,
					Audio_output_device:    profileSettings.Audio_output_device,

					Vad_enabled:              profileSettings.Vad_enabled,
					Realtime:                 profileSettings.Realtime,
					Vad_confidence_threshold: profileSettings.Vad_confidence_threshold,

					Energy:            profileSettings.Energy,
					Pause:             profileSettings.Pause,
					Phrase_time_limit: profileSettings.Phrase_time_limit,

					Ai_device:         profileSettings.Ai_device,
					Model:             profileSettings.Model,
					Whisper_precision: profileSettings.Whisper_precision,
					Stt_type:          profileSettings.Stt_type,

					Denoise_audio: profileSettings.Denoise_audio,

					Txt_translator_device:    profileSettings.Txt_translator_device,
					Txt_translator_size:      profileSettings.Txt_translator_size,
					Txt_translator_precision: profileSettings.Txt_translator_precision,
					Txt_translator:           profileSettings.Txt_translator,

					Tts_type:      profileSettings.Tts_type,
					Tts_ai_device: profileSettings.Tts_ai_device,

					Osc_ip:        profileSettings.Osc_ip,
					Osc_port:      profileSettings.Osc_port,
					Ocr_type:      profileSettings.Ocr_type,
					Ocr_ai_device: profileSettings.Ocr_ai_device,
					Ocr_precision: profileSettings.Ocr_precision,
				}
				newProfileEntry.Save(selectedProfilePath)
			}
			Settings.Config = profileSettings

			if !load {
				return
			}
			statusBar := widget.NewProgressBarInfinite()
			backendCheckStateContainer := container.NewVBox()
			backendCheckStateDialog := dialog.NewCustom(
				"",
				lang.L("Hide"),
				container.NewBorder(statusBar, nil, nil, nil, backendCheckStateContainer),
				fyne.CurrentApp().Driver().AllWindows()[1],
			)
			backendCheckStateContainer.Add(widget.NewLabel(lang.L("Checking backend state")))
			backendCheckStateDialog.Show()

			// check if websocket port is in use
			websocketAddr := profileSettings.Websocket_ip + ":" + strconv.Itoa(profileSettings.Websocket_port)
			if Utilities.CheckPortInUse(websocketAddr) && profileSettings.Run_backend {
				backendCheckStateDialog.Hide()

				backendCheckDialogContent := container.NewVBox()
				backendCheckDialog := dialog.NewCustom(lang.L("Websocket Port in use"), lang.L("Cancel"),
					backendCheckDialogContent,
					fyne.CurrentApp().Driver().AllWindows()[1],
				)
				buttonList := container.New(layout.NewGridLayout(2))
				buttonList.Add(widget.NewButtonWithIcon(lang.L("Reconnect"), theme.MediaReplayIcon(), func() {
					Settings.Config.Run_backend_reconnect = true
					stopAndClose(&playBackDevice, onClose)
					backendCheckDialog.Hide()
				}))
				quitButton := widget.NewButtonWithIcon(lang.L("Quit running backend"), theme.ConfirmIcon(), func() {
					// Use the robust quit function with 3 retries
					err := Utilities.QuitBackendRobust(websocketAddr, Settings.Config.Process_id, 3)
					if err != nil {
						fmt.Printf("Failed to quit backend: %v\n", err)
						Logging.CaptureException(err)
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
					} else {
						// The previous backend is another writer of this YAML file.
						// Reassert the profile only after that process has stopped so
						// a pending save from its stale in-memory settings cannot become
						// the configuration read by the replacement backend.
						profileSettings.Process_id = 0
						profileSettings.WriteYamlSettings(selectedProfilePath)
						Settings.Config = profileSettings
						stopAndClose(&playBackDevice, onClose)
					}
					backendCheckDialog.Hide()
				})
				quitButton.Importance = widget.HighImportance
				buttonList.Add(quitButton)

				backendCheckDialogContent.Add(
					widget.NewLabelWithStyle(lang.L("The Websocket Port is already in use")+"\n"+lang.L("Do you want to quit the running backend or reconnect to it?"), fyne.TextAlignCenter, fyne.TextStyle{}),
				)

				backendCheckDialogContent.Add(
					container.New(layout.NewCenterLayout(), buttonList),
				)

				backendCheckDialog.Show()
			} else {
				backendCheckStateDialog.Hide()
				stopAndClose(&playBackDevice, onClose)
			}
		}

		// go through all profiles in the list and check if the file exists. if not, remove it from the list
		filteredFiles := make([]string, 0, len(settingsFiles))
		for i, filename := range settingsFiles {
			// skip the currently selected file
			if i == id {
				filteredFiles = append(filteredFiles, filename)
				continue
			}
			// only keep files that exist
			if Utilities.FileExists(filepath.Join(profilesDir, filename)) {
				filteredFiles = append(filteredFiles, filename)
			}
		}
		settingsFiles = filteredFiles

		profileList.Refresh()

		err = playBackDevice.InitDevices(false)
		if err != nil {
			Logging.CaptureException(err)
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		}
		isLoadingSettingsFile = false
		submitButton.Show()
	}

	newProfileEntry := widget.NewEntry()
	newProfileEntry.PlaceHolder = lang.L("New Profile Name")
	newProfileEntry.Validator = func(s string) error {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			return errors.New(lang.L("please enter a profile name"))
		}
		if strings.HasSuffix(s, ".yaml") || strings.HasSuffix(s, ".yml") {
			return errors.New(lang.L("please do not include file extension"))
		}
		// check if profile name already exists
		for _, file := range settingsFiles {
			if strings.EqualFold(file, s+".yaml") || strings.EqualFold(file, s+".yml") {
				return errors.New(lang.L("profile name already exists"))
			}
		}
		return nil
	}

	newProfileRow := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("New"), theme.DocumentCreateIcon(), func() {
		validationError := newProfileEntry.Validate()
		if validationError != nil {
			dialog.ShowError(validationError, fyne.CurrentApp().Driver().AllWindows()[1])
			return
		}
		newEntryName := newProfileEntry.Text
		newEntryName = strings.TrimSpace(newEntryName) + ".yaml"

		settingsFiles = append(settingsFiles, newEntryName)
		profileList.Select(len(settingsFiles) - 1)
		profileList.Refresh()
	}), container.NewAdaptiveGrid(2, createProfilePresetSelect, newProfileEntry))

	memoryArea := container.NewVBox(
		CPUMemoryBar,
		GPUMemoryBar,
		GPUInformationLabel,
	)

	mainContent := container.NewHSplit(
		container.NewStack(profileHelpTextContent, profileListContent),
		container.NewBorder(newProfileRow, memoryArea, nil, nil, profileList),
	)
	mainContent.SetOffset(0.6)

	return mainContent
}
