package Pages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

type CurrentPlaybackDevice struct {
	InputDeviceName  string
	OutputDeviceName string
	InputWaveWidget  *widget.ProgressBar
	OutputWaveWidget *widget.ProgressBar
	Context          *malgo.Context

	stopChannel   chan bool
	playTestAudio bool
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

	byteReader := bytes.NewReader(resourceTestWav.Content())

	testAudioReader := wav.NewReader(byteReader)

	testAudioFormat, err := testAudioReader.Format()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testAudioChannels := uint32(testAudioFormat.NumChannels)
	testAudioSampleRate := testAudioFormat.SampleRate

	//######################

	ctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendWinmm}, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	captureDevices, err := ctx.Devices(malgo.Capture)
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

	playbackDevices, err := ctx.Devices(malgo.Playback)
	if err != nil {
		fmt.Println(err)
	}
	selectedPlaybackDeviceIndex := -1
	for index, deviceInfo := range playbackDevices {
		if deviceInfo.Name() == c.OutputDeviceName {
			selectedPlaybackDeviceIndex = index
			fmt.Println("Found input device: ", deviceInfo.Name(), " at index: ", index)
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
	deviceConfig.Playback.Channels = testAudioChannels
	//deviceConfig.SampleRate = 44100
	deviceConfig.SampleRate = testAudioSampleRate
	deviceConfig.Alsa.NoMMap = 1

	sizeInBytesCapture := uint32(malgo.SampleSizeInBytes(deviceConfig.Capture.Format))
	sizeInBytesPlayback := uint32(malgo.SampleSizeInBytes(deviceConfig.Playback.Format))

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
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		//os.Exit(1)
	}

	// run as long as no stop signal is received
	c.stopChannel = make(chan bool)
	for {
		select {
		case <-c.stopChannel:
			fmt.Println("stopping...")
			device.Uninit()
			return
		}
	}

}

func GetAudioDevices(deviceType malgo.DeviceType) ([]CustomWidget.TextValueOption, error) {

	deviceList, _ := Utilities.GetAudioDevices(deviceType)

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

func CreateProfileWindow(onClose func()) fyne.CanvasObject {
	playBackDevice := CurrentPlaybackDevice{}
	go playBackDevice.Init()

	audioInputDevices, _ := GetAudioDevices(malgo.Capture)
	audioOutputDevices, _ := GetAudioDevices(malgo.Playback)

	BuildProfileForm := func() fyne.CanvasObject {
		profileForm := widget.NewForm()
		websocketIp := widget.NewEntry()
		websocketIp.SetText("127.0.0.1")
		websocketPort := widget.NewEntry()
		websocketPort.SetText("5000")

		audioInputProgress := playBackDevice.InputWaveWidget
		audioOutputProgress := container.NewBorder(nil, nil, nil, widget.NewButton("Test", func() {
			playBackDevice.PlayStopTestAudio()
		}), playBackDevice.OutputWaveWidget)

		profileForm.Append("Websocket IP", websocketIp)
		profileForm.Append("Websocket Port", websocketPort)
		profileForm.Append("", layout.NewSpacer())

		profileForm.Append("Audio Input (mic)", CustomWidget.NewTextValueSelect("device_index", audioInputDevices,
			func(s CustomWidget.TextValueOption) {
				println(s.Value)
				playBackDevice.Stop()
				playBackDevice.InputDeviceName = s.Text
				go playBackDevice.Init()
			},
			0),
		)
		profileForm.Append("", audioInputProgress)

		profileForm.Append("Audio Output (speaker)", CustomWidget.NewTextValueSelect("device_out_index", audioOutputDevices,
			func(s CustomWidget.TextValueOption) {
				println(s.Value)
				playBackDevice.Stop()
				playBackDevice.OutputDeviceName = s.Text
				go playBackDevice.Init()
			},
			0),
		)
		profileForm.Append("", audioOutputProgress)

		profileForm.Append("", layout.NewSpacer())

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

		profileForm.Append("", layout.NewSpacer())

		profileForm.Append("Text Translation Size", CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
			{Text: "Small", Value: "small"},
			{Text: "Medium", Value: "medium"},
			{Text: "Large", Value: "large"},
		}, func(s CustomWidget.TextValueOption) {}, 0))

		profileForm.Append("Text to Speech Enable", widget.NewCheck("", func(b bool) {}))

		profileForm.Append("A.I. Device for Text to Speech", CustomWidget.NewTextValueSelect("tts_ai_device", []CustomWidget.TextValueOption{
			{Text: "CUDA", Value: "cuda"},
			{Text: "CPU", Value: "cpu"},
		}, func(s CustomWidget.TextValueOption) {}, 0))
		return profileForm
	}

	//profileOptions := []string{"device_index", "device_out_index", "ai_device", "model", "txt_translator_size", "websocket_ip", "websocket_port", "tts_enabled", "tts_ai_device"}
	//profileWindow := Settings.BuildSettingsForm(profileOptions, "").(*widget.Form)
	profileListContent := container.NewVScroll(BuildProfileForm())

	// build profile list
	var settingsFiles []string
	files, err := os.ReadDir("./")
	if err != nil {
		println(err)
	}
	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml")) {
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
		//profileSettings := Settings.Conf{
		//	Websocket_ip:        "127.0.0.1",
		//	Websocket_port:      5000,
		//	Device_index:        -1,
		//	Device_out_index:    -1,
		//	Ai_device:           "cuda",
		//	Model:               "tiny",
		//	Txt_translator_size: "small",
		//	Tts_enabled:         true,
		//	Tts_ai_device:       "cuda",
		//}
		profileSettings := Settings.Conf{
			SettingsFilename:    settingsFiles[id],
			Websocket_ip:        "127.0.0.1",
			Websocket_port:      5000,
			Device_index:        -1,
			Device_out_index:    -1,
			Ai_device:           "cuda",
			Model:               "tiny",
			Txt_translator_size: "small",
			Tts_enabled:         true,
			Tts_ai_device:       "cuda",
			Current_language:    "",
			Osc_ip:              "127.0.0.1",
			Osc_port:            9000,
		}
		profileSettings.LoadYamlSettings(settingsFiles[id])
		profileForm := profileListContent.Content.(*widget.Form)
		profileForm.SubmitText = "Save and Load Profile"
		profileForm.Items[0].Widget.(*widget.Entry).SetText(profileSettings.Websocket_ip)
		profileForm.Items[1].Widget.(*widget.Entry).SetText(strconv.Itoa(profileSettings.Websocket_port))
		// spacer
		if profileSettings.Device_index != nil {
			deviceInValue := "-1"
			switch profileSettings.Device_index.(type) {
			case int:
				deviceInValue = strconv.Itoa(profileSettings.Device_index.(int))
			case string:
				deviceInValue = profileSettings.Device_index.(string)
			}
			deviceInWidget := profileForm.Items[3].Widget.(*CustomWidget.TextValueSelect)
			if profileSettings.Device_index != nil && deviceInWidget.GetSelected().Value != deviceInValue {
				deviceInWidget.SetSelected(deviceInValue)
			}
		}
		// audio progressbar
		if profileSettings.Device_out_index != nil {
			deviceOutValue := "-1"
			switch profileSettings.Device_out_index.(type) {
			case int:
				deviceOutValue = strconv.Itoa(profileSettings.Device_out_index.(int))
			case string:
				deviceOutValue = profileSettings.Device_out_index.(string)
			}
			deviceOutWidget := profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect)
			if deviceOutWidget.GetSelected().Value != deviceOutValue {
				profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect).SetSelected(deviceOutValue)
			}
		}
		// audio progressbar
		// spacer
		if profileSettings.Ai_device != nil {
			profileForm.Items[8].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Ai_device.(string))
		}
		profileForm.Items[9].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Model)
		// spacer
		profileForm.Items[11].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Txt_translator_size)
		profileForm.Items[12].Widget.(*widget.Check).SetChecked(profileSettings.Tts_enabled)
		profileForm.Items[13].Widget.(*CustomWidget.TextValueSelect).SetSelected(profileSettings.Tts_ai_device)

		profileForm.OnSubmit = func() {
			profileSettings.Websocket_ip = profileForm.Items[0].Widget.(*widget.Entry).Text
			profileSettings.Websocket_port, _ = strconv.Atoi(profileForm.Items[1].Widget.(*widget.Entry).Text)

			profileSettings.Device_index = profileForm.Items[3].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Device_out_index = profileForm.Items[5].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Ai_device = profileForm.Items[8].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Model = profileForm.Items[9].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			profileSettings.Txt_translator_size = profileForm.Items[11].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value
			profileSettings.Tts_enabled = profileForm.Items[12].Widget.(*widget.Check).Checked
			profileSettings.Tts_ai_device = profileForm.Items[13].Widget.(*CustomWidget.TextValueSelect).GetSelected().Value

			// update existing settings or create new one if it does not exist yet
			if Utilities.FileExists(settingsFiles[id]) {
				profileSettings.WriteYamlSettings(settingsFiles[id])
			} else {
				newProfileEntry := Profiles.Profile{
					SettingsFilename:    settingsFiles[id],
					Device_index:        profileSettings.Device_index,
					Device_out_index:    profileSettings.Device_out_index,
					Ai_device:           profileSettings.Ai_device,
					Model:               profileSettings.Model,
					Txt_translator_size: profileSettings.Txt_translator_size,
					Websocket_ip:        profileSettings.Websocket_ip,
					Websocket_port:      profileSettings.Websocket_port,
					Tts_enabled:         profileSettings.Tts_enabled,
					Tts_ai_device:       profileSettings.Tts_ai_device,
				}
				newProfileEntry.Save(settingsFiles[id])
			}
			Settings.Config = profileSettings
			//Settings.ConfigLoaded = true

			// closes profile window, stop audio device and call onClose
			playBackDevice.Stop()
			onClose()
		}

		profileForm.Refresh()

		//profileListContent.Content = profileForm

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

	newProfileRow := container.NewBorder(nil, nil, nil, widget.NewButton("New", func() {
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
		profileListContent,
		container.NewBorder(newProfileRow, nil, nil, nil, profileList),
		//container.NewMax(profileList),
	)
	mainContent.SetOffset(0.6)

	return mainContent
}
