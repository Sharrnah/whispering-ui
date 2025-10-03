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

	languageSelect := CustomWidget.NewCompletionEntry([]string{})
	languageSelect.SetValueOptions([]CustomWidget.TextValueOption{
		{Text: "English (US)", Value: "a"},
		{Text: "English (British)", Value: "b"},
		{Text: "Spanish", Value: "e"},
		{Text: "French", Value: "f"},
		{Text: "Hindi", Value: "h"},
		{Text: "Italian", Value: "i"},
		{Text: "Japanese", Value: "j"},
		{Text: "Brazilian Portuguese", Value: "p"},
		{Text: "Chinese", Value: "z"},
	})

	languageSetting := GetSpecialSettingFallback("tts_kokoro", "language", "a").(string)
	languageSelect.SetSelected(languageSetting)

	updateSpecialTTSSettingsKokoro := func() {
		languageSelection := languageSelect.GetCurrentValueOptionEntry().Value

		UpdateSpecialTTSSettings("tts_kokoro", "language", languageSelection)
	}

	languageSelect.OnSubmitted = func(value string) {
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
