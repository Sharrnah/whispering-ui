package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func BuildMaya1SpecialSettings() fyne.CanvasObject {
	defaultValues := map[string]interface{}{
		"voice_description": "Realistic male voice in the 30s age with american accent. Normal pitch, warm timbre, conversational pacing.",
	}

	voiceDescriptionInput := widget.NewEntry()
	// Load seed (optional)
	if voiceDescription := GetSpecialSettingFallback("tts_maya1", "voice_description", defaultValues["voice_description"]).(string); voiceDescription != "" {
		voiceDescriptionInput.SetText(voiceDescription)
	}

	updateSpecialTTSSettings := func() {
		voiceDescription := voiceDescriptionInput.Text

		UpdateSpecialTTSSettings("tts_maya1", "voice_description", voiceDescription)
	}

	voiceDescriptionInput.OnChanged = func(value string) {
		updateSpecialTTSSettings()
	}

	advancedSettings := container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Voice Description")+":"),
			voiceDescriptionInput,
		),
	)
	return advancedSettings
}
