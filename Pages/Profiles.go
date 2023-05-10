package Pages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"github.com/youpy/go-wav"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Profiles"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
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

func (c *CurrentPlaybackDevice) InitDevices() error {
	byteReader, testAudioReader := c.InitTestAudio()

	if c.device != nil && c.device.IsStarted() {
		c.device.Uninit()
	}

	captureDevices, err := c.Context.Devices(malgo.Capture)
	if err != nil {
		fmt.Println(err)
	}

	selectedCaptureDeviceIndex := -1
	for index, deviceInfo := range captureDevices {
		if deviceInfo.Name() == c.InputDeviceName {
			selectedCaptureDeviceIndex = index
			fmt.Println("Found input device: ", deviceInfo.Name(), " at index: ", index)
			break
		}
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
			// read audio bytes while reading bytes
			readBytes, _ := io.ReadFull(testAudioReader, pOutputSample)
			if readBytes <= 0 {
				c.playTestAudio = false
				byteReader.Seek(0, io.SeekStart)
				testAudioReader = wav.NewReader(byteReader)
			}
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

func (c *CurrentPlaybackDevice) Init() {
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
		os.Exit(1)
	}
	defer func() {
		if c.Context != nil {
			_ = c.Context.Uninit()
			c.Context.Free()
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
			}
			return
		}
	}

}

func GetAudioDevices(audioApi malgo.Backend, deviceType malgo.DeviceType, deviceIndexStartPoint int) ([]CustomWidget.TextValueOption, error) {

	deviceList, _ := Utilities.GetAudioDevices(audioApi, deviceType, deviceIndexStartPoint)

	if deviceList == nil {
		return nil, fmt.Errorf("no devices found")
	}

	devices := make([]CustomWidget.TextValueOption, 0)
	for _, device := range deviceList {
		devices = append(devices, CustomWidget.TextValueOption{
			Text:  device.Name,
			Value: strconv.Itoa(device.Index + 1),
		})
	}

	devices = append([]CustomWidget.TextValueOption{{
		Text:  "Default",
		Value: "-1",
	}}, devices...)

	return devices, nil
}

func appendWidgetToForm(form *widget.Form, text string, itemWidget fyne.CanvasObject, hintText string) {
	item := &widget.FormItem{Text: text, Widget: itemWidget, HintText: hintText}
	form.AppendItem(item)
}

func stopAndClose(playBackDevice CurrentPlaybackDevice, onClose func()) {
	// Pause a bit until the server is closed
	time.Sleep(1 * time.Second)

	// Closes profile window, stop audio device, and call onClose
	playBackDevice.Stop()
	onClose()
}

type ProfileAIModelOption struct {
	AIModel           string
	AIModelType       string
	AIModelSize       string
	Precision         int
	Device            string
	MemoryConsumption float64
}

var AllProfileAIModelOptions = make([]ProfileAIModelOption, 0)

func (p ProfileAIModelOption) CalculateMemoryConsumption(CPUbar *widget.ProgressBar, GPUBar *widget.ProgressBar) {
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
		if strings.ToLower(profileAIModelOption.Device) == "cuda" {
			println("CUDA MEMORY:")
			println(int(profileAIModelOption.MemoryConsumption))
			GPUBar.Value = GPUBar.Value + profileAIModelOption.MemoryConsumption
		} else if strings.ToLower(profileAIModelOption.Device) == "cpu" {
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
	playBackDevice := CurrentPlaybackDevice{}

	playBackDevice.AudioAPI = malgo.BackendWinmm
	go playBackDevice.Init()

	audioInputDevices, _ := GetAudioDevices(playBackDevice.AudioAPI, malgo.Capture, 0)
	audioOutputDevices, _ := GetAudioDevices(playBackDevice.AudioAPI, malgo.Playback, len(audioInputDevices))

	audioInputSelect := CustomWidget.NewTextValueSelect("device_index", audioInputDevices,
		func(s CustomWidget.TextValueOption) {
			println(s.Value)
			playBackDevice.InputDeviceName = s.Text
			err := playBackDevice.InitDevices()
			if err != nil {
				var newError = fmt.Errorf("audio Input (mic): %v", err)
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		},
		0)

	audioOutputSelect := CustomWidget.NewTextValueSelect("device_out_index", audioOutputDevices,
		func(s CustomWidget.TextValueOption) {
			println(s.Value)
			playBackDevice.OutputDeviceName = s.Text
			err := playBackDevice.InitDevices()
			if err != nil {
				var newError = fmt.Errorf("audio Output (speaker): %v", err)
				dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		},
		0)

	audioApiSelect := CustomWidget.NewTextValueSelect("audio_api",
		[]CustomWidget.TextValueOption{
			{Text: "MME", Value: "MME"},
			{Text: "DirectSound", Value: "DirectSound"},
			{Text: "WASAPI", Value: "WASAPI"},
		},
		func(s CustomWidget.TextValueOption) {
			var value malgo.Backend = malgo.BackendWinmm
			switch s.Value {
			case "MME":
				value = malgo.BackendWinmm
			case "DirectSound":
				value = malgo.BackendDsound
			case "WASAPI":
				value = malgo.BackendWasapi
			}
			if playBackDevice.AudioAPI != value && playBackDevice.AudioAPI != malgo.BackendNull {
				playBackDevice.Stop()
				time.Sleep(1 * time.Second)
				playBackDevice.AudioAPI = value

				go playBackDevice.Init()

				audioInputDevices, _ = GetAudioDevices(playBackDevice.AudioAPI, malgo.Capture, 0)
				audioOutputDevices, _ = GetAudioDevices(playBackDevice.AudioAPI, malgo.Playback, len(audioInputDevices))

				audioInputSelect.Options = audioInputDevices
				audioOutputSelect.Options = audioOutputDevices
				audioInputSelect.SetSelectedIndex(0)
				audioOutputSelect.SetSelectedIndex(0)
				audioInputSelect.Refresh()
				audioOutputSelect.Refresh()
			}
		},
		2)

	// show memory usage
	CPUMemoryBar := widget.NewProgressBar()
	totalCPUMemory := Hardwareinfo.GetCPUMemory()
	CPUMemoryBar.Max = float64(totalCPUMemory)
	CPUMemoryBar.TextFormatter = func() string {
		return "Estimated CPU RAM Usage: " + strconv.Itoa(int(CPUMemoryBar.Value)) + " / " + strconv.Itoa(int(CPUMemoryBar.Max)) + " MiB"
	}

	GPUMemoryBar := widget.NewProgressBar()
	totalGPUMemory := int64(0)
	if Hardwareinfo.HasNVIDIACard() {
		_, totalGPUMemory = Hardwareinfo.GetGPUMemory()
		GPUMemoryBar.Max = float64(totalGPUMemory)
	}
	GPUMemoryBar.TextFormatter = func() string {
		return "Estimated Video-RAM Usage: " + strconv.Itoa(int(GPUMemoryBar.Value)) + " / " + strconv.Itoa(int(GPUMemoryBar.Max)) + " MiB"
	}

	BuildProfileForm := func() fyne.CanvasObject {
		profileForm := widget.NewForm()
		websocketIp := widget.NewEntry()
		websocketIp.SetText("127.0.0.1")
		websocketPort := widget.NewEntry()
		websocketPort.SetText("5000")

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("Test", theme.MediaPlayIcon(), func() {
			playBackDevice.PlayStopTestAudio()
		}), playBackDevice.OutputWaveWidget)

		runBackendCheckbox := widget.NewCheck("Run Backend", func(b bool) {
			if !b {
				dialog.ShowInformation("Info", "The backend will not be started. You will have to start it manually. Without it, the UI will have no function.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
		})

		appendWidgetToForm(profileForm, "Websocket IP + Port", container.NewGridWithColumns(3, websocketIp, websocketPort, runBackendCheckbox), "IP + Port of the websocket server the backend will start and the UI will connect to.")
		profileForm.Append("", layout.NewSpacer())

		appendWidgetToForm(profileForm, "Audio API", audioApiSelect, "")

		appendWidgetToForm(profileForm, "Audio Input (mic)", audioInputSelect, "")

		profileForm.Append("", audioInputProgress)

		appendWidgetToForm(profileForm, "Audio Output (speaker)", audioOutputSelect, "")

		profileForm.Append("", audioOutputProgress)

		vadConfidenceSliderState := widget.NewLabel("0.0")
		vadConfidenceSliderWidget := widget.NewSlider(0, 1)
		vadConfidenceSliderWidget.Step = 0.1
		vadConfidenceSliderWidget.OnChanged = func(value float64) {
			vadConfidenceSliderState.SetText(fmt.Sprintf("%.1f", value))
		}

		vadOnFullClipCheckbox := widget.NewCheck("+ Check on Full Clip", func(b bool) {})
		vadOnFullClipCheckbox.Hide() // hide for now as it does not seem very useful
		vadRealtimeCheckbox := widget.NewCheck("Realtime", func(b bool) {})
		vadEnableCheckbox := widget.NewCheck("Enable", func(b bool) {
			if b {
				vadConfidenceSliderWidget.Show()
				// vadOnFullClipCheckbox.Show()
				vadRealtimeCheckbox.Show()
			} else {
				vadConfidenceSliderWidget.Hide()
				vadOnFullClipCheckbox.Hide()
				vadRealtimeCheckbox.Hide()
			}
		})
		profileForm.Append("VAD (Voice activity detection)", container.NewGridWithColumns(3, vadEnableCheckbox, vadOnFullClipCheckbox, vadRealtimeCheckbox))
		appendWidgetToForm(profileForm, "VAD Speech confidence", container.NewBorder(nil, nil, nil, vadConfidenceSliderState, vadConfidenceSliderWidget), "The confidence level required to detect speech.")

		energySliderWidget := widget.NewSlider(0, EnergySliderMax)

		// energy autodetect
		autoDetectEnergyDialog := dialog.NewCustomConfirm("This will detect the current noise level.", "Detect noise level now.", "Cancel",
			container.NewVBox(widget.NewLabel("This will record for "+strconv.Itoa(energyDetectionTime)+" seconds and sets the energy to the max detected level.\nPlease behave normally (breathing etc.) but don't say anything.\n\nThis value can later be fine-tuned without restarting by setting the \"energy\" value in Advanced -> Settings.")), func(b bool) {
				if b {
					statusBar := widget.NewProgressBarInfinite()
					statusBarContainer := container.NewVBox(statusBar)
					statusBarContainer.Add(widget.NewLabel("Please behave normally (breathing etc.) but don't say anything for around " + strconv.Itoa(energyDetectionTime) + " seconds to have it record only your noise level."))
					detectDialog := dialog.NewCustom("detecting...", "Hide", statusBarContainer, fyne.CurrentApp().Driver().AllWindows()[1])
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
						dialog.ShowInformation("Error", "Could not find audioWhisper.py or audioWhisper.exe", fyne.CurrentApp().Driver().AllWindows()[1])
						return
					}
					cmd.SysProcAttr = &syscall.SysProcAttr{
						HideWindow: true,
					}
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
						dialog.ShowInformation("Error", "Could not find detected_energy in output.", fyne.CurrentApp().Driver().AllWindows()[1])
					}
					detectDialog.Hide()
				}
			}, fyne.CurrentApp().Driver().AllWindows()[1])
		energyHelpBtn := widget.NewButtonWithIcon("Autodetect", theme.SearchIcon(), func() {
			autoDetectEnergyDialog.Show()
		})
		energySliderState := widget.NewLabel("0.0")
		energySliderWidget.OnChanged = func(value float64) {
			if value >= energySliderWidget.Max {
				energySliderWidget.Max += 10
			}
			energySliderState.SetText(fmt.Sprintf("%.0f", value))
		}
		appendWidgetToForm(profileForm, "Speech volume Level", container.NewBorder(nil, nil, nil, container.NewHBox(energySliderState, energyHelpBtn), energySliderWidget), "The volume level at which the speech detection will trigger.")

		pauseSliderState := widget.NewLabel("0.0")
		pauseSliderWidget := widget.NewSlider(0, 5)
		pauseSliderWidget.Step = 0.1
		pauseSliderWidget.OnChanged = func(value float64) {
			pauseSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, "Speech pause detection", container.NewBorder(nil, nil, nil, pauseSliderState, pauseSliderWidget), "The pause time in seconds after which the speech detection will stop and A.I. processing starts.")

		phraseLimitSliderState := widget.NewLabel("0.0")
		phraseLimitSliderWidget := widget.NewSlider(0, 30)
		phraseLimitSliderWidget.Step = 0.1
		phraseLimitSliderWidget.OnChanged = func(value float64) {
			phraseLimitSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, "Phrase time limit", container.NewBorder(nil, nil, nil, phraseLimitSliderState, phraseLimitSliderWidget), "The max. time limit in seconds after which the audio processing starts.")

		sttAiDeviceSelect := CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttPrecisionSelect := CustomWidget.NewTextValueSelect("Precision", []CustomWidget.TextValueOption{
			{Text: "float32 precision", Value: "float32"},
			{Text: "float16 precision", Value: "float16"},
			{Text: "int16 precision", Value: "int16"},
			{Text: "int8_float16 precision", Value: "int8_float16"},
			{Text: "int8 precision", Value: "int8"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		sttAiDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
				dialog.ShowInformation("No NVIDIA Card found", "No NVIDIA Card found. You might need to use CPU instead for it to work.", fyne.CurrentApp().Driver().AllWindows()[1])
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
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
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
			}
			if sttAiDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "float16" || s.Value == "int8_float16") {
				dialog.ShowInformation("Information", "Most CPU's do not support float16 computation. Please consider switching to some other precision.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if sttAiDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "int16") {
				dialog.ShowInformation("Information", "Most CUDA GPU's do not support int16 computation. Please consider switching to some other precision.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:   "Whisper",
				Precision: precisionType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}

		profileForm.Append("A.I. Device for Speech to Text", sttAiDeviceSelect)

		sttModelSize := CustomWidget.NewTextValueSelect("model", []CustomWidget.TextValueOption{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large (Defaults to Version 2)", Value: "large-v2"},
			{Text: "Large Version 1", Value: "large-v1"},
			{Text: "Large Version 2", Value: "large-v2"},
		}, func(s CustomWidget.TextValueOption) {
			sizeName, _ := strings.CutSuffix(s.Value, ".en")
			sizeName, _ = strings.CutSuffix(sizeName, "-v1")
			sizeName, _ = strings.CutSuffix(sizeName, "-v2")
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "Whisper",
				AIModelSize: sizeName,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}, 0)

		profileForm.Append("Speech to Text Size", container.NewGridWithColumns(2, sttModelSize, sttPrecisionSelect))

		sttFasterWhisperCheckbox := widget.NewCheck("Faster Whisper", func(b bool) {
			selectedPrecision := sttPrecisionSelect.GetSelected().Value
			AIModelType := ""
			if b {
				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 precision", Value: "float32"},
					{Text: "float16 precision", Value: "float16"},
					{Text: "int16 precision", Value: "int16"},
					{Text: "int8_float16 precision", Value: "int8_float16"},
					{Text: "int8 precision", Value: "int8"},
				}
				AIModelType = "CT2"
			} else {
				sttPrecisionSelect.Options = []CustomWidget.TextValueOption{
					{Text: "float32 precision", Value: "float32"},
					{Text: "float16 precision", Value: "float16"},
				}
				if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" {
					sttPrecisionSelect.SetSelected("float16")
				}
				AIModelType = "O"
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "Whisper",
				AIModelType: AIModelType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		})
		profileForm.Append("Speech to Text Options", container.NewGridWithColumns(1, sttFasterWhisperCheckbox))

		profileForm.Append("", layout.NewSpacer())

		txtTranslatorDeviceSelect := CustomWidget.NewTextValueSelect("txt_translator_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		txtTranslatorPrecisionSelect := CustomWidget.NewTextValueSelect("txt_translator_precision", []CustomWidget.TextValueOption{
			{Text: "float32 precision", Value: "float32"},
			{Text: "float16 precision", Value: "float16"},
			{Text: "int16 precision", Value: "int16"},
			{Text: "int8_float16 precision", Value: "int8_float16"},
			{Text: "int8 precision", Value: "int8"},
		}, func(s CustomWidget.TextValueOption) {}, 0)

		txtTranslatorDeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
				dialog.ShowInformation("No NVIDIA Card found", "No NVIDIA Card found. You might need to use CPU instead for it to work.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if s.Value == "cpu" && (txtTranslatorPrecisionSelect.GetSelected().Value == "float16" || txtTranslatorPrecisionSelect.GetSelected().Value == "int8_float16") {
				txtTranslatorPrecisionSelect.SetSelected("float32")
			}
			if s.Value == "cuda" && txtTranslatorPrecisionSelect.GetSelected().Value == "int16" {
				txtTranslatorPrecisionSelect.SetSelected("float16")
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "NLLB200",
				AIModelType: "CT2",
				Device:      s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
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
			}
			if txtTranslatorDeviceSelect.GetSelected().Value == "cpu" && (s.Value == "float16" || s.Value == "int8_float16") {
				dialog.ShowInformation("Information", "Most CPU's do not support float16 computation. Please consider switching to some other precision.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if txtTranslatorDeviceSelect.GetSelected().Value == "cuda" && (s.Value == "int16") {
				dialog.ShowInformation("Information", "Most CUDA GPU's do not support int16 computation. Please consider switching to some other precision.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "NLLB200",
				AIModelType: "CT2",
				Precision:   precisionType,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}

		profileForm.Append("A.I. Device for Text Translation", txtTranslatorDeviceSelect)

		txtTranslatorSizeSelect := CustomWidget.NewTextValueSelect("txt_translator_size", []CustomWidget.TextValueOption{
			{Text: "Small", Value: "small"},
			{Text: "Medium", Value: "medium"},
			{Text: "Large", Value: "large"},
		}, func(s CustomWidget.TextValueOption) {
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:     "NLLB200",
				AIModelType: "CT2",
				AIModelSize: s.Value,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}, 0)

		profileForm.Append("Text Translation Size", container.NewGridWithColumns(2, txtTranslatorSizeSelect, txtTranslatorPrecisionSelect))

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append("Text to Speech Enable", widget.NewCheck("", func(b bool) {
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
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}))

		profileForm.Append("A.I. Device for Text to Speech", CustomWidget.NewTextValueSelect("tts_ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {
			if !Hardwareinfo.HasNVIDIACard() && s.Value == "cuda" {
				dialog.ShowInformation("No NVIDIA Card found", "No NVIDIA Card found. You might need to use CPU instead for it to work.", fyne.CurrentApp().Driver().AllWindows()[1])
			}
			// calculate memory consumption
			AIModel := ProfileAIModelOption{
				AIModel:   "Silero",
				Device:    s.Value,
				Precision: Hardwareinfo.Float32,
			}
			AIModel.CalculateMemoryConsumption(CPUMemoryBar, GPUMemoryBar)
		}, 0))
		return profileForm
	}

	profileListContent := container.NewVScroll(BuildProfileForm())
	profileListContent.Hide()

	profileHelpTextContent := container.NewVScroll(widget.NewLabel("Select an existing Profile or create a new one.\n\nClick Save and Load Profile."))

	// build profile list
	var settingsFiles []string
	files, err := os.ReadDir("./")
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
		profileHelpTextContent.Hide()
		profileListContent.Show()

		profileSettings := Settings.Conf{
			SettingsFilename:         settingsFiles[id],
			Websocket_ip:             "127.0.0.1",
			Websocket_port:           5000,
			Run_backend:              true,
			Device_index:             -1,
			Device_out_index:         -1,
			Audio_api:                "WASAPI",
			Audio_input_device:       "",
			Audio_output_device:      "",
			Ai_device:                "cpu",
			Model:                    "tiny",
			Txt_translator_size:      "small",
			Txt_translator_device:    "cpu",
			Txt_translator_precision: "float32",
			Txt_translate_realtime:   false,
			Tts_enabled:              true,
			Tts_ai_device:            "cpu",
			Current_language:         "",
			Osc_ip:                   "127.0.0.1",
			Osc_port:                 9000,
			Logprob_threshold:        "-1.0",
			No_speech_threshold:      "0.6",

			Vad_enabled:              true,
			Vad_on_full_clip:         false,
			Vad_confidence_threshold: "0.4",
			Vad_num_samples:          3000,
			Vad_thread_num:           1,

			Speaker_change_check:            false,
			Speaker_similarity_threshold:    0.7,
			Speaker_diarization_window_size: 15,
			Speaker_min_duration:            0.5,

			Whisper_precision:             "float32",
			Faster_whisper:                true,
			Temperature_fallback:          true,
			Phrase_time_limit:             30.0,
			Pause:                         1.0,
			Energy:                        300,
			Beam_size:                     5,
			Whisper_cpu_threads:           0,
			Whisper_num_workers:           1,
			Realtime:                      false,
			Realtime_frame_multiply:       15,
			Realtime_frequency_time:       1.0,
			Realtime_whisper_model:        "",
			Realtime_whisper_precision:    "float32",
			Realtime_whisper_beam_size:    1,
			Realtime_temperature_fallback: false,
		}
		if Utilities.FileExists(settingsFiles[id]) {
			err = profileSettings.LoadYamlSettings(settingsFiles[id])
			if err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
			}
		}
		profileSettings.SettingsFilename = settingsFiles[id]
		profileForm := profileListContent.Content.(*widget.Form)
		profileForm.SubmitText = "Save and Load Profile"
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
			for i := 0; i < len(audioInputDevices); i++ {
				if audioInputDevices[i].Value != deviceInValue && audioInputDevices[i].Text == profileSettings.Audio_input_device {
					deviceInValue = audioInputDevices[i].Value
					break
				}
			}
		}
		if deviceInWidget.GetSelected().Value != deviceInValue {
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
			for i := 0; i < len(audioOutputDevices); i++ {
				if audioOutputDevices[i].Value != deviceOutValue && audioOutputDevices[i].Text == profileSettings.Audio_output_device {
					deviceOutValue = audioOutputDevices[i].Value
					break
				}
			}
		}
		if deviceOutWidget.GetSelected().Value != deviceOutValue {
			deviceOutWidget.SetSelected(deviceOutValue)
		}

		// audio progressbar
		// spacer
		profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(profileSettings.Vad_enabled)
		profileForm.Items[7].Widget.(*fyne.Container).Objects[1].(*widget.Check).SetChecked(profileSettings.Vad_on_full_clip)
		profileForm.Items[7].Widget.(*fyne.Container).Objects[2].(*widget.Check).SetChecked(profileSettings.Realtime)

		VadConfidenceThreshold, _ := strconv.ParseFloat(profileSettings.Vad_confidence_threshold, 64)

		profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(VadConfidenceThreshold)
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
		profileForm.Items[10].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Pause))
		profileForm.Items[11].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Phrase_time_limit))

		if profileSettings.Ai_device != nil {
			profileForm.Items[12].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Ai_device.(string))
		}
		profileForm.Items[13].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Model)
		profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Whisper_precision)
		// show only available precision options depending on whisper project
		selectedPrecision := profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value
		if profileSettings.Faster_whisper {
			profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{
				{Text: "float32 precision", Value: "float32"},
				{Text: "float16 precision", Value: "float16"},
				{Text: "int16 precision", Value: "int16"},
				{Text: "int8_float16 precision", Value: "int8_float16"},
				{Text: "int8 precision", Value: "int8"},
			}
		} else {
			profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{
				{Text: "float32 precision", Value: "float32"},
				{Text: "float16 precision", Value: "float16"},
			}
			if selectedPrecision == "int8_float16" || selectedPrecision == "int8" || selectedPrecision == "int16" {
				profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected("float16")
			}
		}
		profileForm.Items[14].Widget.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(profileSettings.Faster_whisper)

		// spacer
		profileForm.Items[16].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_device)
		profileForm.Items[17].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_size)
		profileForm.Items[17].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_precision)

		profileForm.Items[19].Widget.(*widget.Check).SetChecked(profileSettings.Tts_enabled)
		profileForm.Items[20].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Tts_ai_device)

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
			profileSettings.Vad_confidence_threshold = fmt.Sprintf("%f", profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value)

			profileSettings.Energy = int(profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value)
			profileSettings.Pause = profileForm.Items[10].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value
			profileSettings.Phrase_time_limit = profileForm.Items[11].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value

			profileSettings.Ai_device = profileForm.Items[12].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Model = profileForm.Items[13].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Whisper_precision = profileForm.Items[13].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Faster_whisper = profileForm.Items[14].Widget.(*fyne.Container).Objects[0].(*widget.Check).Checked

			profileSettings.Txt_translator_device = profileForm.Items[16].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_size = profileForm.Items[17].Widget.(*fyne.Container).Objects[0].(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_precision = profileForm.Items[17].Widget.(*fyne.Container).Objects[1].(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Tts_enabled = profileForm.Items[19].Widget.(*widget.Check).Checked
			profileSettings.Tts_ai_device = profileForm.Items[20].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			// update existing settings or create new one if it does not exist yet
			if Utilities.FileExists(settingsFiles[id]) {
				profileSettings.WriteYamlSettings(settingsFiles[id])
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
					Faster_whisper:    profileSettings.Faster_whisper,

					Txt_translator_device:    profileSettings.Txt_translator_device,
					Txt_translator_size:      profileSettings.Txt_translator_size,
					Txt_translator_precision: profileSettings.Txt_translator_precision,

					Tts_enabled:   profileSettings.Tts_enabled,
					Tts_ai_device: profileSettings.Tts_ai_device,

					Osc_ip:   profileSettings.Osc_ip,
					Osc_port: profileSettings.Osc_port,
				}
				newProfileEntry.Save(settingsFiles[id])
			}
			Settings.Config = profileSettings

			statusBar := widget.NewProgressBarInfinite()
			backendCheckStateContainer := container.NewVBox()
			backendCheckStateDialog := dialog.NewCustom(
				"",
				"Hide",
				container.NewBorder(statusBar, nil, nil, nil, backendCheckStateContainer),
				fyne.CurrentApp().Driver().AllWindows()[1],
			)
			backendCheckStateContainer.Add(widget.NewLabel("Checking backend state..."))
			backendCheckStateDialog.Show()

			// check if websocket port is in use
			websocketAddr := profileSettings.Websocket_ip + ":" + strconv.Itoa(profileSettings.Websocket_port)
			if Utilities.CheckPortInUse(websocketAddr) && profileSettings.Run_backend {
				backendCheckStateDialog.Hide()
				dialog.ShowConfirm("Websocket Port in use", "The Websocket Port is already in use. Do you want to quit the running backend?", func(b bool) {
					if b {
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
					}
				}, fyne.CurrentApp().Driver().AllWindows()[1])
			} else {
				backendCheckStateDialog.Hide()
				stopAndClose(playBackDevice, onClose)
			}
		}

		profileForm.Refresh()

		err = playBackDevice.InitDevices()
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		}
	}

	newProfileEntry := widget.NewEntry()
	newProfileEntry.PlaceHolder = "New Profile Name"
	newProfileEntry.Validator = func(s string) error {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			return fmt.Errorf("please enter a profile name")
		}
		if strings.HasSuffix(s, ".yaml") || strings.HasSuffix(s, ".yml") {
			return fmt.Errorf("please do not include file extension")
		}
		return nil
	}

	newProfileRow := container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), func() {
		validationError := newProfileEntry.Validate()
		if validationError != nil {
			return
		}
		newEntryName := newProfileEntry.Text
		newEntryName = strings.TrimSpace(newEntryName) + ".yaml"

		settingsFiles = append(settingsFiles, newEntryName)
		profileList.Select(len(settingsFiles) - 1)
		profileList.Refresh()
	}), newProfileEntry)

	memoryArea := container.NewVBox(
		CPUMemoryBar,
		GPUMemoryBar,
	)

	mainContent := container.NewHSplit(
		container.NewMax(profileHelpTextContent, profileListContent),
		container.NewBorder(newProfileRow, memoryArea, nil, nil, profileList),
	)
	mainContent.SetOffset(0.6)

	return mainContent
}
