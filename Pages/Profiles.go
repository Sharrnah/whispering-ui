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
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/ProfileSettings"
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
}

func (c *CurrentPlaybackDevice) Stop() {
	c.stopChannel <- true
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
	defer func() {
		c.isInitializing = false
	}()

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
	c.InputWaveWidget.Refresh()

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

	// wait for a single stop signal and then clean up
	c.stopChannel = make(chan bool)
	<-c.stopChannel
	fmt.Println("stopping...")
	if c.device != nil {
		c.device.Uninit()
		c.device = nil
	}
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
		return devicesOptions, nil, fmt.Errorf("no devices found")
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

	// Audio section will be built inside BuildProfileForm after engine initialization

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
	var coord *Coordinator
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

		// Aktualisiere ggf. den Coordinator mit dem ermittelten Gesamt-GPU-RAM,
		// damit spätere Modellwechsel den Maximalwert korrekt setzen können
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
		// Zeige den Maximalwert aus der ProgressBar (wird nach GPU-Detect gesetzt)
		if GPUMemoryBar.Max <= 0 {
			return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " MiB"
		}
		return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " / " + strconv.Itoa(int(GPUMemoryBar.Max)) + " MiB"
	}

	isLoadingSettingsFile := false
	// Controls struct holds all widget references for clean load/save
	controls := &AllProfileControls{}
	// Form engine for generic load/save mapping
	var engine *FormEngine
	// Coordinator wird weiter unten initialisiert (siehe coord = &Coordinator{...})

	BuildProfileForm := func() fyne.CanvasObject {
		profileForm := widget.NewForm()
		// Form engine to centralize option updates and fallbacks
		engine = NewFormEngine(controls, nil)
		// Use builder for connection row
		builder := NewProfileBuilder()
		connectionRow := builder.BuildConnectionSection(engine)
		// Build audio section now that engine exists
		audioSection := builder.BuildAudioSection(engine, audioInputDevicesOptions, audioOutputDevicesOptions)

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("Test"), theme.MediaPlayIcon(), func() {
			playBackDevice.PlayStopTestAudio()
		}), playBackDevice.OutputWaveWidget)

		runBackendCheckbox := engine.Controls.RunBackend
		runBackendCheckbox.OnChanged = func(b bool) {
			if !b {
				dialog.ShowInformation(lang.L("Information"), lang.L("The backend will not be started. You will have to start it manually or remotely. Without it, the UI will have no function."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		engine.Register("run_backend", runBackendCheckbox)

		// Build OCR section via builder
		ocrSection := builder.BuildOCRSection(engine)
		controls.OCRDevice = ocrSection.DeviceSelect
		controls.OCRType = ocrSection.TypeSelect
		controls.OCRPrecision = ocrSection.PrecisionSelect
		// Memory calc handlers preserved
		ocrSection.DeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", Device: s.Value}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.HandleMultiModalAllSync()
			}
		}
		ocrSection.PrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16":
				precisionType = Hardwareinfo.Float16
			case "int32":
				precisionType = Hardwareinfo.Int32
			case "int16":
				precisionType = Hardwareinfo.Int16
			case "int8_float16":
				precisionType = Hardwareinfo.Int8
			case "int8":
				precisionType = Hardwareinfo.Int8
			case "bfloat16":
				precisionType = Hardwareinfo.Float16
			case "int8_bfloat16":
				precisionType = Hardwareinfo.Int8
			case "8bit":
				precisionType = Hardwareinfo.Bit8
			case "4bit":
				precisionType = Hardwareinfo.Bit4
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", Precision: precisionType}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.HandleMultiModalAllSync()
			}
		}

		appendWidgetToForm(profileForm, lang.L("Websocket IP + Port"), connectionRow, lang.L("IP + Port of the websocket server the backend will start and the UI will connect to."))
		profileForm.Append("", layout.NewSpacer())

		appendWidgetToForm(profileForm, lang.L("Audio API"), audioSection.ApiSelect, "")
		controls.AudioAPI = audioSection.ApiSelect
		engine.Register("audio_api", audioSection.ApiSelect)

		appendWidgetToForm(profileForm, lang.L("Audio Input (mic)"), audioSection.InputSelect, "")
		controls.AudioInput = audioSection.InputSelect
		engine.Register("device_index", audioSection.InputSelect)

		profileForm.Append("", audioInputProgress)

		appendWidgetToForm(profileForm, lang.L("Audio Output (speaker)"), audioSection.OutputSelect, "")
		controls.AudioOutput = audioSection.OutputSelect
		engine.Register("device_out_index", audioSection.OutputSelect)

		profileForm.Append("", audioOutputProgress)

		// Build VAD section via UI builder
		vadSection := builder.BuildVADSection(engine)
		controls.VadEnable = vadSection.EnableCheck
		controls.VadOnFullClip = vadSection.OnFullClipCheck
		controls.VadRealtime = vadSection.RealtimeCheck
		controls.PushToTalk = vadSection.PushToTalk
		controls.VadConfidence = vadSection.ConfidenceSlider

		// Wire audio input/output selection behavior
		audioSection.InputSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			println(s.Value)
			if s.Text != "" {
				playBackDevice.InputDeviceName = s.Text
				if err := playBackDevice.InitDevices(false); err != nil {
					var newError = fmt.Errorf("audio Input (mic): %v", err)
					Logging.CaptureException(newError)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			} else {
				var newError = fmt.Errorf("audio Input (mic): No device selected")
				Logging.CaptureException(newError)
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		audioSection.OutputSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			println(s.Value)
			if s.Text != "" {
				playBackDevice.OutputDeviceName = s.Text
				if err := playBackDevice.InitDevices(true); err != nil {
					var newError = fmt.Errorf("audio Output (speaker): %v", err)
					Logging.CaptureException(newError)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			} else {
				var newError = fmt.Errorf("audio Output (speaker): No device selected")
				Logging.CaptureException(newError)
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}

		audioSection.ApiSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			var value malgo.Backend = AudioAPI.AudioBackends[0].Backend
			value = AudioAPI.GetAudioBackendByName(s.Value).Backend
			if value != malgo.BackendWinmm && !controls.VadEnable.Checked && !isLoadingSettingsFile {
				dialog.ShowInformation(lang.L("Information"), lang.L("Disabled VAD is only supported with MME Audio API. Please make sure MME is selected as audio API. (Enabling VAD is highly recommended)"), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if playBackDevice.AudioAPI != value && playBackDevice.AudioAPI != malgo.BackendNull {
				oldAudioInputSelection := audioSection.InputSelect.GetSelected()
				oldAudioOutputSelection := audioSection.OutputSelect.GetSelected()

				playBackDevice.Stop()
				playBackDevice.AudioAPI = value

				audioInputDevicesOptions, _, err = GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
				if err != nil {
					Logging.CaptureException(err)
				}
				audioOutputDevicesOptions, _, err = GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "", "")
				if err != nil {
					Logging.CaptureException(err)
				}

				go playBackDevice.Init()

				playBackDevice.WaitUntilInitialized(5)

				audioSection.InputSelect.Options = audioInputDevicesOptions
				if audioSection.InputSelect.ContainsEntry(oldAudioInputSelection, CustomWidget.CompareText) {
					audioSection.InputSelect.SetSelectedByText(oldAudioInputSelection.Text)
				} else {
					engine.SetOptionsWithFallback(audioSection.InputSelect, audioInputDevicesOptions)
				}
				audioSection.OutputSelect.Options = audioOutputDevicesOptions
				if audioSection.OutputSelect.ContainsEntry(oldAudioOutputSelection, CustomWidget.CompareText) {
					audioSection.OutputSelect.SetSelectedByText(oldAudioOutputSelection.Text)
				} else {
					engine.SetOptionsWithFallback(audioSection.OutputSelect, audioOutputDevicesOptions)
				}
			}
		}

		appendWidgetToForm(profileForm, lang.L("VAD (Voice activity detection)"), vadSection.GroupRow, lang.L("Press ESC in Push to Talk field to clear the keybinding."))
		appendWidgetToForm(profileForm, lang.L("vad_confidence_threshold.Name"), vadSection.ConfidenceRow, lang.L("The confidence level required to detect speech."))

		energySliderWidget := widget.NewSlider(0, EnergySliderMax)
		controls.Energy = energySliderWidget
		engine.Register("energy", energySliderWidget)

		// energy autodetect
		autoDetectEnergyDialog := dialog.NewCustomConfirm(lang.L("This will detect the current noise level."), lang.L("Detect noise level now."), lang.L("Cancel"),
			container.NewVBox(widget.NewLabel(lang.L("This will record for energyDetectionTime seconds and sets the energy to the max detected level. Please behave normally (breathing etc.) but don't say anything. This value can later be fine-tuned without restarting by setting the energy value in Advanced-Settings.", map[string]interface{}{
				"EnergyDetectionTime": lang.N("TimeSeconds", energyDetectionTime, map[string]interface{}{"RecordingTime": energyDetectionTime}),
			}))), func(b bool) {
				if b {
					statusBar := widget.NewProgressBarInfinite()
					statusBarContainer := container.NewVBox(statusBar)
					statusBarContainer.Add(widget.NewLabel(lang.L("Please behave normally (breathing etc.) but don't say anything for around energyDetectionTime seconds to have it record only your noise level.", map[string]interface{}{
						"EnergyDetectionTime": lang.N("TimeSeconds", energyDetectionTime, map[string]interface{}{"RecordingTime": energyDetectionTime}),
					})))
					detectDialog := dialog.NewCustom(lang.L("Detecting..."), lang.L("Hide"), statusBarContainer, fyne.CurrentApp().Driver().AllWindows()[1])
					detectDialog.Show()

					cmd := exec.Command("---")
					// start application that detects the energy level and returns the value before exiting.
					cmdArguments := []string{"--audio_api", audioSection.ApiSelect.GetSelected().Value, "--device_index", audioSection.InputSelect.GetSelected().Value, "--audio_input_device", audioSection.InputSelect.GetSelected().Text, "--detect_energy", "--detect_energy_time", strconv.Itoa(energyDetectionTime)}
					if Utilities.FileExists("audioWhisper.py") {
						cmdArguments = append([]string{"-u", "audioWhisper.py"}, cmdArguments...)
						cmd = exec.Command("python", cmdArguments...)
					} else if Utilities.FileExists("audioWhisper/audioWhisper.exe") {
						cmd = exec.Command("audioWhisper/audioWhisper.exe", cmdArguments...)
					} else {
						dialog.ShowInformation(lang.L("Error"), lang.L("Could not find audioWhisper.py or audioWhisper.exe"), fyne.CurrentApp().Driver().AllWindows()[1])
						return
					}
					Utilities.ProcessHideWindowAttr(cmd)
					out, err := cmd.Output()
					if err != nil {
						Logging.CaptureException(err)
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
						return
					}
					// find and convert cmd detected energy output to float64
					re := regexp.MustCompile(`detected_energy: (\d+)`)
					matches := re.FindStringSubmatch(string(out))
					if len(matches) > 0 {
						detectedEnergy, _ := strconv.ParseFloat(matches[1], 64)
						energySliderWidget.Max = EnergySliderMax
						if detectedEnergy >= energySliderWidget.Max {
							energySliderWidget.Max = detectedEnergy + 200
						}
						energySliderWidget.SetValue(detectedEnergy + 20)
					} else {
						dialog.ShowInformation(lang.L("Error"), lang.L("Could not find detected_energy in output."), fyne.CurrentApp().Driver().AllWindows()[1])
					}
					detectDialog.Hide()

					// reinit devices after detection
					_ = playBackDevice.InitDevices(false)
				}
			}, fyne.CurrentApp().Driver().AllWindows()[1])
		energyHelpBtn := widget.NewButtonWithIcon(lang.L("Autodetect"), theme.VolumeUpIcon(), func() {
			autoDetectEnergyDialog.Show()
		})
		energySliderState := widget.NewLabel("0.0")
		energySliderWidgetZeroValueInfo := dialog.NewError(errors.New(lang.L("You did set Speech volume level to 0 and have no PushToTalk Button set.This would prevent the app from recording anything.")), fyne.CurrentApp().Driver().AllWindows()[1])
		energySliderWidget.OnChanged = func(value float64) {
			if value >= energySliderWidget.Max {
				energySliderWidget.Max += 10
			}
			energySliderState.SetText(fmt.Sprintf("%.0f", value))

			if controls.PushToTalk.Text == "" && value == 0 {
				energySliderWidget.SetValue(1)
				energySliderWidgetZeroValueInfo.Show()
			}
		}
		appendWidgetToForm(profileForm, lang.L("energy.Name"), container.NewBorder(nil, nil, nil, container.NewHBox(energySliderState, energyHelpBtn), energySliderWidget), lang.L("The volume level at which the speech detection will trigger. (0 = Disabled, useful for Push2Talk)"))

		denoiseSelect := CustomWidget.NewTextValueSelect("denoise_audio", []CustomWidget.TextValueOption{
			{Text: lang.L("Disabled"), Value: ""},
			{Text: "Noise Reduce", Value: "noise_reduce"},
			{Text: "DeepFilterNet", Value: "deepfilter"},
		}, func(s CustomWidget.TextValueOption) {}, 0)
		controls.DenoiseAudio = denoiseSelect
		engine.Register("denoise_audio", denoiseSelect)

		appendWidgetToForm(profileForm, lang.L("denoise_audio.Name"), denoiseSelect, "")

		pauseSliderState := widget.NewLabel("0.0")
		pauseSliderWidget := widget.NewSlider(0, 5)
		pauseSliderWidget.Step = 0.1
		controls.PauseSeconds = pauseSliderWidget
		engine.Register("pause", pauseSliderWidget)
		pauseSliderWidgetZeroValueInfo := dialog.NewError(errors.New(lang.L("You did set Speech pause detection to 0 and have no PushToTalk Button set.This would prevent the app from stopping recording automatically.")), fyne.CurrentApp().Driver().AllWindows()[1])
		pauseSliderWidget.OnChanged = func(value float64) {
			pauseSliderState.SetText(fmt.Sprintf("%.1f", value))

			if controls.PushToTalk.Text == "" && value == 0 {
				pauseSliderWidget.SetValue(0.5)
				pauseSliderWidgetZeroValueInfo.Show()
			}
		}
		appendWidgetToForm(profileForm, lang.L("pause.Name"), container.NewBorder(nil, nil, nil, pauseSliderState, pauseSliderWidget), lang.L("pause.Description"))

		phraseLimitSliderState := widget.NewLabel("0.0")
		phraseLimitSliderWidget := widget.NewSlider(0, 30)
		phraseLimitSliderWidget.Step = 0.1
		controls.PhraseTimeLimit = phraseLimitSliderWidget
		engine.Register("phrase_time_limit", phraseLimitSliderWidget)
		phraseLimitSliderWidget.OnChanged = func(value float64) {
			phraseLimitSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, lang.L("phrase_time_limit.Name"), container.NewBorder(nil, nil, nil, phraseLimitSliderState, phraseLimitSliderWidget), lang.L("phrase_time_limit.Description"))

		controls.VadEnable.OnChanged = func(b bool) {
			if b {
				pauseSliderWidget.Min = 0.0
				phraseLimitSliderWidget.Min = 0.0

				controls.VadConfidence.Show()
				// vadOnFullClipCheckbox.Show()
				controls.VadRealtime.Show()
				vadSection.PushToTalkBlock.Show()
			} else {
				controls.VadConfidence.Hide()
				controls.VadOnFullClip.Hide()
				controls.VadRealtime.Hide()
				vadSection.PushToTalkBlock.Hide()
				if audioSection.ApiSelect.Selected != "MME" && !isLoadingSettingsFile {
					dialog.ShowInformation(lang.L("Information"), lang.L("Disabled VAD is only supported with MME Audio API. Please make sure MME is selected as audio API. (Enabling VAD is highly recommended)"), fyne.CurrentApp().Driver().AllWindows()[1])
				}
				if pauseSliderWidget.Value == 0 || phraseLimitSliderWidget.Value == 0 && !isLoadingSettingsFile {
					dialog.ShowInformation(lang.L("Information"), lang.L("You disabled VAD but have set the pause or phrase limit to 0. This is not supported. Setting Pause and Phrase limits to non-zero values."), fyne.CurrentApp().Driver().AllWindows()[1])
					if pauseSliderWidget.Value == 0 {
						pauseSliderWidget.SetValue(1.2)
					}
					if phraseLimitSliderWidget.Value == 0 {
						phraseLimitSliderWidget.SetValue(30)
					}
				}
				// set min values for pause and phrase limit for non VAD mode
				pauseSliderWidget.Min = 0.1
				phraseLimitSliderWidget.Min = 0.1
			}
		}

		// Build STT section via builder
		sttSection := builder.BuildSTTSection(engine)
		sttTypeSelect := sttSection.TypeSelect
		sttAiDeviceSelect := sttSection.DeviceSelect
		sttPrecisionSelect := sttSection.PrecisionSelect
		sttModelSize := sttSection.SizeSelect

		appendWidgetToForm(profileForm, lang.L("Speech-to-Text Type"), container.NewGridWithColumns(2, sttTypeSelect), "")
		appendWidgetToForm(profileForm, lang.L("A.I. Device for Speech-to-Text"), sttAiDeviceSelect, "")
		appendWidgetToForm(profileForm, lang.L("Speech-to-Text A.I. Size"), container.NewGridWithColumns(2, sttModelSize, sttPrecisionSelect), "")

		// STT handlers: memory + coordinator
		sttTypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.ApplySTTTypeChange(s.Value)
			}
		}
		sttAiDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", Device: s.Value}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			// Hinweis bei potenziell inkompatiblen Kombinationen
			if coord != nil {
				prec := ""
				if sttPrecisionSelect.GetSelected() != nil {
					prec = sttPrecisionSelect.GetSelected().Value
				}
				coord.EnsurePrecisionDeviceCompatibility(s.Value, prec)
				// Multimodal sync ggf. anstoßen
				if !coord.InProgrammaticUpdate {
					coord.HandleMultiModalAllSync()
				}
			}
		}
		sttPrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16":
				precisionType = Hardwareinfo.Float16
			case "int32":
				precisionType = Hardwareinfo.Int32
			case "int16":
				precisionType = Hardwareinfo.Int16
			case "int8_float16":
				precisionType = Hardwareinfo.Int8
			case "int8":
				precisionType = Hardwareinfo.Int8
			case "bfloat16":
				precisionType = Hardwareinfo.Float16
			case "int8_bfloat16":
				precisionType = Hardwareinfo.Int8
			case "8bit":
				precisionType = Hardwareinfo.Bit8
			case "4bit":
				precisionType = Hardwareinfo.Bit4
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", Precision: precisionType}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			// Hinweis bei potenziell inkompatiblen Kombinationen
			if coord != nil && sttAiDeviceSelect.GetSelected() != nil {
				coord.EnsurePrecisionDeviceCompatibility(sttAiDeviceSelect.GetSelected().Value, s.Value)
				// Multimodal sync ggf. anstoßen
				if !coord.InProgrammaticUpdate {
					coord.HandleMultiModalAllSync()
				}
			}
		}
		sttModelSize.OnChanged = func(s CustomWidget.TextValueOption) {
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", AIModelSize: s.Value}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.HandleMultiModalAllSync()
			}
		}

		profileForm.Append("", layout.NewSpacer())

		// Build TXT section via builder
		txtSection := builder.BuildTXTSection(engine)
		txtTranslatorTypeSelect := txtSection.TypeSelect
		txtTranslatorDeviceSelect := txtSection.DeviceSelect
		txtTranslatorPrecisionSelect := txtSection.PrecisionSelect
		txtTranslatorSizeSelect := txtSection.SizeSelect

		appendWidgetToForm(profileForm, lang.L("Text-Translation Type"), container.NewGridWithColumns(2, txtTranslatorTypeSelect), "")

		txtTranslatorDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if s.Value == "cuda" && !HasNvidiaGPU && (coord == nil || !coord.InProgrammaticUpdate) {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if s.Value == "cpu" && txtTranslatorPrecisionSelect.GetSelected() != nil && (txtTranslatorPrecisionSelect.GetSelected().Value == "float16" || txtTranslatorPrecisionSelect.GetSelected().Value == "int8_float16") {
				txtTranslatorPrecisionSelect.SetSelected("float32")
			}
			if s.Value == "cpu" && txtTranslatorPrecisionSelect.GetSelected() != nil && (txtTranslatorPrecisionSelect.GetSelected().Value == "bfloat16" || txtTranslatorPrecisionSelect.GetSelected().Value == "int8_bfloat16") {
				txtTranslatorPrecisionSelect.SetSelected("float32")
			}
			if s.Value == "cuda" && txtTranslatorPrecisionSelect.GetSelected() != nil && txtTranslatorPrecisionSelect.GetSelected().Value == "int16" {
				txtTranslatorPrecisionSelect.SetSelected("float16")
			}
			// calculate memory consumption
			AIModel := Hardwareinfo.ProfileAIModelOption{
				AIModel: "TxtTranslator",
				Device:  s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil {
				prec := ""
				if txtTranslatorPrecisionSelect.GetSelected() != nil {
					prec = txtTranslatorPrecisionSelect.GetSelected().Value
				}
				coord.EnsurePrecisionDeviceCompatibility(s.Value, prec)
				if !coord.InProgrammaticUpdate {
					coord.HandleMultiModalAllSync()
				}
			}
		}

		txtTranslatorPrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16":
				precisionType = Hardwareinfo.Float16
			case "int32":
				precisionType = Hardwareinfo.Int32
			case "int16":
				precisionType = Hardwareinfo.Int16
			case "int8_float16":
				precisionType = Hardwareinfo.Int8
			case "int8":
				precisionType = Hardwareinfo.Int8
			case "bfloat16":
				precisionType = Hardwareinfo.Float16
			case "int8_bfloat16":
				precisionType = Hardwareinfo.Int8
			case "8bit":
				precisionType = Hardwareinfo.Bit8
			case "4bit":
				precisionType = Hardwareinfo.Bit4
			}
			if txtTranslatorDeviceSelect.GetSelected() != nil && txtTranslatorDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "float16" || s.Value == "int8_float16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "float16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if txtTranslatorDeviceSelect.GetSelected() != nil && txtTranslatorDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "bfloat16" || s.Value == "int8_bfloat16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "bfloat16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if txtTranslatorDeviceSelect.GetSelected() != nil && txtTranslatorDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "int16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU's", "Precision": "int16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if sttAiDeviceSelect.GetSelected() != nil && sttAiDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "bfloat16" || s.Value == "int8_bfloat16") && ComputeCapability < 8.0 {
				dialog.ShowInformation(lang.L("Information"), lang.L("Your Device most likely does not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU", "Precision": "bfloat16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := Hardwareinfo.ProfileAIModelOption{
				AIModel:   "TxtTranslator",
				Precision: precisionType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.HandleMultiModalAllSync()
			}
		}

		appendWidgetToForm(profileForm, lang.L("A.I. Device for Text-Translation"), txtTranslatorDeviceSelect, "")

		txtTranslatorSizeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			// calculate memory consumption
			AIModel := Hardwareinfo.ProfileAIModelOption{
				AIModel:     "TxtTranslator",
				AIModelSize: s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.HandleMultiModalAllSync()
			}
		}

		txtTranslatorTypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.ApplyTXTTypeChange(s.Value)
			}
		}

		appendWidgetToForm(profileForm, lang.L("Text-Translation A.I. Size"), container.NewGridWithColumns(2, txtTranslatorSizeSelect, txtTranslatorPrecisionSelect), "")

		profileForm.Append("", layout.NewSpacer())

		// Build TTS section via builder
		ttsSection := builder.BuildTTSSection(engine)
		ttsTypeSelect := ttsSection.TypeSelect
		ttsAiDeviceSelect := ttsSection.DeviceSelect
		controls.TTSDevice = ttsAiDeviceSelect
		controls.TTSType = ttsTypeSelect
		// retain memory calc/info dialogs for TTS device
		ttsAiDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if s.Value == "cuda" && !HasNvidiaGPU && (coord == nil || !coord.InProgrammaticUpdate) {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ttsType", Device: s.Value, Precision: Hardwareinfo.Float32}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}

		ttsTypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.ApplyTTSTypeChange(s.Value)
			}
		}

		appendWidgetToForm(profileForm, lang.L("Integrated Text-to-Speech"), container.NewGridWithColumns(2, ttsTypeSelect), "")

		appendWidgetToForm(profileForm, lang.L("A.I. Device for Text-to-Speech"), ttsAiDeviceSelect, "")

		profileForm.Append("", layout.NewSpacer())

		ocrSection.TypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if coord != nil && !coord.InProgrammaticUpdate {
				coord.ApplyOCRTypeChange(s.Value)
			}
		}
		appendWidgetToForm(profileForm, lang.L("Integrated Image-to-Text"), ocrSection.TypeSelect, "")
		//profileForm.Append(lang.L("Integrated Image-to-Text"), container.NewGridWithColumns(2, ocrTypeSelect))
		appendWidgetToForm(profileForm, lang.L("A.I. Device for Image-to-Text"), container.NewGridWithColumns(2, ocrSection.DeviceSelect, ocrSection.PrecisionSelect), "")
		//profileForm.Append(lang.L("A.I. Device for Image-to-Text"), ocrAiDeviceSelect)

		profileForm.Append("", layout.NewSpacer())
		pushToTalkChanged := false
		controls.PushToTalk.OnChanged = func(s string) {
			if s != "" && !isLoadingSettingsFile {
				pushToTalkChanged = true
			}
		}
		controls.PushToTalk.OnFocusChanged = func(focusGained bool) {
			if !focusGained && pushToTalkChanged && controls.PushToTalk.Text != "" {
				dialog.NewConfirm(lang.L("Change speech trigger settings?"), lang.L("You did set a PushToTalk Button. Do you want to set settings to trigger with only a Button press?"), func(b bool) {
					if b {
						energySliderWidget.SetValue(0)
						pauseSliderWidget.SetValue(0)
						phraseLimitSliderWidget.SetValue(0)
					}
				}, fyne.CurrentApp().Driver().AllWindows()[1]).Show()
				pushToTalkChanged = false
			}
		}

		// Initialize coordinator now that all relevant controls exist
		coord = &Coordinator{
			Controls: &ProfileControls{
				STTType:      sttTypeSelect,
				STTDevice:    sttAiDeviceSelect,
				STTPrecision: sttPrecisionSelect,
				STTModelSize: sttModelSize,
				TxtType:      txtTranslatorTypeSelect,
				TxtDevice:    txtTranslatorDeviceSelect,
				TxtPrecision: txtTranslatorPrecisionSelect,
				TxtSize:      txtTranslatorSizeSelect,
				OCRType:      ocrSection.TypeSelect,
				OCRDevice:    ocrSection.DeviceSelect,
				OCRPrecision: ocrSection.PrecisionSelect,
				TTSType:      ttsTypeSelect,
				TTSDevice:    ttsAiDeviceSelect,
			},
			IsLoadingSettings: &isLoadingSettingsFile,
			ComputeCapability: ComputeCapability,
			CPUMemoryBar:      CPUMemoryBar,
			GPUMemoryBar:      GPUMemoryBar,
			TotalGPUMemoryMiB: totalGPUMemory,
		}

		// Nach Initialisierung: Falls GPU-Total bereits ermittelt wurde, direkt setzen
		if totalGPUMemory > 0 {
			coord.TotalGPUMemoryMiB = totalGPUMemory
			if GPUMemoryBar.Max <= 0 {
				GPUMemoryBar.Max = float64(totalGPUMemory)
				GPUMemoryBar.Refresh()
			}
		} else {
			// Warte kurz asynchron auf GPU-Erkennung und setze dann den Max-Wert
			go func() {
				for i := 0; i < 50; i++ { // bis zu ~5s warten
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
		formSubmitFunction()
	}

	//profileListContent := container.NewVScroll(profileForm)

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
