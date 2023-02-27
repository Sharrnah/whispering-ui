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
	"strconv"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Profiles"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

type CurrentPlaybackDevice struct {
	InputDeviceName  string
	OutputDeviceName string
	InputWaveWidget  *widget.ProgressBar
	OutputWaveWidget *widget.ProgressBar
	Context          *malgo.AllocatedContext

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
	c.Context, err = malgo.InitContext([]malgo.Backend{malgo.BackendWinmm}, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = c.Context.Uninit()
		c.Context.Free()
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

func GetAudioDevices(deviceType malgo.DeviceType, deviceIndexStartPoint int) ([]CustomWidget.TextValueOption, error) {

	deviceList, _ := Utilities.GetAudioDevices(deviceType, deviceIndexStartPoint)

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

func CreateProfileWindow(onClose func()) fyne.CanvasObject {
	playBackDevice := CurrentPlaybackDevice{}

	go playBackDevice.Init()

	audioInputDevices, _ := GetAudioDevices(malgo.Capture, 0)
	audioOutputDevices, _ := GetAudioDevices(malgo.Playback, len(audioInputDevices))

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

		appendWidgetToForm(profileForm, "Websocket IP + Port", container.NewGridWithColumns(2, websocketIp, websocketPort), "IP + Port of the websocket server the backend will start and the UI will connect to.")
		profileForm.Append("", layout.NewSpacer())

		appendWidgetToForm(profileForm, "Audio Input (mic)", CustomWidget.NewTextValueSelect("device_index", audioInputDevices,
			func(s CustomWidget.TextValueOption) {
				println(s.Value)
				playBackDevice.InputDeviceName = s.Text
				err := playBackDevice.InitDevices()
				if err != nil {
					var newError = fmt.Errorf("audio Input (mic): %v", err)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			},
			0),
			"")

		profileForm.Append("", audioInputProgress)

		appendWidgetToForm(profileForm, "Audio Output (speaker)", CustomWidget.NewTextValueSelect("device_out_index", audioOutputDevices,
			func(s CustomWidget.TextValueOption) {
				println(s.Value)
				playBackDevice.OutputDeviceName = s.Text
				err := playBackDevice.InitDevices()
				if err != nil {
					var newError = fmt.Errorf("audio Output (speaker): %v", err)
					dialog.ShowError(newError, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			},
			0),
			"")

		profileForm.Append("", audioOutputProgress)

		vadConfidenceSliderState := widget.NewLabel("0.0")
		vadConfidenceSliderWidget := widget.NewSlider(0, 1)
		vadConfidenceSliderWidget.Step = 0.1
		vadConfidenceSliderWidget.OnChanged = func(value float64) {
			vadConfidenceSliderState.SetText(fmt.Sprintf("%.1f", value))
		}

		vadOnFullClipCheckbox := widget.NewCheck("Additional Check on Full Clip", func(b bool) {})
		vadEnableCheckbox := widget.NewCheck("Enable", func(b bool) {
			if b {
				vadConfidenceSliderWidget.Show()
				vadOnFullClipCheckbox.Show()
			} else {
				vadConfidenceSliderWidget.Hide()
				vadOnFullClipCheckbox.Hide()
			}
		})
		profileForm.Append("VAD (Voice activity detection)", container.NewGridWithColumns(2, vadEnableCheckbox, vadOnFullClipCheckbox))
		appendWidgetToForm(profileForm, "VAD Speech confidence", container.NewBorder(nil, nil, nil, vadConfidenceSliderState, vadConfidenceSliderWidget), "The confidence level required to detect speech.")

		energySliderState := widget.NewLabel("0.0")
		energySliderWidget := widget.NewSlider(0, 1000)
		energySliderWidget.OnChanged = func(value float64) {
			energySliderState.SetText(fmt.Sprintf("%.0f", value))
		}
		appendWidgetToForm(profileForm, "Speech detection Level", container.NewBorder(nil, nil, nil, energySliderState, energySliderWidget), "The volume level at which the speech detection will trigger.")

		pauseSliderState := widget.NewLabel("0.0")
		pauseSliderWidget := widget.NewSlider(0, 5)
		pauseSliderWidget.Step = 0.1
		pauseSliderWidget.OnChanged = func(value float64) {
			pauseSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, "Speech pause detection", container.NewBorder(nil, nil, nil, pauseSliderState, pauseSliderWidget), "The pause time in seconds after which the speech detection will stop and A.I. processing starts.")

		phraseLimitSliderState := widget.NewLabel("0.0")
		phraseLimitSliderWidget := widget.NewSlider(0, 50)
		phraseLimitSliderWidget.Step = 0.1
		phraseLimitSliderWidget.OnChanged = func(value float64) {
			phraseLimitSliderState.SetText(fmt.Sprintf("%.1f", value))
		}
		appendWidgetToForm(profileForm, "Phrase time limit", container.NewBorder(nil, nil, nil, phraseLimitSliderState, phraseLimitSliderWidget), "The max. time limit in seconds after which the audio processing starts.")

		profileForm.Append("A.I. Device for Speech to Text", CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0))

		profileForm.Append("Speech to Text Size", CustomWidget.NewTextValueSelect("model", []CustomWidget.TextValueOption{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large (Defaults to Version 2)", Value: "large"},
			{Text: "Large Version 1", Value: "large-v1"},
			{Text: "Large Version 2", Value: "large-v2"},
		}, func(s CustomWidget.TextValueOption) {}, 0))

		profileForm.Append("Speech to Text use FP16", widget.NewCheck("", func(b bool) {}))

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append("A.I. Device for Text Translation", CustomWidget.NewTextValueSelect("txt_translator_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0))

		profileForm.Append("Text Translation Size", CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
			{Text: "Small", Value: "small"},
			{Text: "Medium", Value: "medium"},
			{Text: "Large", Value: "large"},
		}, func(s CustomWidget.TextValueOption) {}, 0))

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append("Text to Speech Enable", widget.NewCheck("", func(b bool) {}))

		profileForm.Append("A.I. Device for Text to Speech", CustomWidget.NewTextValueSelect("tts_ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0))
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
			SettingsFilename:      settingsFiles[id],
			Websocket_ip:          "127.0.0.1",
			Websocket_port:        5000,
			Device_index:          -1,
			Device_out_index:      -1,
			Ai_device:             "cuda",
			Model:                 "tiny",
			Txt_translator_size:   "small",
			Txt_translator_device: "cuda",
			Tts_enabled:           true,
			Tts_ai_device:         "cuda",
			Current_language:      "",
			Osc_ip:                "127.0.0.1",
			Osc_port:              9000,
			Logprob_threshold:     "-1.0",
			No_speech_threshold:   "0.6",

			Vad_enabled:              true,
			Vad_on_full_clip:         false,
			Vad_confidence_threshold: "0.4",
			Vad_num_samples:          3000,

			Fp16:              false,
			Phrase_time_limit: 0.0,
			Pause:             0.8,
			Energy:            300,
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
		// spacer
		deviceInValue := "-1"
		deviceInWidget := profileForm.Items[2].Widget.(*CustomWidget.TextValueSelect)
		if profileSettings.Device_index != nil {
			switch profileSettings.Device_index.(type) {
			case int:
				deviceInValue = strconv.Itoa(profileSettings.Device_index.(int))
			case string:
				deviceInValue = profileSettings.Device_index.(string)
			}
		}
		if deviceInWidget.GetSelected().Value != deviceInValue {
			deviceInWidget.SetSelected(deviceInValue)
		}
		// audio progressbar
		deviceOutValue := "-1"
		deviceOutWidget := profileForm.Items[4].Widget.(*CustomWidget.TextValueSelect)
		if profileSettings.Device_out_index != nil {
			switch profileSettings.Device_out_index.(type) {
			case int:
				deviceOutValue = strconv.Itoa(profileSettings.Device_out_index.(int))
			case string:
				deviceOutValue = profileSettings.Device_out_index.(string)
			}
		}
		if deviceOutWidget.GetSelected().Value != deviceOutValue {
			deviceOutWidget.SetSelected(deviceOutValue)
		}

		// audio progressbar
		// spacer
		profileForm.Items[6].Widget.(*fyne.Container).Objects[0].(*widget.Check).SetChecked(profileSettings.Vad_enabled)
		profileForm.Items[6].Widget.(*fyne.Container).Objects[1].(*widget.Check).SetChecked(profileSettings.Vad_on_full_clip)

		VadConfidenceThreshold, _ := strconv.ParseFloat(profileSettings.Vad_confidence_threshold, 64)

		profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(VadConfidenceThreshold)
		if profileSettings.Vad_enabled {
			profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Show()
		} else {
			profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Hide()
		}
		profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Energy))
		profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Pause))
		profileForm.Items[10].Widget.(*fyne.Container).Objects[0].(*widget.Slider).SetValue(float64(profileSettings.Phrase_time_limit))

		if profileSettings.Ai_device != nil {
			profileForm.Items[11].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Ai_device.(string))
		}
		profileForm.Items[12].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Model)
		profileForm.Items[13].Widget.(*widget.Check).SetChecked(profileSettings.Fp16)
		// spacer
		profileForm.Items[15].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_device)
		profileForm.Items[16].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_size)
		profileForm.Items[18].Widget.(*widget.Check).SetChecked(profileSettings.Tts_enabled)
		profileForm.Items[19].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Tts_ai_device)

		profileForm.OnSubmit = func() {
			profileSettings.Websocket_ip = profileForm.Items[0].Widget.(*fyne.Container).Objects[0].(*widget.Entry).Text
			profileSettings.Websocket_port, _ = strconv.Atoi(profileForm.Items[0].Widget.(*fyne.Container).Objects[1].(*widget.Entry).Text)

			profileSettings.Device_index, _ = strconv.Atoi(profileForm.Items[2].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value)

			profileSettings.Device_out_index, _ = strconv.Atoi(profileForm.Items[4].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value)

			profileSettings.Vad_enabled = profileForm.Items[6].Widget.(*fyne.Container).Objects[0].(*widget.Check).Checked
			profileSettings.Vad_on_full_clip = profileForm.Items[6].Widget.(*fyne.Container).Objects[1].(*widget.Check).Checked
			profileSettings.Vad_confidence_threshold = fmt.Sprintf("%f", profileForm.Items[7].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value)

			profileSettings.Energy = int(profileForm.Items[8].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value)
			profileSettings.Pause = profileForm.Items[9].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value
			profileSettings.Phrase_time_limit = profileForm.Items[10].Widget.(*fyne.Container).Objects[0].(*widget.Slider).Value

			profileSettings.Ai_device = profileForm.Items[11].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Model = profileForm.Items[12].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Fp16 = profileForm.Items[13].Widget.(*widget.Check).Checked

			profileSettings.Txt_translator_device = profileForm.Items[15].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Txt_translator_size = profileForm.Items[16].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Tts_enabled = profileForm.Items[18].Widget.(*widget.Check).Checked
			profileSettings.Tts_ai_device = profileForm.Items[19].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			// update existing settings or create new one if it does not exist yet
			if Utilities.FileExists(settingsFiles[id]) {
				profileSettings.WriteYamlSettings(settingsFiles[id])
			} else {
				newProfileEntry := Profiles.Profile{
					SettingsFilename:      settingsFiles[id],
					Device_index:          profileSettings.Device_index,
					Device_out_index:      profileSettings.Device_out_index,
					Ai_device:             profileSettings.Ai_device,
					Model:                 profileSettings.Model,
					Txt_translator_size:   profileSettings.Txt_translator_size,
					Txt_translator_device: profileSettings.Txt_translator_device,
					Websocket_ip:          profileSettings.Websocket_ip,
					Websocket_port:        profileSettings.Websocket_port,
					Osc_ip:                profileSettings.Osc_ip,
					Osc_port:              profileSettings.Osc_port,
					Tts_enabled:           profileSettings.Tts_enabled,
					Tts_ai_device:         profileSettings.Tts_ai_device,
					Fp16:                  profileSettings.Fp16,

					Phrase_time_limit: profileSettings.Phrase_time_limit,
					Pause:             profileSettings.Pause,
					Energy:            profileSettings.Energy,
				}
				newProfileEntry.Save(settingsFiles[id])
			}
			Settings.Config = profileSettings

			// closes profile window, stop audio device and call onClose
			playBackDevice.Stop()
			onClose()
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

	mainContent := container.NewHSplit(
		container.NewMax(profileHelpTextContent, profileListContent),
		container.NewBorder(newProfileRow, nil, nil, nil, profileList),
	)
	mainContent.SetOffset(0.6)

	return mainContent
}
