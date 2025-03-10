package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/SendMessageChannel"
)

func BuildKokoroSpecialSettings() fyne.CanvasObject {
	languageSelect := CustomWidget.NewTextValueSelect("language",
		[]CustomWidget.TextValueOption{
			{Text: "English (US)", Value: "a"},
			{Text: "English (British)", Value: "b"},
			{Text: "Spanish", Value: "e"},
			{Text: "French", Value: "f"},
			{Text: "Hindi", Value: "h"},
			{Text: "Italian", Value: "i"},
			{Text: "Japanese", Value: "j"},
			{Text: "Brazilian Portuguese", Value: "p"},
			{Text: "Chinese", Value: "z"},
		},
		func(option CustomWidget.TextValueOption) {

		},
		0,
	)

	languageSelect.SetSelected("a")

	updateSpecialTTSSettings := func() {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type: "tts_setting_special",
			Value: struct {
				Language string `json:"language"`
			}{
				Language: languageSelect.GetSelected().Value,
			},
		}
		sendMessage.SendMessage()
	}
	languageSelect.OnChanged = func(option CustomWidget.TextValueOption) {
		updateSpecialTTSSettings()
	}

	advancedSettings := container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Language")+":"),
			languageSelect,
		),
	)
	return advancedSettings
}
