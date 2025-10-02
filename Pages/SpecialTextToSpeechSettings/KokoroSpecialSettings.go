package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
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

	// check if Settings.Config.Special_settings has tts_language key
	//if val, ok := Settings.Config.Special_settings["tts_kokoro_language"]; ok {
	//	languageSelect.SetSelected(val.(string))
	//} else {
	//	languageSelect.SetSelected("a")
	//}

	languageSetting := GetSpecialTTSSettings("tts_kokoro", "language")
	if languageSetting != nil {
		languageSelect.SetSelected(languageSetting.(string))
	} else {
		languageSelect.SetSelected("a")
	}

	updateSpecialTTSSettingsKokoro := func() {
		languageSelection := languageSelect.GetSelected().Value

		UpdateSpecialTTSSettings("tts_kokoro", "language", languageSelection)

		//Settings.Config.Special_settings["tts_kokoro"].(map[string]interface{})["tts_kokoro_language"] = languageSelection
		//sendMessage := SendMessageChannel.SendMessageStruct{
		//	Type: "special_settings",
		//	Name: "tts_kokoro",
		//	Value: struct {
		//		TtsLanguage string `json:"language"`
		//	}{
		//		TtsLanguage: languageSelection,
		//	},
		//}
		//sendMessage.SendMessage()
	}
	languageSelect.OnChanged = func(option CustomWidget.TextValueOption) {
		updateSpecialTTSSettingsKokoro()
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
