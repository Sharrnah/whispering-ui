package SpecialTextToSpeechSettings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
)

func BuildChatterboxSpecialSettings() fyne.CanvasObject {

	languageSelect := CustomWidget.NewCompletionEntry([]string{})
	languageSelect.SetValueOptions([]CustomWidget.TextValueOption{
		{Text: "Auto", Value: "auto"},
		{Text: "Arabic", Value: "ar"},
		{Text: "Danish", Value: "da"},
		{Text: "German", Value: "de"},
		{Text: "Greek", Value: "el"},
		{Text: "English", Value: "en"},
		{Text: "Spanish", Value: "es"},
		{Text: "Finnish", Value: "fi"},
		{Text: "French", Value: "fr"},
		{Text: "Hebrew", Value: "he"},
		{Text: "Hindi", Value: "hi"},
		{Text: "Italian", Value: "it"},
		{Text: "Japanese", Value: "ja"},
		{Text: "Korean", Value: "ko"},
		{Text: "Malay", Value: "ms"},
		{Text: "Dutch", Value: "nl"},
		{Text: "Norwegian", Value: "no"},
		{Text: "Polish", Value: "pl"},
		{Text: "Portuguese", Value: "pt"},
		{Text: "Russian", Value: "ru"},
		{Text: "Swedish", Value: "sv"},
		{Text: "Swahili", Value: "sw"},
		{Text: "Turkish", Value: "tr"},
		{Text: "Chinese", Value: "zh"},
	})
	languageSetting := GetSpecialSettingFallback("tts_chatterbox", "language", "en").(string)
	languageSelect.SetSelected(languageSetting)

	streamingModeSelect := CustomWidget.NewTextValueSelect("streaming_mode", []CustomWidget.TextValueOption{
		{Text: "Segments", Value: "segment"},
		{Text: "Tokens (Extreme)", Value: "token"},
	}, nil, 0)
	streamingModeSetting := GetSpecialSettingFallback("tts_chatterbox", "streaming_mode", "segment").(string)
	streamingModeSelect.SetSelected(streamingModeSetting)

	// Clamp helper to keep values in [0,1]
	clamp01 := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	seedInput := widget.NewEntry()
	seedInput.PlaceHolder = lang.L("Enter manual seed")
	// Load seed (optional)
	if seed := GetSpecialSettingFallback("tts_chatterbox", "seed", "").(string); seed != "" {
		seedInput.SetText(seed)
	}

	temperatureSlider := widget.NewSlider(0.05, 5.0)
	temperatureSlider.Step = 0.05
	temperatureSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_chatterbox", "temperature", 0.8).(float64)))
	temperatureSliderState := widget.NewLabel(fmt.Sprintf("%.2f", temperatureSlider.Value))

	exaggerationSlider := widget.NewSlider(0.25, 2.0)
	exaggerationSlider.Step = 0.05
	exaggerationSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_chatterbox", "exaggeration", 0.5).(float64)))
	exaggerationSliderState := widget.NewLabel(fmt.Sprintf("%.2f", exaggerationSlider.Value))

	cfgSlider := widget.NewSlider(0.2, 1.0)
	cfgSlider.Step = 0.05
	cfgSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_chatterbox", "cfg_weight", 0.5).(float64)))
	cfgSliderState := widget.NewLabel(fmt.Sprintf("%.2f", cfgSlider.Value))

	updateSpecialTTSSettings := func() {
		UpdateSpecialTTSSettings("tts_chatterbox", "language", languageSelect.GetCurrentValueOptionEntry().Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "streaming_mode", streamingModeSelect.GetSelected().Value)

		UpdateSpecialTTSSettings("tts_chatterbox", "seed", seedInput.Text)
		UpdateSpecialTTSSettings("tts_chatterbox", "temperature", temperatureSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "exaggeration", exaggerationSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "cfg_weight", cfgSlider.Value)
	}

	languageSelect.OnSubmitted = func(value string) {
		updateSpecialTTSSettings()
	}
	streamingModeSelect.OnChanged = func(value CustomWidget.TextValueOption) {
		updateSpecialTTSSettings()
	}

	seedInput.OnChanged = func(s string) {
		updateSpecialTTSSettings()
	}
	temperatureSlider.OnChanged = func(f float64) {
		temperatureSliderState.SetText(fmt.Sprintf("%.2f", f))
		updateSpecialTTSSettings()
	}
	exaggerationSlider.OnChanged = func(f float64) {
		exaggerationSliderState.SetText(fmt.Sprintf("%.2f", f))
		updateSpecialTTSSettings()
	}
	cfgSlider.OnChanged = func(f float64) {
		cfgSliderState.SetText(fmt.Sprintf("%.2f", f))
		updateSpecialTTSSettings()
	}

	advancedSettings := container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Language")+":"),
			languageSelect,
			widget.NewLabel(lang.L("Streaming Mode")+":"),
			streamingModeSelect,
		),
		widget.NewAccordion(
			widget.NewAccordionItem(lang.L("More Options"),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Seed")+":"),
						seedInput,
						widget.NewLabel(lang.L("Temperature")+":"),
						container.NewBorder(nil, nil, nil, temperatureSliderState, temperatureSlider),
						widget.NewLabel(lang.L("Emotion exaggeration")+":"),
						container.NewBorder(nil, nil, nil, exaggerationSliderState, exaggerationSlider),
						widget.NewLabel(lang.L("CFG/Pace")+":"),
						container.NewBorder(nil, nil, nil, cfgSliderState, cfgSlider),
					),
				),
			),
		),
	)
	return advancedSettings
}
