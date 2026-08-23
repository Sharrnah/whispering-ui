package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
)

const (
	kokoroThorstenModel = "Kokoro-German-Thorsten"
	kokoroRussianModel  = "Kokoro-Russian-v2"
)

func kokoroCanonicalModel(displayValue string) string {
	if canonical, ok := Fields.TtsModelSelectionValues[displayValue]; ok && len(canonical) >= 2 {
		return canonical[1]
	}
	return displayValue
}

func kokoroForcedLanguage(displayValue string) (CustomWidget.TextValueOption, bool) {
	switch kokoroCanonicalModel(displayValue) {
	case kokoroThorstenModel:
		return CustomWidget.TextValueOption{Text: lang.L("German"), Value: "d"}, true
	case kokoroRussianModel:
		return CustomWidget.TextValueOption{Text: "Russian", Value: "r"}, true
	default:
		return CustomWidget.TextValueOption{}, false
	}
}

func kokoroStockLanguageOptions() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{
		{Text: "English (US)", Value: "a"},
		{Text: "English (British)", Value: "b"},
		{Text: "Spanish", Value: "e"},
		{Text: "French", Value: "f"},
		{Text: "Hindi", Value: "h"},
		{Text: "Italian", Value: "i"},
		{Text: "Japanese", Value: "j"},
		{Text: "Brazilian Portuguese", Value: "p"},
		{Text: "Chinese", Value: "z"},
	}
}

func setKokoroLanguageOptions(languageSelect *CustomWidget.CompletionEntry, options []CustomWidget.TextValueOption) {
	languageSelect.OptionsTextValue = append([]CustomWidget.TextValueOption(nil), options...)
	displayOptions := make([]string, 0, len(options))
	for _, option := range options {
		displayOptions = append(displayOptions, option.Text)
	}
	languageSelect.SetOptions(displayOptions)
}

func buildKokoroSpecialSettings() (fyne.CanvasObject, *CustomWidget.CompletionEntry) {

	languageSelect := CustomWidget.NewCompletionEntry([]string{})
	setKokoroLanguageOptions(languageSelect, kokoroStockLanguageOptions())

	languageSetting := GetSpecialSettingFallback("tts_kokoro", "language", "a").(string)
	languageSelect.SetSelected(languageSetting)

	updateSpecialTTSSettingsKokoro := func() {
		selected := languageSelect.GetCurrentValueOptionEntry()
		if selected == nil {
			return
		}

		UpdateSpecialTTSSettings("tts_kokoro", "language", selected.Value)
	}

	languageSelect.OnSubmitted = func(value string) {
		updateSpecialTTSSettingsKokoro()
	}

	applyLanguageForModel := func(displayValue string, notifyBackend bool) {
		targetLanguage := ""
		selected := languageSelect.GetCurrentValueOptionEntry()
		if forcedLanguage, forced := kokoroForcedLanguage(displayValue); forced {
			targetLanguage = forcedLanguage.Value
			setKokoroLanguageOptions(languageSelect, []CustomWidget.TextValueOption{forcedLanguage})
			languageSelect.SetSelected(forcedLanguage.Value)
			languageSelect.Disable()
		} else {
			setKokoroLanguageOptions(languageSelect, kokoroStockLanguageOptions())
			languageSelect.Enable()
			if selected == nil || selected.Value == "d" || selected.Value == "r" {
				targetLanguage = "a"
				languageSelect.SetSelected("a")
			} else {
				languageSelect.SetSelected(selected.Value)
			}
		}
		languageSelect.Refresh()

		if targetLanguage == "" {
			return
		}
		if notifyBackend {
			updateSpecialTTSSettingsKokoro()
		} else {
			// Page construction can happen before the unbuffered WebSocket send
			// channel has a receiver. Keep the profile state coherent locally and
			// wait for an actual model change before sending anything.
			SetSpecialTTSSetting("tts_kokoro", "language", targetLanguage)
		}
	}
	Fields.TtsModelSelectionChanged = func(displayValue string) {
		applyLanguageForModel(displayValue, true)
	}
	applyLanguageForModel(Fields.Field.TtsModelCombo.Selected, false)

	advancedSettings := container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Language")+":"),
			languageSelect,
		),
	)
	return advancedSettings, languageSelect
}

func BuildKokoroSpecialSettings() fyne.CanvasObject {
	advancedSettings, _ := buildKokoroSpecialSettings()
	return advancedSettings
}
