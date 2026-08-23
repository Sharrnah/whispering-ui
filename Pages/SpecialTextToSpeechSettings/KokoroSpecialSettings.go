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

const kokoroThorstenModel = "Kokoro-German-Thorsten"

func kokoroCanonicalModel(displayValue string) string {
	if canonical, ok := Fields.TtsModelSelectionValues[displayValue]; ok && len(canonical) >= 2 {
		return canonical[1]
	}
	return displayValue
}

func BuildKokoroSpecialSettings() fyne.CanvasObject {

	languageSelect := CustomWidget.NewCompletionEntry([]string{})
	languageSelect.SetValueOptions([]CustomWidget.TextValueOption{
		{Text: "English (US)", Value: "a"},
		{Text: "English (British)", Value: "b"},
		{Text: lang.L("German"), Value: "d"},
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
		if kokoroCanonicalModel(displayValue) == kokoroThorstenModel {
			targetLanguage = "d"
			selected := languageSelect.GetCurrentValueOptionEntry()
			if selected == nil || selected.Value != "d" {
				languageSelect.SetSelected("d")
				languageSelect.Refresh()
			}
			languageSelect.Disable()
		} else {
			languageSelect.Enable()
			selected := languageSelect.GetCurrentValueOptionEntry()
			if selected != nil && selected.Value == "d" {
				targetLanguage = "a"
				languageSelect.SetSelected("a")
				languageSelect.Refresh()
			}
		}

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
	return advancedSettings
}
