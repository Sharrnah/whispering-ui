package Pages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"github.com/youpy/go-wav"
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
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Pages/ProfileSettings"
	"whispering-tiger-ui/Profiles"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Utilities/AudioAPI"
	"whispering-tiger-ui/Utilities/Hardwareinfo"
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
		os.Exit(1)
	}

	c.testAudioChannels = uint32(testAudioFormat.NumChannels)
	c.testAudioSampleRate = testAudioFormat.SampleRate

	return byteReader, testAudioReader
}

func (c *CurrentPlaybackDevice) InitDevices(isPlayback bool) error {
	defer Utilities.PanicLogger()

	byteReader, testAudioReader := c.InitTestAudio()

	if c.device != nil && c.device.IsStarted() {
		if c.device != nil {
			c.device.Uninit()
		}
	}

	// wait in a loop until c.Context is not nil before trying to initialize
	for {
		if c.Context != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if c.Context == nil {
		time.Sleep(1 * time.Second)
		if c.Context == nil {
			c.Init()
			time.Sleep(1 * time.Second)
		}
	}

	captureDevices, err := c.Context.Devices(malgo.Capture)
	if err != nil {
		fmt.Println(err)
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
	//deviceConfig.SampleRate = 44100
	deviceConfig.SampleRate = c.testAudioSampleRate
	deviceConfig.Alsa.NoMMap = 1

	sizeInBytesCapture := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	sizeInBytesPlayback := uint32(malgo.SampleSizeInBytes(deviceConfig.Playback.Format))

	c.InputWaveWidget.Max = 0.1
	c.InputWaveWidget.Refresh()

	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		sampleCountCapture := framecount * deviceConfig.Capture.Channels * sizeInBytesCapture
		sampleCountPlayback := framecount * deviceConfig.Playback.Channels * sizeInBytesCapture

		// play test audio
		if c.playTestAudio {
			go func() {
				// read audio bytes while reading bytes
				readBytes, _ := io.ReadFull(testAudioReader, pOutputSample)
				if readBytes <= 0 {
					c.playTestAudio = false
					byteReader.Seek(0, io.SeekStart)
					testAudioReader = wav.NewReader(byteReader)
				}
			}()
		} else {
			byteReader.Seek(0, io.SeekStart)
			testAudioReader = wav.NewReader(byteReader)
		}

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
				if c.InputWaveWidget.Max < currentVolume {
					c.InputWaveWidget.Max = currentVolume * 2
					c.InputWaveWidget.Refresh()
				}
				c.InputWaveWidget.SetValue(currentVolume)
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
				c.OutputWaveWidget.SetValue(currentVolume)
			}
		}

		/*sampleCountCapture := framecount * deviceConfig.Capture.Channels * sizeInBytesCapture

		newCapturedSampleCount := capturedSampleCount + sampleCountCapture

		pCapturedSamples = append(pCapturedSamples, pInputSamples...)

		capturedSampleCount = newCapturedSampleCount*/

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
	if c.device != nil {
		c.device.Uninit()
		c.device = nil
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
	defer Utilities.PanicLogger()

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

	// run as long as no stop signal is received
	c.stopChannel = make(chan bool)
	for {
		select {
		case <-c.stopChannel:
			fmt.Println("stopping...")
			if c.device != nil {
				c.device.Uninit()
				c.device = nil
			}
			return
		}
	}
}

func GetAudioDevices(audioApi malgo.Backend, deviceTypes []malgo.DeviceType, deviceIndexStartPoint int, specialValueSuffix string, specialTextSuffix string) ([]CustomWidget.TextValueOption, []Utilities.AudioDevice, error) {
	defer Utilities.PanicLogger()

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

	if deviceList == nil || len(deviceList) == 0 {
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
		audioInputDevicesOptions, audioInputDevices, _ := GetAudioDevices(backendItem.Backend, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "#|"+backendItem.Id+",input", " - API: "+backendItem.Name)
		audioOutputDevicesOptions, audioOutputDevices, _ := GetAudioDevices(backendItem.Backend, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "#|"+backendItem.Id+",output", " - API: "+backendItem.Name)

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

func stopAndClose(playBackDevice CurrentPlaybackDevice, onClose func()) {
	defer Utilities.PanicLogger()

	// Pause a bit until the server is closed
	time.Sleep(1 * time.Second)

	// Closes profile window, stop audio device, and call onClose
	playBackDevice.Stop()
	time.Sleep(500 * time.Millisecond) // wait for device to stop (hopefully fixes a crash when closing the profile window)
	onClose()
}

type ProfileAIModelOption struct {
	AIModel           string
	AIModelType       string
	AIModelSize       string
	Precision         float64
	Device            string
	MemoryConsumption float64
}

var AllProfileAIModelOptions = make([]ProfileAIModelOption, 0)

func (p ProfileAIModelOption) CalculateMemoryConsumption(CPUbar *widget.ProgressBar, GPUBar *widget.ProgressBar, totalGPUMemory int64) {
	addToList := true
	lastIndex := -1
	for index, profileAIModelOption := range AllProfileAIModelOptions {
		if profileAIModelOption.AIModel == p.AIModel {
			// update existing entry
			println("Device updated...")
			if p.Device != "" {
				AllProfileAIModelOptions[index].Device = p.Device
			}
			if p.AIModelType != "" {
				AllProfileAIModelOptions[index].AIModelType = p.AIModelType
			}
			if p.AIModelSize != "" {
				AllProfileAIModelOptions[index].AIModelSize = p.AIModelSize
			}
			if p.Precision != 0 {
				AllProfileAIModelOptions[index].Precision = p.Precision
			}
			AllProfileAIModelOptions[index].MemoryConsumption = p.MemoryConsumption
			addToList = false
			lastIndex = index
			break
		}
	}
	if lastIndex > -1 && len(AllProfileAIModelOptions) >= lastIndex+1 {
		// iterate through all Hardwareinfo.Models structs and find the one that matches the current Name
		for _, model := range Hardwareinfo.Models {
			fullModelName := AllProfileAIModelOptions[lastIndex].AIModel + AllProfileAIModelOptions[lastIndex].AIModelType + "_" + AllProfileAIModelOptions[lastIndex].AIModelSize
			if model.Name == fullModelName {
				finalMemoryUsage := Hardwareinfo.EstimateMemoryUsage(model.Float32PrecisionMemoryUsage, AllProfileAIModelOptions[lastIndex].Precision)
				println("FullName:")
				println(model.Name)
				println("finalMemoryUsage:")
				println(int(finalMemoryUsage))

				AllProfileAIModelOptions[lastIndex].MemoryConsumption = finalMemoryUsage
			}
		}
	}

	if addToList {
		println("Device added...")
		AllProfileAIModelOptions = append(AllProfileAIModelOptions, p)
	}

	// update memory usage bars
	GPUBar.Value = 0.0
	CPUbar.Value = 0.0
	for _, profileAIModelOption := range AllProfileAIModelOptions {
		println(profileAIModelOption.AIModel, profileAIModelOption.MemoryConsumption)
		if strings.HasPrefix(strings.ToLower(profileAIModelOption.Device), "cuda") || strings.HasPrefix(strings.ToLower(profileAIModelOption.Device), "direct-ml") {
			println("CUDA MEMORY:")
			println(int(profileAIModelOption.MemoryConsumption))
			if totalGPUMemory == 0 {
				GPUBar.Max = GPUBar.Value + profileAIModelOption.MemoryConsumption
			}
			GPUBar.Value = GPUBar.Value + profileAIModelOption.MemoryConsumption
		} else if strings.HasPrefix(strings.ToLower(profileAIModelOption.Device), "cpu") {
			println("CPU MEMORY:")
			println(int(profileAIModelOption.MemoryConsumption))
			CPUbar.Value = CPUbar.Value + profileAIModelOption.MemoryConsumption
		}
	}
	CPUbar.Refresh()
	GPUBar.Refresh()
}

const energyDetectionTime = 10
const EnergySliderMax = 2000

func CreateProfileWindow(onClose func()) fyne.CanvasObject {
	defer Utilities.PanicLogger()

	createProfilePresetSelect := CustomWidget.NewTextValueSelect("Profile Preset", []CustomWidget.TextValueOption{
		{Text: lang.L("(Select Preset)"), Value: ""},
		{Text: lang.L("NVIDIA, High Performance, Accuracy optimized"), Value: "NVIDIA-HighPerformance-Accuracy"},
		{Text: lang.L("NVIDIA, Low Performance, Accuracy optimized"), Value: "NVIDIA-LowPerformance-Accuracy"},
		{Text: lang.L("NVIDIA, High Performance, Realtime optimized"), Value: "NVIDIA-HighPerformance-Realtime"},
		{Text: lang.L("NVIDIA, Low Performance, Realtime optimized"), Value: "NVIDIA-LowPerformance-Realtime"},
		{Text: lang.L("AMD / Intel, High Performance, Accuracy optimized"), Value: "AMDIntel-HighPerformance-Accuracy"},
		{Text: lang.L("AMD / Intel, Low Performance, Accuracy optimized"), Value: "AMDIntel-LowPerformance-Accuracy"},
		{Text: lang.L("AMD / Intel, High Performance, Realtime optimized"), Value: "AMDIntel-HighPerformance-Realtime"},
		{Text: lang.L("AMD / Intel, Low Performance, Realtime optimized"), Value: "AMDIntel-LowPerformance-Realtime"},
		{Text: lang.L("CPU, High Performance, Accuracy optimized"), Value: "CPU-HighPerformance-Accuracy"},
		{Text: lang.L("CPU, Low Performance, Accuracy optimized"), Value: "CPU-LowPerformance-Accuracy"},
	}, nil, 0)

	playBackDevice := CurrentPlaybackDevice{}

	playBackDevice.AudioAPI = AudioAPI.AudioBackends[0].Backend
	go playBackDevice.Init()

	audioInputDevicesOptions, _, _ := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
	audioOutputDevicesOptions, _, _ := GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "", "")

	// fill audio device lists for later access
	fillAudioDeviceLists()

	audioInputSelect := CustomWidget.NewTextValueSelect("device_index", audioInputDevicesOptions,
		func(s CustomWidget.TextValueOption) {
			println(s.Value)
			if s.Text != "" {
				playBackDevice.InputDeviceName = s.Text
				err := playBackDevice.InitDevices(false)
				if err != nil {
					var newError = fmt.Errorf("audio Input (mic): %v", err)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			} else {
				var newError = fmt.Errorf("audio Input (mic): No device selected")
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		},
		0)

	audioOutputSelect := CustomWidget.NewTextValueSelect("device_out_index", audioOutputDevicesOptions,
		func(s CustomWidget.TextValueOption) {
			println(s.Value)
			if s.Text != "" {
				playBackDevice.OutputDeviceName = s.Text
				err := playBackDevice.InitDevices(false)
				if err != nil {
					var newError = fmt.Errorf("audio Output (speaker): %v", err)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			} else {
				var newError = fmt.Errorf("audio Output (speaker): No device selected")
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		},
		0)

	var audioOptions []CustomWidget.TextValueOption
	for _, backend := range AudioAPI.AudioBackends {
		audioOptions = append(audioOptions, CustomWidget.TextValueOption{
			Text:  backend.Name,
			Value: backend.Name,
		})
	}

	audioApiSelect := CustomWidget.NewTextValueSelect("audio_api",
		audioOptions,
		func(s CustomWidget.TextValueOption) {
			var value malgo.Backend = AudioAPI.AudioBackends[0].Backend
			value = AudioAPI.GetAudioBackendByName(s.Value).Backend
			if playBackDevice.AudioAPI != value && playBackDevice.AudioAPI != malgo.BackendNull {
				oldAudioInputSelection := audioInputSelect.GetSelected()
				oldAudioOutputSelection := audioOutputSelect.GetSelected()

				playBackDevice.Stop()
				time.Sleep(1 * time.Second)
				playBackDevice.AudioAPI = value

				audioInputDevicesOptions, _, _ = GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Capture, malgo.Loopback}, 0, "", "")
				audioOutputDevicesOptions, _, _ = GetAudioDevices(playBackDevice.AudioAPI, []malgo.DeviceType{malgo.Playback}, len(audioInputDevicesOptions), "", "")

				go playBackDevice.Init()

				playBackDevice.WaitUntilInitialized(5)

				audioInputSelect.Options = audioInputDevicesOptions
				if audioInputSelect.ContainsEntry(oldAudioInputSelection, CustomWidget.CompareText) {
					audioInputSelect.SetSelectedByText(oldAudioInputSelection.Text)
				} else {
					audioInputSelect.SetSelectedIndex(0)
				}
				audioOutputSelect.Options = audioOutputDevicesOptions
				if audioOutputSelect.ContainsEntry(oldAudioOutputSelection, CustomWidget.CompareText) {
					audioOutputSelect.SetSelectedByText(oldAudioOutputSelection.Text)
				} else {
					audioOutputSelect.SetSelectedIndex(0)
				}
			}
		},
		2)

	// show memory usage
	CPUMemoryBar := widget.NewProgressBar()
	totalCPUMemory := Hardwareinfo.GetCPUMemory()
	CPUMemoryBar.Max = float64(totalCPUMemory)
	CPUMemoryBar.TextFormatter = func() string {
		return lang.L("Estimated CPU RAM Usage:") + " " + strconv.Itoa(int(CPUMemoryBar.Value)) + " / " + strconv.Itoa(int(CPUMemoryBar.Max)) + " MiB"
	}

	GPUMemoryBar := widget.NewProgressBar()
	totalGPUMemory := int64(0)
	var ComputeCapability float32 = 0.0
	if Hardwareinfo.HasNVIDIACard() {
		_, totalGPUMemory = Hardwareinfo.GetGPUMemory()
		GPUMemoryBar.Max = float64(totalGPUMemory)

		ComputeCapability = Hardwareinfo.GetGPUComputeCapability()
	} else {
		_, totalGPUMemory = Hardwareinfo.GetWinGPUMemory("")
		GPUMemoryBar.Max = float64(totalGPUMemory)
	}

	GPUMemoryBar.TextFormatter = func() string {
		if totalGPUMemory == 0 {
			return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " MiB"
		}
		return lang.L("Estimated Video-RAM Usage:") + " " + strconv.Itoa(int(GPUMemoryBar.Value)) + " / " + strconv.Itoa(int(GPUMemoryBar.Max)) + " MiB"
	}
	GPUInformationLabel := widget.NewLabel("Compute Capability: " + fmt.Sprintf("%.1f", ComputeCapability))

	isLoadingSettingsFile := false

	BuildProfileForm := func() fyne.CanvasObject {
		profileForm := widget.NewForm()
		websocketIp := widget.NewEntry()
		websocketIp.SetText("127.0.0.1")
		websocketPort := widget.NewEntry()
		websocketPort.SetText("5000")

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("Test"), theme.MediaPlayIcon(), func() {
			playBackDevice.PlayStopTestAudio()
		}), playBackDevice.OutputWaveWidget)

		runBackendCheckbox := widget.NewCheck(lang.L("Run Backend"), func(b bool) {
			if !b {
				dialog.ShowInformation(lang.L("Information"), lang.L("The backend will not be started. You will have to start it manually or remotely. Without it, the UI will have no function."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
		})

		appendWidgetToForm(profileForm, lang.L("Websocket IP + Port"), container.NewGridWithColumns(3, websocketIp, websocketPort, runBackendCheckbox), lang.L("IP + Port of the websocket server the backend will start and the UI will connect to."))
		profileForm.Append("", layout.NewSpacer())

		appendWidgetToForm(profileForm, lang.L("Audio API"), audioApiSelect, "")

		appendWidgetToForm(profileForm, lang.L("Audio Input (mic)"), audioInputSelect, "")

		profileForm.Append("", audioInputProgress)

		appendWidgetToForm(profileForm, lang.L("Audio Output (speaker)"), audioOutputSelect, "")

		profileForm.Append("", audioOutputProgress)

		vadConfidenceSliderState := widget.NewLabel("0.00")
		vadConfidenceSliderWidget := widget.NewSlider(0, 1)
		vadConfidenceSliderWidget.Step = 0.01
		vadConfidenceSliderWidget.OnChanged = func(value float64) {
			vadConfidenceSliderState.SetText(fmt.Sprintf("%.2f", value))
		}

		vadOnFullClipCheckbox := widget.NewCheck("+ Check on Full Clip", func(b bool) {})
		vadOnFullClipCheckbox.Hide() // hide for now as it does not seem very useful

		vadRealtimeCheckbox := widget.NewCheck(lang.L("Realtime"), func(b bool) {})

		PushToTalkInput := CustomWidget.NewHotKeyEntry()
		PushToTalkInput.PlaceHolder = lang.L("Keypress")

		pushToTalkBlock := container.NewBorder(nil, nil, container.NewHBox(widget.NewLabel(lang.L("Push to Talk")), widget.NewIcon(theme.ComputerIcon())), nil, PushToTalkInput)

		vadEnableCheckbox := widget.NewCheck(lang.L("Enable"), func(b bool) {
			if b {
				vadConfidenceSliderWidget.Show()
				// vadOnFullClipCheckbox.Show()
				vadRealtimeCheckbox.Show()
				pushToTalkBlock.Show()
			} else {
				vadConfidenceSliderWidget.Hide()
				vadOnFullClipCheckbox.Hide()
				vadRealtimeCheckbox.Hide()
				pushToTalkBlock.Hide()
				if audioApiSelect.Selected != "MME" {
					dialog.ShowInformation(lang.L("Information"), lang.L("Disabled VAD is only supported with MME Audio API. Please make sure MME is selected as audio API. (Enabling VAD is highly recommended)"), fyne.CurrentApp().Driver().AllWindows()[1])
				}
			}
		})

		appendWidgetToForm(profileForm, lang.L("VAD (Voice activity detection)"), container.NewGridWithColumns(3, vadEnableCheckbox, vadOnFullClipCheckbox, vadRealtimeCheckbox, pushToTalkBlock), lang.L("Press ESC in Push to Talk field to clear the keybinding."))
		appendWidgetToForm(profileForm, lang.L("vad_confidence_threshold.Name"), container.NewBorder(nil, nil, nil, vadConfidenceSliderState, vadConfidenceSliderWidget), lang.L("The confidence level required to detect speech."))

		energySliderWidget := widget.NewSlider(0, EnergySliderMax)

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
					cmdArguments := []string{"--audio_api", audioApiSelect.GetSelected().Value, "--device_index", audioInputSelect.GetSelected().Value, "--audio_input_device", audioInputSelect.GetSelected().Text, "--detect_energy", "--detect_energy_time", strconv.Itoa(energyDetectionTime)}
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
		energySliderWidgetZeroValueInfo := dialog.NewError(fmt.Errorf(lang.L("You did set Speech volume level to 0 and have no PushToTalk Button set.This would prevent the app from recording anything.")), fyne.CurrentApp().Driver().AllWindows()[1])
		energySliderWidget.OnChanged = func(value float64) {
			if value >= energySliderWidget.Max {
				energySliderWidget.Max += 10
			}
			energySliderState.SetText(fmt.Sprintf("%.0f", value))

			if PushToTalkInput.Text == "" && value == 0 {
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

		profileForm.Append(lang.L("denoise_audio.Name"), denoiseSelect)

		pauseSliderState := widget.NewLabel("0.0")
		pauseSliderWidget := widget.NewSlider(0, 5)
		pauseSliderWidget.Step = 0.1
		pauseSliderWidgetZeroValueInfo := dialog.NewError(fmt.Errorf(lang.L("You did set Speech pause detection to 0 and have no PushToTalk Button set.This would prevent the app from stopping recording automatically.")), fyne.CurrentApp().Driver().AllWindows()[1])
		pauseSliderWidget.OnChanged = func(value float64) {
			pauseSliderState.SetText(fmt.Sprintf("%.1f", value))

			if PushToTalkInput.Text == "" && value == 0 {
				pauseSliderWidget.SetValue(0.5)
				pauseSliderWidgetZeroValueInfo.Show()
			}
		}
		appendWidgetToForm(profileForm, lang.L("pause.Name"), container.NewBorder(nil, nil, nil, pauseSliderState, pauseSliderWidget), lang.L("pause.Description"))

		phraseLimitSliderState := widget.NewLabel("0.0")
		phraseLimitSliderWidget := widget.NewSlider(0, 30)
		phraseLimitSliderWidget.Step = 0.1
		phraseLimitSliderWidget.OnChanged = func(value float64) {
			phraseLimitSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, lang.L("phrase_time_limit.Name"), container.NewBorder(nil, nil, nil, phraseLimitSliderState, phraseLimitSliderWidget), lang.L("phrase_time_limit.Description"))

		txtTranslatorSizeSelect := CustomWidget.NewTextValueSelect("txt_translator_size", []CustomWidget.TextValueOption{
			{Text: "Small", Value: "small"},
			{Text: "Medium", Value: "medium"},
			{Text: "Large", Value: "large"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		txtTranslatorPrecisionSelect := CustomWidget.NewTextValueSelect("txt_translator_precision", []CustomWidget.TextValueOption{
			{Text: "float32 " + lang.L("precision"), Value: "float32"},
			{Text: "float16 " + lang.L("precision"), Value: "float16"},
			{Text: "int16 " + lang.L("precision"), Value: "int16"},
			{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
			{Text: "int8 " + lang.L("precision"), Value: "int8"},
			{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
			{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		txtTranslatorTypeSelect := CustomWidget.NewTextValueSelect("txt_translator", []CustomWidget.TextValueOption{
			{Text: "Faster NLLB200 (200 languages)", Value: "NLLB200_CT2"},
			{Text: "Original NLLB200 (200 languages)", Value: "NLLB200"},
			{Text: "M2M100 (100 languages)", Value: "M2M100"},
			{Text: "Seamless M4T (101 languages)", Value: "Seamless_M4T"},
			{Text: lang.L("Disabled"), Value: ""},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		txtTranslatorDeviceSelect := CustomWidget.NewTextValueSelect("txt_translator_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
			{Text: "DIRECT-ML - Device 0", Value: "direct-ml:0"},
			{Text: "DIRECT-ML - Device 1", Value: "direct-ml:1"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttAiDeviceSelect := CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
			{Text: "DIRECT-ML - Device 0", Value: "direct-ml:0"},
			{Text: "DIRECT-ML - Device 1", Value: "direct-ml:1"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttPrecisionSelect := CustomWidget.NewTextValueSelect("Precision", []CustomWidget.TextValueOption{
			{Text: "float32 " + lang.L("precision"), Value: "float32"},
			{Text: "float16 " + lang.L("precision"), Value: "float16"},
			{Text: "int16 " + lang.L("precision"), Value: "int16"},
			{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
			{Text: "int8 " + lang.L("precision"), Value: "int8"},
			{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
			{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
			{Text: "8bit " + lang.L("precision"), Value: "8bit"},
			{Text: "4bit " + lang.L("precision"), Value: "4bit"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttTypeSelect := CustomWidget.NewTextValueSelect("stt_type", []CustomWidget.TextValueOption{
			{Text: "Faster Whisper", Value: "faster_whisper"},
			{Text: "Original Whisper", Value: "original_whisper"},
			{Text: "Transformer Whisper", Value: "transformer_whisper"},
			//{Text: "Medusa Whisper", Value: "medusa_whisper"},
			//{Text: "TensorRT Whisper", Value: "tensorrt_whisper"},
			//{Text: "Whisper CPP", Value: "whisper_cpp"},
			{Text: "Seamless M4T", Value: "seamless_m4t"},
			{Text: "MMS", Value: "mms"},
			{Text: "Speech T5 (English only)", Value: "speech_t5"},
			{Text: "Wav2Vec Bert 2.0", Value: "wav2vec_bert"},
			{Text: "NeMo Canary", Value: "nemo_canary"},
			{Text: lang.L("Disabled"), Value: ""},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttAiDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if s.Value == "cpu" && (sttPrecisionSelect.GetSelected().Value == "float16" || sttPrecisionSelect.GetSelected().Value == "int8_float16") {
				sttPrecisionSelect.SetSelected("float32")
			}
			if s.Value == "cuda" && sttPrecisionSelect.GetSelected().Value == "int16" {
				sttPrecisionSelect.SetSelected("float16")
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel: "Whisper",
				Device:  s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)

			/**
			special case for Seamless M4T since its a multi-modal model and does not need additional memory when used for Text translation and Speech-to-text
			*/
			if txtTranslatorTypeSelect.GetSelected().Value == "Seamless_M4T" && sttTypeSelect.GetSelected().Value == "seamless_m4t" {
				txtTranslatorSizeSelect.SetSelected(s.Value)
				txtTranslatorPrecisionSelect.SetSelected(sttPrecisionSelect.GetSelected().Value)
				txtTranslatorDeviceSelect.SetSelected(sttAiDeviceSelect.GetSelected().Value)
				txtTranslatorSizeSelect.Disable()
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
			} else if txtTranslatorTypeSelect.GetSelected().Value != "" {
				txtTranslatorSizeSelect.Enable()
				txtTranslatorPrecisionSelect.Enable()
				txtTranslatorDeviceSelect.Enable()
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
			if sttAiDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "float16" || s.Value == "int8_float16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "float16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if sttAiDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "bfloat16" || s.Value == "int8_bfloat16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "bfloat16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if sttAiDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "int16") {
				dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU's", "Precision": "int16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if sttAiDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "bfloat16" || s.Value == "int8_bfloat16") && ComputeCapability < 8.0 {
				dialog.ShowInformation(lang.L("Information"), lang.L("Your Device most likely does not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU", "Precision": "bfloat16"}), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:   "Whisper",
				Precision: precisionType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)

			/**
			special case for Seamless M4T since its a multi-modal model and does not need additional memory when used for Text translation and Speech-to-text
			*/
			if txtTranslatorTypeSelect.GetSelected().Value == "Seamless_M4T" && sttTypeSelect.GetSelected().Value == "seamless_m4t" {
				txtTranslatorSizeSelect.SetSelected(s.Value)
				txtTranslatorPrecisionSelect.SetSelected(sttPrecisionSelect.GetSelected().Value)
				txtTranslatorDeviceSelect.SetSelected(sttAiDeviceSelect.GetSelected().Value)
				txtTranslatorSizeSelect.Disable()
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
			} else if txtTranslatorTypeSelect.GetSelected().Value != "" {
				txtTranslatorSizeSelect.Enable()
				txtTranslatorPrecisionSelect.Enable()
				txtTranslatorDeviceSelect.Enable()
			}
		}

		originalWhisperModelList := []CustomWidget.TextValueOption{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large V1", Value: "large-v1"},
			{Text: "Large V2", Value: "large-v2"},
			{Text: "Large V3", Value: "large-v3"},
		}

		medusaWhisperModelList := []CustomWidget.TextValueOption{
			{Text: "V1", Value: "v1"},
		}

		fasterWhisperModelList := []CustomWidget.TextValueOption{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			//{Text: "Large (Defaults to V3)", Value: "large-v3"},
			{Text: "Large V1", Value: "large-v1"},
			{Text: "Large V2", Value: "large-v2"},
			{Text: "Large V3", Value: "large-v3"},
			{Text: "Medium Distilled (English)", Value: "medium-distilled.en"},
			{Text: "Large V2 Distilled (English)", Value: "large-distilled-v2.en"},
			{Text: "Large V3 Distilled (English)", Value: "large-distilled-v3.en"},
			{Text: "Small (European finetune)", Value: "small.eu"},
			{Text: "Medium (European finetune)", Value: "medium.eu"},
			{Text: "Small (German finetune)", Value: "small.de"},
			{Text: "Medium (German finetune)", Value: "medium.de"},
			{Text: "Large V2 (German finetune)", Value: "large-v2.de2"},
			{Text: "Large V3 Distilled (German finetune)", Value: "large-distilled-v3.de"},
			{Text: "Small (German-Swiss finetune)", Value: "small.de-swiss"},
			{Text: "Medium (Mix-Japanese-v2 finetune)", Value: "medium.mix-jpv2"},
			{Text: "Large V2 (Mix-Japanese finetune)", Value: "large-v2.mix-jp"},
			{Text: "Small (Japanese finetune)", Value: "small.jp"},
			{Text: "Medium (Japanese finetune)", Value: "medium.jp"},
			{Text: "Large V2 (Japanese finetune)", Value: "large-v2.jp"},
			{Text: "Medium (Korean finetune)", Value: "medium.ko"},
			{Text: "Large V2 (Korean finetune)", Value: "large-v2.ko"},
			{Text: "Small (Chinese finetune)", Value: "small.zh"},
			{Text: "Medium (Chinese finetune)", Value: "medium.zh"},
			{Text: "Large V2 (Chinese finetune)", Value: "large-v2.zh"},
		}

		originalSeamlessM4TModelList := []CustomWidget.TextValueOption{
			{Text: "Medium", Value: "medium"},
			{Text: "Large", Value: "large"},
			{Text: "Large V2", Value: "large-v2"},
		}

		originalMmsModelList := []CustomWidget.TextValueOption{
			{Text: "1b-fl102 (102 languages)", Value: "mms-1b-fl102"},
			{Text: "1b-l1107 (1107 languages)", Value: "mms-1b-l1107"},
			{Text: "1b-all (1162 languages)", Value: "1b-all"},
		}

		sttModelSize := CustomWidget.NewTextValueSelect("model", fasterWhisperModelList, func(s CustomWidget.TextValueOption) {
			// remove last suffix starting with a dot
			sizeName := strings.Split(s.Value, ".")[0]
			sizeName, _ = strings.CutSuffix(sizeName, "-v1")
			sizeName, _ = strings.CutSuffix(sizeName, "-v2")
			sizeName, _ = strings.CutSuffix(sizeName, "-v3")

			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "Whisper",
				AIModelSize: sizeName,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)

			/**
			special case for Seamless M4T since its a multi-modal model and does not need additional memory when used for Text translation and Speech-to-text
			*/
			if txtTranslatorTypeSelect.GetSelected().Value == "Seamless_M4T" && sttTypeSelect.GetSelected().Value == "seamless_m4t" {
				txtTranslatorSizeSelect.SetSelected(s.Value)
				txtTranslatorPrecisionSelect.SetSelected(sttPrecisionSelect.GetSelected().Value)
				txtTranslatorDeviceSelect.SetSelected(sttAiDeviceSelect.GetSelected().Value)
				txtTranslatorSizeSelect.Disable()
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
			} else if txtTranslatorTypeSelect.GetSelected().Value != "" {
				txtTranslatorSizeSelect.Enable()
				txtTranslatorPrecisionSelect.Enable()
				txtTranslatorDeviceSelect.Enable()
			}
		}, 0)

		sttTypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			sttPrecisionSelectOption := sttPrecisionSelect.GetSelected()
			selectedPrecision := ""
			if sttPrecisionSelectOption != nil {
				selectedPrecision = sttPrecisionSelect.GetSelected().Value
			}
			AIModelType := ""
			sttPrecisionSelect.Enable()
			sttModelSize.Enable()
			sttAiDeviceSelect.Enable()

			selectedModelSizeOption := sttModelSize.GetSelected()

			sttAiDeviceSelect.Options = []CustomWidget.TextValueOption{
				{Text: "CUDA", Value: "cuda"},
				{Text: "CPU", Value: "cpu"},
				{Text: "DIRECT-ML - Device 0", Value: "direct-ml:0"},
				{Text: "DIRECT-ML - Device 1", Value: "direct-ml:1"},
			}

			if s.Value == "faster_whisper" {
				sttModelSize.Options = fasterWhisperModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(0)
				}

				sttAiDeviceSelect.Options = []CustomWidget.TextValueOption{
					{Text: "CUDA", Value: "cuda"},
					{Text: "CPU", Value: "cpu"},
				}

				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "int16 " + lang.L("precision"), Value: "int16"},
					{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
					{Text: "int8 " + lang.L("precision"), Value: "int8"},
					{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
					{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
				}
				AIModelType = "CT2"
			} else if s.Value == "original_whisper" {
				sttModelSize.Options = originalWhisperModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(0)
				}

				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "O"
			} else if s.Value == "transformer_whisper" {
				sttModelSize.Options = originalWhisperModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(0)
				}

				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "8bit " + lang.L("precision"), Value: "8bit"},
					{Text: "4bit " + lang.L("precision"), Value: "4bit"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "O"
				//} else if s.Value == "tensorrt_whisper" {
				//	sttModelSize.Options = originalWhisperModelList
				//	// unselect if not in list
				//	if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
				//		sttModelSize.SetSelectedIndex(0)
				//	}
				//
				//	sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
				//		{Text: "float32 " + lang.L("precision"), Value: "float32"},
				//		{Text: "float16 " + lang.L("precision"), Value: "float16"},
				//		{Text: "8bit " + lang.L("precision"), Value: "8bit"},
				//		{Text: "4bit " + lang.L("precision"), Value: "4bit"},
				//	}
				//	if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
				//		sttPrecisionSelect.SetSelected("float16")
				//	}
				//	AIModelType = "O"
			} else if s.Value == "medusa_whisper" {
				sttModelSize.Options = medusaWhisperModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(0)
				}

				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "O"
			} else if s.Value == "seamless_m4t" {
				sttModelSize.Options = originalSeamlessM4TModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(1)
				}

				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
					{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
					{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
				}
				if selectedPrecision == "int8" || selectedPrecision == "int16" {
					sttPrecisionSelect.SetSelected("float32")
				}
				//sttPrecisionSelect.Disable()
				AIModelType = "m4t"

				if txtTranslatorTypeSelect.GetSelected().Value != "Seamless_M4T" && !isLoadingSettingsFile {
					dialog.NewConfirm(lang.L("Usage of Multi-Modal Model."), lang.L("Use Multi-Modal model for Text-Translation as well?"), func(b bool) {
						if b {
							txtTranslatorTypeSelect.SetSelected("Seamless_M4T")
						}
					}, fyne.CurrentApp().Driver().AllWindows()[1]).Show()
				}
			} else if s.Value == "speech_t5" {
				sttPrecisionSelect.Disable()
				sttModelSize.Disable()
				AIModelType = "t5"
			} else if s.Value == "wav2vec_bert" {
				sttModelSize.Disable()
				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "8bit " + lang.L("precision"), Value: "8bit"},
					{Text: "4bit " + lang.L("precision"), Value: "4bit"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "wav2vec-bert"
			} else if s.Value == "mms" {
				sttModelSize.Options = originalMmsModelList
				// unselect if not in list
				if selectedModelSizeOption == nil || !sttModelSize.ContainsEntry(selectedModelSizeOption, CustomWidget.CompareValue) {
					sttModelSize.SetSelectedIndex(1)
				}
				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "8bit " + lang.L("precision"), Value: "8bit"},
					{Text: "4bit " + lang.L("precision"), Value: "4bit"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" || selectedPrecision == "bfloat16" || selectedPrecision == "int8_bfloat16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "mms"
			} else if s.Value == "nemo_canary" {
				sttPrecisionSelect.Disable()
				sttPrecisionSelect.SetSelected("float32") // only available in float32
				sttModelSize.Disable()
				AIModelType = "nemo-canary"
			} else {
				sttPrecisionSelect.Disable()
				sttModelSize.Disable()
				sttAiDeviceSelect.Disable()
				AIModelType = "disabled"
			}

			sttAiDeviceSelect.Refresh()

			/**
			special case for Seamless M4T since its a multi-modal model and does not need additional memory when used for Text translation and Speech-to-text
			*/
			if txtTranslatorTypeSelect.GetSelected().Value == "Seamless_M4T" && s.Value == "seamless_m4t" {
				if txtTranslatorSizeSelect.ContainsEntry(sttModelSize.GetSelected(), CustomWidget.CompareValue) {
					txtTranslatorSizeSelect.SetSelected(sttModelSize.GetSelected().Value)
				}
				txtTranslatorPrecisionSelect.SetSelected(sttPrecisionSelect.GetSelected().Value)
				txtTranslatorDeviceSelect.SetSelected(sttAiDeviceSelect.GetSelected().Value)
				txtTranslatorSizeSelect.Disable()
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
			} else if txtTranslatorTypeSelect.GetSelected().Value != "" {
				txtTranslatorSizeSelect.Enable()
				txtTranslatorPrecisionSelect.Enable()
				txtTranslatorDeviceSelect.Enable()
			}

			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "Whisper",
				AIModelType: AIModelType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)

		}

		//denoiseCheckbox := widget.NewCheck("A.I. Denoise", func(b bool) {})
		profileForm.Append(lang.L("Speech-to-Text Type"), container.NewGridWithColumns(2, sttTypeSelect))

		profileForm.Append(lang.L("A.I. Device for Speech-to-Text"), sttAiDeviceSelect)

		profileForm.Append(lang.L("Speech-to-Text A.I. Size"), container.NewGridWithColumns(2, sttModelSize, sttPrecisionSelect))

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append(lang.L("Text-Translation Type"), container.NewGridWithColumns(2, txtTranslatorTypeSelect))

		txtTranslatorDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
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
			AIModel := ProfileAIModelOption{
				AIModel: "TxtTranslator",
				Device:  s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
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
			AIModel := ProfileAIModelOption{
				AIModel:   "TxtTranslator",
				Precision: precisionType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}

		profileForm.Append(lang.L("A.I. Device for Text-Translation"), txtTranslatorDeviceSelect)

		txtTranslatorSizeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "TxtTranslator",
				AIModelSize: s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}

		txtTranslatorTypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			selectedPrecisionOption := txtTranslatorPrecisionSelect.GetSelected()
			selectedPrecision := ""
			if selectedPrecisionOption != nil {
				selectedPrecision = selectedPrecisionOption.Value
			}
			selectedSizeOption := txtTranslatorSizeSelect.GetSelected()
			selectedSize := ""
			if selectedSizeOption != nil {
				selectedSize = selectedSizeOption.Value
			}

			txtTranslatorDeviceSelect.Enable()
			txtTranslatorPrecisionSelect.Enable()
			txtTranslatorSizeSelect.Enable()

			modelType := s.Value

			txtTranslatorDeviceSelect.Options = []CustomWidget.TextValueOption{
				{Text: "CUDA", Value: "cuda"},
				{Text: "CPU", Value: "cpu"},
				{Text: "DIRECT-ML - Device 0", Value: "direct-ml:0"},
				{Text: "DIRECT-ML - Device 1", Value: "direct-ml:1"},
			}

			if s.Value == "NLLB200" {
				txtTranslatorPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" {
					txtTranslatorPrecisionSelect.SetSelected("float16")
				}
			} else if s.Value == "NLLB200_CT2" || s.Value == "M2M100" {
				txtTranslatorPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "int16 " + lang.L("precision"), Value: "int16"},
					{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
					{Text: "int8 " + lang.L("precision"), Value: "int8"},
					{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
					{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
				}
				// CTranslate2 models do not support direct-ml
				txtTranslatorDeviceSelect.Options = []CustomWidget.TextValueOption{
					{Text: "CUDA", Value: "cuda"},
					{Text: "CPU", Value: "cpu"},
				}
			} else if s.Value == "Seamless_M4T" {
				txtTranslatorPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 " + lang.L("precision"), Value: "float32"},
					{Text: "float16 " + lang.L("precision"), Value: "float16"},
					{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
					{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
					{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
				}
			} else if s.Value == "" {
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorSizeSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
				modelType = "N"
			}

			txtTranslatorDeviceSelect.Refresh()

			if s.Value == "M2M100" {
				txtTranslatorSizeSelect.Options = []CustomWidget.TextValueOption{
					{Text: "Small", Value: "small"},
					{Text: "Large", Value: "large"},
				}
				if selectedSize == "medium" {
					txtTranslatorSizeSelect.SetSelected("small")
				}
			} else if s.Value == "NLLB200_CT2" || s.Value == "NLLB200" {
				txtTranslatorSizeSelect.Options = []CustomWidget.TextValueOption{
					{Text: "Small", Value: "small"},
					{Text: "Medium", Value: "medium"},
					{Text: "Large", Value: "large"},
				}
			} else if s.Value == "Seamless_M4T" {
				txtTranslatorSizeSelect.Options = []CustomWidget.TextValueOption{
					{Text: "Medium", Value: "medium"},
					{Text: "Large", Value: "large"},
					{Text: "Large V2", Value: "large-v2"},
				}
				if selectedSize == "small" {
					txtTranslatorSizeSelect.SetSelected("medium")
				}
			}

			/**
			special case for Seamless M4T since its a multi-modal model and does not need additional memory when used for Text translation and Speech-to-text
			*/
			if s.Value == "Seamless_M4T" && sttTypeSelect.GetSelected().Value == "seamless_m4t" {
				modelType = "N"
				if txtTranslatorSizeSelect.ContainsEntry(sttModelSize.GetSelected(), CustomWidget.CompareValue) {
					txtTranslatorSizeSelect.SetSelected(sttModelSize.GetSelected().Value)
				}
				txtTranslatorPrecisionSelect.SetSelected(sttPrecisionSelect.GetSelected().Value)
				txtTranslatorDeviceSelect.SetSelected(sttAiDeviceSelect.GetSelected().Value)
				txtTranslatorSizeSelect.Disable()
				txtTranslatorPrecisionSelect.Disable()
				txtTranslatorDeviceSelect.Disable()
			}

			AIModel := ProfileAIModelOption{
				AIModel:     "TxtTranslator",
				AIModelType: modelType,
			}

			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}

		profileForm.Append(lang.L("Text-Translation A.I. Size"), container.NewGridWithColumns(2, txtTranslatorSizeSelect, txtTranslatorPrecisionSelect))

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append(lang.L("Integrated Text-to-Speech"), widget.NewCheck("", func(b bool) {
			enabledType := "N"
			if b {
				enabledType = "O"
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "Silero",
				AIModelType: enabledType,
				Precision:   Hardwareinfo.Float32,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}))

		profileForm.Append(lang.L("A.I. Device for Text-to-Speech"), CustomWidget.NewTextValueSelect("tts_ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
			{Text: "DIRECT-ML - Device 0", Value: "direct-ml:0"},
			{Text: "DIRECT-ML - Device 1", Value: "direct-ml:1"},
		}, func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:   "Silero",
				Device:    s.Value,
				Precision: Hardwareinfo.Float32,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar, totalGPUMemory)
		}, 0))

		pushToTalkChanged := false
		PushToTalkInput.OnChanged = func(s string) {
			if s != "" && !isLoadingSettingsFile {
				pushToTalkChanged = true
			}
		}
		PushToTalkInput.OnFocusChanged = func(focusGained bool) {
			if !focusGained && pushToTalkChanged && PushToTalkInput.Text != "" {
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

		return profileForm
	}

	profileListContent := container.NewVScroll(BuildProfileForm())
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

		profileSettings := ProfileSettings.Presets[createProfilePresetSelect.GetSelected().Value]
		profileSettings.SettingsFilename = settingsFiles[id]

		if Utilities.FileExists(filepath.Join(profilesDir, settingsFiles[id])) {
			err = profileSettings.LoadYamlSettings(filepath.Join(profilesDir, settingsFiles[id]))
			if err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		profileSettings.SettingsFilename = settingsFiles[id]
		profileForm := profileListContent.Content.(*widget.Form)
		profileForm.SubmitText = lang.L("Save and Load Profile")
		profileForm.Items[0].Widget.(*fyne.Container).Objects[0].(*widget.Entry).SetText(profileSettings.Websocket_ip)
		profileForm.Items[0].Widget.(*fyne.Container).Objects[1].(*widget.Entry).SetText(strconv.Itoa(profileSettings.Websocket_port))
		profileForm.Items[0].Widget.(*fyne.Container).Objects[2].(*widget.Check).SetChecked(profileSettings.Run_backend)
		// spacer
		profileForm.Items[2].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Audio_api)

		deviceInValue := "-1"
		deviceInWidget := profileForm.Items[3].Widget.(*CustomWidget.TextValueSelect)
		if profileSettings.Device_index != nil {
			switch profileSettings.Device_index.(type) {
			case int:
				deviceInValue = strconv.Itoa(profileSettings.Device_index.(int))
			case string:
				deviceInValue = profileSettings.Device_index.(string)
			}
		}
		// select audio input device by name instead of index if possible
		if profileSettings.Audio_input_device != "" && profileSettings.Audio_input_device != "Default" && deviceInValue != "-1" {
			for i := 0; i < len(audioInputDevicesOptions); i++ {
				if audioInputDevicesOptions[i].Value != deviceInValue && audioInputDevicesOptions[i].Text == profileSettings.Audio_input_device {
					deviceInValue = audioInputDevicesOptions[i].Value
					break
				}
			}
		}
		deviceInWidgetOption := deviceInWidget.GetSelected()
		if deviceInWidgetOption != nil && deviceInWidgetOption.Value != deviceInValue {
			deviceInWidget.SetSelected(deviceInValue)
		}
		// audio progressbar
		deviceOutValue := "-1"
		deviceOutWidget := profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect)
		if profileSettings.Device_out_index != nil {
			switch profileSettings.Device_out_index.(type) {
			case int:
				deviceOutValue = strconv.Itoa(profileSettings.Device_out_index.(int))
			case string:
				deviceOutValue = profileSettings.Device_out_index.(string)
			}
		}
		// select audio output device by name instead of index if possible
		if profileSettings.Audio_output_device != "" && profileSettings.Audio_output_device != "Default" && deviceOutValue != "-1" {
			for i := 0; i < len(audioOutputDevicesOptions); i++ {
				if audioOutputDevicesOptions[i].Value != deviceOutValue && audioOutputDevicesOptions[i].Text == profileSettings.Audio_output_device {
					deviceOutValue = audioOutputDevicesOptions[i].Value
					break
				}
			}
		}
		deviceOutWidgetOption := deviceOutWidget.GetSelected()
		if deviceOutWidgetOption != nil && deviceOutWidgetOption.Value != deviceOutValue {
			deviceOutWidget.SetSelected(deviceOutValue)
		}

		// audio progressbar
		// spacer
		profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(profileSettings.Vad_enabled)
		profileForm.Items[7].Widget.(*fyne.Container).Objects[1].(*widget.Check).SetChecked(profileSettings.Vad_on_full_clip)
		profileForm.Items[7].Widget.(*fyne.Container).Objects[2].(*widget.Check).SetChecked(profileSettings.Realtime)
		profileForm.Items[7].Widget.(*fyne.Container).Objects[3].(*fyne.Container).Objects[0].(*CustomWidget.HotKeyEntry).SetText(profileSettings.Push_to_talk_key)

		profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Vad_confidence_threshold))
		if profileSettings.Vad_enabled {
			profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Show()
		} else {
			profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Hide()
		}

		profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Max = EnergySliderMax
		if float64(profileSettings.Energy) >= profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Max-10 {
			profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Max = float64(profileSettings.Energy + 200)
		}
		profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Energy))
		profileForm.Items[10].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Denoise_audio) // !!!!!!!!
		profileForm.Items[11].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Pause))
		profileForm.Items[12].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Phrase_time_limit))

		profileForm.Items[13].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Stt_type)
		//profileForm.Items[12].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Denoise_audio)

		if profileSettings.Ai_device != nil {
			profileForm.Items[14].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Ai_device.(string))
		}
		profileForm.Items[15].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Model)
		profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Whisper_precision)
		// show only available precision options depending on whisper project
		selectedPrecisionOption := profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected()
		selectedPrecision := ""
		if selectedPrecisionOption != nil {
			selectedPrecision = selectedPrecisionOption.Value
		}
		if profileSettings.Stt_type == "faster_whisper" {
			profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{
				{Text: "float32 " + lang.L("precision"), Value: "float32"},
				{Text: "float16 " + lang.L("precision"), Value: "float16"},
				{Text: "int16 " + lang.L("precision"), Value: "int16"},
				{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
				{Text: "int8 " + lang.L("precision"), Value: "int8"},
				{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
				{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
			}
		} else if profileSettings.Stt_type == "original_whisper" {
			profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{
				{Text: "float32 " + lang.L("precision"), Value: "float32"},
				{Text: "float16 " + lang.L("precision"), Value: "float16"},
			}
			if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" {
				profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected("float16")
			}
		}

		// spacer (15)
		profileForm.Items[17].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator)
		profileForm.Items[18].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_device)
		profileForm.Items[19].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_size)
		profileForm.Items[19].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_precision)
		// spacer (19)
		profileForm.Items[21].Widget.(*widget.Check).SetChecked(profileSettings.Tts_enabled)
		profileForm.Items[22].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Tts_ai_device)

		profileForm.OnSubmit = func() {
			profileSettings.Websocket_ip = profileForm.Items[0].Widget.(*fyne.Container).Objects[0].(*widget.Entry).Text
			profileSettings.Websocket_port, _ = strconv.Atoi(profileForm.Items[0].Widget.(*fyne.Container).Objects[1].(*widget.Entry).Text)
			profileSettings.Run_backend = profileForm.Items[0].Widget.(*fyne.Container).Objects[2].(*widget.Check).Checked

			profileSettings.Audio_api = profileForm.Items[2].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Device_index, _ = strconv.Atoi(profileForm.Items[3].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value)
			profileSettings.Audio_input_device = profileForm.Items[3].Widget.(*CustomWidget.TextValueSelect).GetSelected().Text

			profileSettings.Device_out_index, _ = strconv.Atoi(profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value)
			profileSettings.Audio_output_device = profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect).GetSelected().Text

			profileSettings.Vad_enabled = profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Check).Checked
			profileSettings.Vad_on_full_clip = profileForm.Items[7].Widget.(*fyne.Container).Objects[1].(*widget.Check).Checked
			profileSettings.Realtime = profileForm.Items[7].Widget.(*fyne.Container).Objects[2].(*widget.Check).Checked
			profileSettings.Push_to_talk_key = profileForm.Items[7].Widget.(*fyne.Container).Objects[3].(*fyne.Container).Objects[0].(*CustomWidget.HotKeyEntry).Text
			profileSettings.Vad_confidence_threshold = profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value

			profileSettings.Energy = int(profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value)
			profileSettings.Denoise_audio = profileForm.Items[10].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value // !!!!!!
			profileSettings.Pause = profileForm.Items[11].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value
			profileSettings.Phrase_time_limit = profileForm.Items[12].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value

			profileSettings.Stt_type = profileForm.Items[13].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			//profileSettings.Denoise_audio = profileForm.Items[12].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Ai_device = profileForm.Items[14].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Model = profileForm.Items[15].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Whisper_precision = profileForm.Items[15].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Txt_translator = profileForm.Items[17].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_device = profileForm.Items[18].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_size = profileForm.Items[19].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_precision = profileForm.Items[19].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Tts_enabled = profileForm.Items[21].Widget.(*widget.Check).Checked
			profileSettings.Tts_ai_device = profileForm.Items[22].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

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

					Tts_enabled:   profileSettings.Tts_enabled,
					Tts_ai_device: profileSettings.Tts_ai_device,

					Osc_ip:   profileSettings.Osc_ip,
					Osc_port: profileSettings.Osc_port,
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
					stopAndClose(playBackDevice, onClose)
					backendCheckDialog.Hide()
				}))
				yesButton := widget.NewButtonWithIcon(lang.L("Yes"), theme.ConfirmIcon(), func() {
					err := Utilities.KillProcessById(Settings.Config.Process_id)
					if err != nil {
						err = Utilities.SendQuitMessage(websocketAddr)
					}
					if err != nil {
						fmt.Printf("Failed to send quit message: %v\n", err)
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
					} else {
						stopAndClose(playBackDevice, onClose)
					}
					backendCheckDialog.Hide()
				})
				yesButton.Importance = widget.HighImportance
				buttonList.Add(yesButton)

				backendCheckDialogContent.Add(
					widget.NewLabelWithStyle(lang.L("The Websocket Port is already in use")+"\n"+lang.L("Do you want to quit the running backend or reconnect to it?"), fyne.TextAlignCenter, fyne.TextStyle{}),
				)

				backendCheckDialogContent.Add(
					container.New(layout.NewCenterLayout(), buttonList),
				)

				backendCheckDialog.Show()
			} else {
				backendCheckStateDialog.Hide()
				stopAndClose(playBackDevice, onClose)
			}
		}

		profileForm.Refresh()

		err = playBackDevice.InitDevices(false)
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		}
		isLoadingSettingsFile = false
	}

	newProfileEntry := widget.NewEntry()
	newProfileEntry.PlaceHolder = lang.L("New Profile Name")
	newProfileEntry.Validator = func(s string) error {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			return fmt.Errorf(lang.L("please enter a profile name"))
		}
		if strings.HasSuffix(s, ".yaml") || strings.HasSuffix(s, ".yml") {
			return fmt.Errorf(lang.L("please do not include file extension"))
		}
		return nil
	}

	newProfileRow := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon(lang.L("New"), theme.DocumentCreateIcon(), func() {
		validationError := newProfileEntry.Validate()
		if validationError != nil {
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
