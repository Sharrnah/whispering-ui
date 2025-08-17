package Pages

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/ProfileSettings"
	PF "whispering-tiger-ui/ProfileForm"
	"whispering-tiger-ui/Profiles"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
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
	// stop-Koordination, um verlorene Stop-Signale vor Kanal-Initialisierung zu vermeiden
	stopPending bool
	stopMu      sync.Mutex
}

func (c *CurrentPlaybackDevice) Stop() {
	// Nicht-blockierendes Stop-Signal. Falls der Kanal noch nicht existiert,
	// merken wir uns das Stoppen als Pending und liefern es nach.
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

// From arduino map() lol
func int32Map(x int32, in_min int32, in_max int32, out_min int32, out_max int32) int32 {
	var _x = int64(x)
	var _in_min = int64(in_min)
	var _in_max = int64(in_max)
	var _out_min = int64(out_min)
	var _out_max = int64(out_max)
	var r = int64((_x-_in_min)*(_out_max-_out_min)/(_in_max-_in_min) + _out_min)
	return int32(r)
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

	sizeInBytesCapture := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	sizeInBytesPlayback := uint32(malgo.SampleSizeInBytes(deviceConfig.Playback.Format))

	c.InputWaveWidget.Max = 0.1
	fyne.Do(func() {
		c.InputWaveWidget.Refresh()
	})

	// Add mutex for test audio synchronization
	var testAudioMutex sync.Mutex

	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		sampleCountCapture := framecount * deviceConfig.Capture.Channels * sizeInBytesCapture
		sampleCountPlayback := framecount * deviceConfig.Playback.Channels * sizeInBytesPlayback

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

		// single samples inside a frame
		if pInputSamples != nil {
			sampleVolume := 0.0
			singleSampleSize := deviceConfig.Capture.Channels * sizeInBytesCapture
			for i := uint32(0); i < sampleCountCapture; i += singleSampleSize {
				sample := binary.LittleEndian.Uint32(pInputSamples[i : i+singleSampleSize])
				sampleHeight := int32Map(int32(sample), 0, math.MaxInt32, 0, 100)

				sampleVolume += math.Max(0, float64(sampleHeight))
			}

			currentVolume := sampleVolume / float64(framecount)
			if currentVolume >= 0 {
				fyne.Do(func() {
					if c.InputWaveWidget.Max < currentVolume {
						c.InputWaveWidget.Max = currentVolume * 2
						c.InputWaveWidget.Refresh()
					}
					c.InputWaveWidget.SetValue(currentVolume)
				})
			}
		}

		if pOutputSample != nil {
			sampleVolume := 0.0
			singleSampleSize := deviceConfig.Playback.Channels * sizeInBytesPlayback
			for i := uint32(0); i < sampleCountPlayback; i += singleSampleSize {
				sample := binary.LittleEndian.Uint32(pOutputSample[i : i+singleSampleSize])
				sampleHeight := int32Map(int32(sample), 0, math.MaxInt32, 0, 100)

				sampleVolume += math.Max(0, float64(sampleHeight))
			}

			currentVolume := sampleVolume / float64(framecount)
			if currentVolume >= 0 {
				fyne.Do(func() {
					c.OutputWaveWidget.SetValue(currentVolume)
				})
			}
		}
	}

	fmt.Println("Recording...")
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}
	c.device, err = malgo.InitDevice(c.Context.Context, deviceConfig, captureCallbacks)
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

	// Warte auf ein einzelnes Stop-Signal und dann aufräumen
	// Verwende einen gepufferten Kanal, sodass Stop() schon vorher senden kann
	c.stopChannel = make(chan bool, 1)
	// Wenn Stop() bereits vor der Kanal-Erstellung aufgerufen wurde, Signal nachliefern
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
	// Kanal schließen und auf nil setzen, um künftige Stop()-Aufrufe als pending markieren zu können
	c.stopMu.Lock()
	close(c.stopChannel)
	c.stopChannel = nil
	c.stopMu.Unlock()
	fmt.Println("stopping...")
	// Schütze Geräte-Cleanup mit dem gleichen Mutex wie die (Re-)Initialisierung,
	// um Race-Conditions zwischen InitDevices/UnInitDevices/Init zu vermeiden.
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
	go playBackDevice.Init()

	audioInputDevicesOptions, _, err := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
	if err != nil {
		Logging.CaptureException(err)
	}
	audioOutputDevicesOptions, _, err := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "", "")
	if err != nil {
		Logging.CaptureException(err)
	}

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
	// Coordinator-Zeiger früh deklarieren, damit spätere Async-Updates zugreifen können
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

		// Cache, ob NVIDIA vorhanden ist
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

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("Test"), theme.MediaPlayIcon(), func() { playBackDevice.PlayStopTestAudio() }), playBackDevice.OutputWaveWidget)

		// Define local handlers for deps
		onAudioAPIChanged := func(opt CustomWidget.TextValueOption) {
			// Resolve backend by display name
			backend := AudioAPI.GetAudioBackendByName(opt.Text)

			// Backend im Gerät setzen (UI-Optionen aktualisieren wir gleich); teure Re-Inits nur wenn nicht geladen wird
			playBackDevice.AudioAPI = backend.Backend

			// Try to refresh device option lists for this backend (plain values)
			// Remember previously selected labels (text) to attempt preservation
			prevInputLabel := ""
			prevOutputLabel := ""
			if engine.Controls.AudioInput != nil {
				prevInputLabel = engine.Controls.AudioInput.Selected
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

			inOpts, _, _ := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
			outOpts, _, _ := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(inOpts), "", "")
			if engine.Controls.AudioInput != nil {
				engine.Controls.AudioInput.Options = inOpts
				preserved := false
				if prevInputLabel != "" {
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
				engine.Controls.AudioOutput.Options = outOpts
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

			// Während Profil-Laden keine Re-Init/Context-Neustarts
			if isLoadingSettingsFile {
				return
			}

			// Stop current context/device and switch backend (teuer)
			playBackDevice.Stop()          // end malgo context goroutine
			playBackDevice.UnInitDevices() // ensure device is stopped/uninitialized
			// Start a fresh malgo context for the new backend
			go playBackDevice.Init()
			// Apply current selections to playback device and re-init
			go func() {
				if engine.Controls.AudioInput != nil && engine.Controls.AudioInput.GetSelected() != nil {
					playBackDevice.InputDeviceName = engine.Controls.AudioInput.GetSelected().Text
				}
				if engine.Controls.AudioOutput != nil && engine.Controls.AudioOutput.GetSelected() != nil {
					playBackDevice.OutputDeviceName = engine.Controls.AudioOutput.GetSelected().Text
				}
				playBackDevice.WaitUntilInitialized(5)
				_ = playBackDevice.InitDevices(false)
			}()
		}
		onAudioInputChanged := func(opt CustomWidget.TextValueOption) {
			playBackDevice.InputDeviceName = opt.Text
			// Während Profil-Laden kein Re-Init
			if isLoadingSettingsFile {
				return
			}
			// Re-init to apply new input immediately
			go func() { _ = playBackDevice.InitDevices(false) }()
		}
		onAudioOutputChanged := func(opt CustomWidget.TextValueOption) {
			playBackDevice.OutputDeviceName = opt.Text
			// Während Profil-Laden kein Re-Init
			if isLoadingSettingsFile {
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
			InputOptions:         audioInputDevicesOptions,
			OutputOptions:        audioOutputDevicesOptions,
			AudioInputProgress:   audioInputProgress,
			AudioOutputProgress:  audioOutputProgress,
			OnAudioAPIChanged:    onAudioAPIChanged,
			OnAudioInputChanged:  onAudioInputChanged,
			OnAudioOutputChanged: onAudioOutputChanged,
			OnDetectEnergy:       onDetectEnergy,
			AfterDetectEnergy:    afterDetectEnergy,
			CPUMemoryBar:         CPUMemoryBar,
			GPUMemoryBar:         GPUMemoryBar,
			TotalGPUMemory:       func() int64 { return totalGPUMemory },
			HasNvidiaGPU:         func() bool { return HasNvidiaGPU },
		}
		controls = PF.BuildAndRenderFullProfile(profileForm, engine, deps)

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

	formSubmitFunction := func() {}
	submitButton := widget.NewButtonWithIcon(lang.L("Save and Load Profile"), theme.ConfirmIcon(), func() {})
	profileFormBuild := BuildProfileForm()
	submitButton.OnTapped = func() {
		go formSubmitFunction()
	}

	submitButton.Importance = widget.HighImportance
	profileListContent := container.NewBorder(
		nil, submitButton, nil, nil,
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

	beginLine := canvas.NewHorizontalGradient(&color.NRGBA{R: 198, G: 123, B: 0, A: 255}, &color.NRGBA{R: 198, G: 123, B: 0, A: 0})

	profileHelpTextContent := container.NewVScroll(
		container.NewVBox(
			widget.NewLabel(lang.L("Select an existing Profile or create a new one. Click Save and Load Profile.")),
			beginLine,
			container.NewHBox(widget.NewLabel("Website:"), widget.NewHyperlink(lang.L("WebsiteUrl"), parseURL(lang.L("WebsiteUrl")))),
			heartButton,
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

		profileSettings := ProfileSettings.Presets[createProfilePresetSelect.GetSelected().Value]
		profileSettings.SettingsFilename = settingsFiles[id]

		if Utilities.FileExists(filepath.Join(profilesDir, settingsFiles[id])) {
			err = profileSettings.LoadYamlSettings(filepath.Join(profilesDir, settingsFiles[id]))
			if err != nil {
				Logging.CaptureException(err)
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		profileSettings.SettingsFilename = settingsFiles[id]
		// Generic load of all registered controls
		engine.LoadFromSettings(&profileSettings)
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
			// Initiale Memory-Berechnungen für Balken, damit VRAM sofort korrekt erscheint
			AIModel := Hardwareinfo.ProfileAIModelOption{}
			if controls.STTType != nil && controls.STTType.GetSelected() != nil {
				AIModel = Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", AIModelType: controls.STTType.GetSelected().Value}
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.TxtType != nil && controls.TxtType.GetSelected() != nil {
				AIModel = Hardwareinfo.ProfileAIModelOption{AIModel: "TxtTranslator", AIModelType: controls.TxtType.GetSelected().Value}
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.TTSType != nil && controls.TTSType.GetSelected() != nil {
				AIModel = Hardwareinfo.ProfileAIModelOption{AIModel: "ttsType", AIModelType: controls.TTSType.GetSelected().Value, Precision: Hardwareinfo.Float32}
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
			if controls.OCRType != nil && controls.OCRType.GetSelected() != nil {
				AIModel = Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", AIModelType: controls.OCRType.GetSelected().Value}
				AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			}
		}

		formSubmitFunction = func() {
			loadingDialog := dialog.NewCustomWithoutButtons(lang.L("Loading..."), widget.NewProgressBarInfinite(), fyne.CurrentApp().Driver().AllWindows()[1])
			fyne.Do(func() {
				loadingDialog.Show()
			})
			defer fyne.Do(func() {
				loadingDialog.Hide()
			})

			// Generic save of all registered controls
			engine.SaveToSettings(&profileSettings)

			// update existing settings or create new one if it does not exist yet
			if Utilities.FileExists(filepath.Join(profilesDir, settingsFiles[id])) {
				profileSettings.WriteYamlSettings(filepath.Join(profilesDir, settingsFiles[id]))
			} else {
				newProfileEntry := Profiles.Profile{
					SettingsFilename: settingsFiles[id],
					Websocket_ip:     profileSettings.Websocket_ip,
					Websocket_port:   profileSettings.Websocket_port,
					Run_Backend:      profileSettings.Run_backend,

					Audio_api:           profileSettings.Audio_api,
					Device_index:        profileSettings.Device_index,
					Audio_input_device:  profileSettings.Audio_input_device,
					Device_out_index:    profileSettings.Device_out_index,
					Audio_output_device: profileSettings.Audio_output_device,

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
				newProfileEntry.Save(filepath.Join(profilesDir, settingsFiles[id]))
			}
			Settings.Config = profileSettings

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
