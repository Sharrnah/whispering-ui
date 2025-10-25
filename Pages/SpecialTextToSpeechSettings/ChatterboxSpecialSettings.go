package SpecialTextToSpeechSettings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"whispering-tiger-ui/CustomWidget"
)

func BuildChatterboxSpecialSettings() fyne.CanvasObject {

	defaultValues := map[string]interface{}{
		"language":           "en",
		"streaming_mode":     "segment",
		"precision":          "float32",
		"seed":               "",
		"temperature":        0.8,
		"exaggeration":       0.5,
		"cfg_weight":         0.5,
		"max_new_tokens":     256,
		"repetition_penalty": 1.7,
	}

	// Helper: convert various numeric types (int/float/string) to float64
	asFloat64 := func(v interface{}) float64 {
		switch t := v.(type) {
		case float64:
			return t
		case float32:
			return float64(t)
		case int:
			return float64(t)
		case int64:
			return float64(t)
		case int32:
			return float64(t)
		case uint:
			return float64(t)
		case uint64:
			return float64(t)
		case uint32:
			return float64(t)
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
		}
		return 0
	}

	// Helper: clamp to an arbitrary [min,max]
	clamp := func(v, min, max float64) float64 {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}

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
	languageSetting := GetSpecialSettingFallback("tts_chatterbox", "language", defaultValues["language"]).(string)
	languageSelect.SetSelected(languageSetting)

	streamingModeSelect := CustomWidget.NewTextValueSelect("streaming_mode", []CustomWidget.TextValueOption{
		{Text: "Segments", Value: "segment"},
		{Text: "Tokens (Extreme)", Value: "token"},
	}, nil, 0)
	streamingModeSetting := GetSpecialSettingFallback("tts_chatterbox", "streaming_mode", defaultValues["streaming_mode"]).(string)
	streamingModeSelect.SetSelected(streamingModeSetting)

	precisionInput := CustomWidget.NewTextValueSelect("precision", []CustomWidget.TextValueOption{
		{Text: "Float32", Value: "float32"},
		{Text: "Float16", Value: "float16"},
	}, nil, 0)
	precisionInputSetting := GetSpecialSettingFallback("tts_chatterbox", "precision", defaultValues["precision"]).(string)
	precisionInput.SetSelected(precisionInputSetting)

	seedInput := widget.NewEntry()
	seedInput.PlaceHolder = lang.L("Enter manual seed")
	// Load seed (optional)
	if seed := GetSpecialSettingFallback("tts_chatterbox", "seed", defaultValues["seed"]).(string); seed != "" {
		seedInput.SetText(seed)
	}

	temperatureSlider := widget.NewSlider(0.05, 5.0)
	if precisionInputSetting == "float16" {
		temperatureSlider.Min = 0.55
	} else {
		temperatureSlider.Min = 0.05
	}
	temperatureSlider.Step = 0.05
	{
		val := asFloat64(GetSpecialSettingFallback("tts_chatterbox", "temperature", defaultValues["temperature"]))
		temperatureSlider.SetValue(clamp(val, temperatureSlider.Min, temperatureSlider.Max))
	}
	temperatureSliderState := widget.NewLabel(fmt.Sprintf("%.2f", temperatureSlider.Value))

	exaggerationSlider := widget.NewSlider(0.25, 2.0)
	exaggerationSlider.Step = 0.05
	{
		val := asFloat64(GetSpecialSettingFallback("tts_chatterbox", "exaggeration", defaultValues["exaggeration"]))
		exaggerationSlider.SetValue(clamp(val, exaggerationSlider.Min, exaggerationSlider.Max))
	}
	exaggerationSliderState := widget.NewLabel(fmt.Sprintf("%.2f", exaggerationSlider.Value))

	cfgSlider := widget.NewSlider(0.2, 1.0)
	cfgSlider.Step = 0.05
	{
		val := asFloat64(GetSpecialSettingFallback("tts_chatterbox", "cfg_weight", defaultValues["cfg_weight"]))
		cfgSlider.SetValue(clamp(val, cfgSlider.Min, cfgSlider.Max))
	}
	cfgSliderState := widget.NewLabel(fmt.Sprintf("%.2f", cfgSlider.Value))

	maxNewTokensSlider := widget.NewSlider(1, 1000)
	maxNewTokensSlider.Step = 1
	{
		val := asFloat64(GetSpecialSettingFallback("tts_chatterbox", "max_new_tokens", defaultValues["max_new_tokens"]))
		maxNewTokensSlider.SetValue(clamp(val, maxNewTokensSlider.Min, maxNewTokensSlider.Max))
	}
	maxNewTokensSliderState := widget.NewLabel(fmt.Sprintf("%.0f", maxNewTokensSlider.Value))

	repetitionPenaltySlider := widget.NewSlider(0.1, 2.0)
	repetitionPenaltySlider.Step = 0.1
	{
		val := asFloat64(GetSpecialSettingFallback("tts_chatterbox", "repetition_penalty", defaultValues["repetition_penalty"]))
		repetitionPenaltySlider.SetValue(clamp(val, repetitionPenaltySlider.Min, repetitionPenaltySlider.Max))
	}
	repetitionPenaltySliderState := widget.NewLabel(fmt.Sprintf("%.1f", repetitionPenaltySlider.Value))

	updateSpecialTTSSettings := func() {
		UpdateSpecialTTSSettings("tts_chatterbox", "language", languageSelect.GetCurrentValueOptionEntry().Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "streaming_mode", streamingModeSelect.GetSelected().Value)

		UpdateSpecialTTSSettings("tts_chatterbox", "precision", precisionInput.GetSelected().Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "seed", seedInput.Text)
		UpdateSpecialTTSSettings("tts_chatterbox", "temperature", temperatureSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "exaggeration", exaggerationSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "cfg_weight", cfgSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "max_new_tokens", maxNewTokensSlider.Value)
		UpdateSpecialTTSSettings("tts_chatterbox", "repetition_penalty", repetitionPenaltySlider.Value)
	}

	languageSelect.OnSubmitted = func(value string) {
		updateSpecialTTSSettings()
	}
	streamingModeSelect.OnChanged = func(value CustomWidget.TextValueOption) {
		updateSpecialTTSSettings()
	}

	precisionInput.OnChanged = func(value CustomWidget.TextValueOption) {
		updateSpecialTTSSettings()
		if value.Value == "float16" {
			temperatureSlider.Min = 0.55
		} else {
			temperatureSlider.Min = 0.05
		}
		// Ensure value respects new min
		temperatureSlider.SetValue(clamp(temperatureSlider.Value, temperatureSlider.Min, temperatureSlider.Max))
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
	maxNewTokensSlider.OnChanged = func(f float64) {
		maxNewTokensSliderState.SetText(fmt.Sprintf("%.0f", f))
		updateSpecialTTSSettings()
	}
	repetitionPenaltySlider.OnChanged = func(f float64) {
		repetitionPenaltySliderState.SetText(fmt.Sprintf("%.1f", f))
		updateSpecialTTSSettings()
	}

	resetBtn := widget.NewButton(lang.L("Reset"), func() {
		languageSelect.SetSelected(defaultValues["language"].(string))
		streamingModeSelect.SetSelected(defaultValues["streaming_mode"].(string))
		precisionInput.SetSelected(defaultValues["precision"].(string))
		seedInput.SetText(defaultValues["seed"].(string))
		temperatureSlider.SetValue(clamp(asFloat64(defaultValues["temperature"]), temperatureSlider.Min, temperatureSlider.Max))
		exaggerationSlider.SetValue(clamp(asFloat64(defaultValues["exaggeration"]), exaggerationSlider.Min, exaggerationSlider.Max))
		cfgSlider.SetValue(clamp(asFloat64(defaultValues["cfg_weight"]), cfgSlider.Min, cfgSlider.Max))
		maxNewTokensSlider.SetValue(clamp(asFloat64(defaultValues["max_new_tokens"]), maxNewTokensSlider.Min, maxNewTokensSlider.Max))
		repetitionPenaltySlider.SetValue(clamp(asFloat64(defaultValues["repetition_penalty"]), repetitionPenaltySlider.Min, repetitionPenaltySlider.Max))
		//updateSpecialTTSSettings()
	})

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
				container.NewVBox(
					container.NewGridWithColumns(2,
						container.New(layout.NewFormLayout(),
							widget.NewLabel(lang.L("Precision")+":"),
							precisionInput,
							widget.NewLabel(lang.L("Seed")+":"),
							seedInput,
						),
						container.New(layout.NewFormLayout(),
							widget.NewLabel(lang.L("Reset to defaults")+":"),
							resetBtn,
						),
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Temperature")+":"),
						container.NewBorder(nil, nil, nil, temperatureSliderState, temperatureSlider),
						widget.NewLabel(lang.L("Emotion exaggeration")+":"),
						container.NewBorder(nil, nil, nil, exaggerationSliderState, exaggerationSlider),
						widget.NewLabel(lang.L("CFG/Pace")+":"),
						container.NewBorder(nil, nil, nil, cfgSliderState, cfgSlider),
						widget.NewLabel(lang.L("Max new tokens")+":"),
						container.NewBorder(nil, nil, nil, maxNewTokensSliderState, maxNewTokensSlider),
						widget.NewLabel(lang.L("Repetition penalty")+":"),
						container.NewBorder(nil, nil, nil, repetitionPenaltySliderState, repetitionPenaltySlider),
					),
				),
			),
		),
	)
	return advancedSettings
}
