package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/Fields"
)

func CreateTextToSpeechWindow() fyne.CanvasObject {
	ttsModels := container.New(layout.NewFormLayout(), widget.NewLabel("Model:"), Fields.Field.TtsModelCombo)

	saveRandomVoiceButton := widget.NewButtonWithIcon("Save Random Voice", theme.DocumentSaveIcon(), func() {
		sendMessage := Fields.SendMessageStruct{
			Type: "tts_voice_save_req",
		}
		sendMessage.SendMessage()
		Fields.Field.TtsVoiceCombo.SetSelected("last")
	})
	ttsVoices := container.New(layout.NewFormLayout(), widget.NewLabel("Voice:"), Fields.Field.TtsVoiceCombo)

	ttsVoicesSaveBtnLayout := container.NewBorder(nil, nil, nil, saveRandomVoiceButton, ttsVoices)

	ttsModelVoiceRow := container.New(layout.NewGridLayout(2), ttsModels, ttsVoicesSaveBtnLayout)

	transcriptionRow := container.New(layout.NewGridLayout(1), Fields.Field.TranscriptionTranslationInput)

	sendButton := widget.NewButtonWithIcon("Send to Text 2 Speech", theme.MediaPlayIcon(), func() {
		valueData := struct {
			Text     string `json:"text"`
			ToDevice bool   `json:"to_device"`
			Download bool   `json:"download"`
		}{
			Text:     Fields.Field.TranscriptionTranslationInput.Text,
			ToDevice: true,
			Download: false,
		}
		sendMessage := Fields.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
	})
	sendButton.Importance = widget.HighImportance

	testButton := widget.NewButton("Test the Voice", func() {
		valueData := struct {
			Text        string      `json:"text"`
			ToDevice    bool        `json:"to_device"`
			DeviceIndex interface{} `json:"device_index"`
			Download    bool        `json:"download"`
		}{
			Text:        Fields.Field.TranscriptionTranslationInput.Text,
			ToDevice:    true,
			DeviceIndex: nil, // send to default device
			Download:    false,
		}
		sendMessage := Fields.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
	})

	buttonRow := container.NewHBox(layout.NewSpacer(),
		testButton,
		sendButton,
	)

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			ttsModelVoiceRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.New(layout.NewVBoxLayout(), buttonRow),
		),
	)

	return mainContent
}
