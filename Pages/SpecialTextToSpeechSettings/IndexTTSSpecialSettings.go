package SpecialTextToSpeechSettings

import (
	"strconv"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const indexTTSGermanModel = "IndexTTS-2.5-German"

func indexTTSCanonicalModel(displayValue string) string {
	if canonical, ok := Fields.TtsModelSelectionValues[displayValue]; ok && len(canonical) >= 2 {
		return canonical[1]
	}
	return displayValue
}

func indexTTSStockLanguageOptions() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{
		{Text: "Auto", Value: "auto"},
		{Text: "Chinese", Value: "zh"},
		{Text: "English", Value: "en"},
		{Text: "Japanese", Value: "ja"},
		{Text: "Arabic", Value: "ar"},
		{Text: "Spanish", Value: "es"},
	}
}

func indexTTSLanguageOptions(displayValue string) []CustomWidget.TextValueOption {
	options := indexTTSStockLanguageOptions()
	if indexTTSCanonicalModel(displayValue) == indexTTSGermanModel {
		options = append(options, CustomWidget.TextValueOption{Text: lang.L("German"), Value: "de"})
	}
	return options
}

func containsIndexTTSLanguage(options []CustomWidget.TextValueOption, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func buildIndexTTSSpecialSettings() (fyne.CanvasObject, *CustomWidget.TextValueSelect) {
	modelDisplay := Fields.Field.TtsModelCombo.Selected
	languageOptions := indexTTSLanguageOptions(modelDisplay)
	languageSelect := CustomWidget.NewTextValueSelect("index_tts_language", languageOptions, nil, 0)
	languageSetting := GetSpecialSettingFallback("tts_index_tts", "language", "auto").(string)
	if !containsIndexTTSLanguage(languageOptions, languageSetting) {
		languageSetting = "auto"
		SetSpecialTTSSetting("tts_index_tts", "language", languageSetting)
	}
	languageSelect.SetSelected(languageSetting)

	precisionSelect := CustomWidget.NewTextValueSelect("index_tts_precision", []CustomWidget.TextValueOption{
		{Text: "BFloat16 (recommended for CUDA)", Value: "bfloat16"},
		{Text: "Float32", Value: "float32"},
	}, nil, 0)
	precisionSelect.SetSelected(GetSpecialSettingFallback("tts_index_tts", "precision", "bfloat16").(string))

	seedInput := widget.NewEntry()
	seedInput.SetText(strconv.Itoa(GetSpecialSettingFallback("tts_index_tts", "seed", -1).(int)))

	newSlider := func(min, max, step, value float64, decimals int) (*widget.Slider, *widget.Label, fyne.CanvasObject) {
		slider := widget.NewSlider(min, max)
		slider.Step = step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(strconv.FormatFloat(value, 'f', decimals, 64))
		valueLabel.Alignment = fyne.TextAlignTrailing
		return slider, valueLabel, container.NewBorder(nil, nil, nil, valueLabel, slider)
	}

	duration, durationLabel, durationControl := newSlider(0.5, 2.0, 0.01, GetSpecialSettingFallback("tts_index_tts", "duration_factor", 1.0).(float64), 2)
	beams, beamsLabel, beamsControl := newSlider(1, 5, 1, GetSpecialSettingFallback("tts_index_tts", "num_beams", 3.0).(float64), 0)
	temperature, temperatureLabel, temperatureControl := newSlider(0.01, 2.0, 0.01, GetSpecialSettingFallback("tts_index_tts", "temperature", 0.8).(float64), 2)
	topP, topPLabel, topPControl := newSlider(0.01, 1.0, 0.01, GetSpecialSettingFallback("tts_index_tts", "top_p", 0.8).(float64), 2)
	topK, topKLabel, topKControl := newSlider(0, 200, 1, GetSpecialSettingFallback("tts_index_tts", "top_k", 30.0).(float64), 0)
	repetition, repetitionLabel, repetitionControl := newSlider(1, 20, 0.1, GetSpecialSettingFallback("tts_index_tts", "repetition_penalty", 10.0).(float64), 1)
	maxMel, maxMelLabel, maxMelControl := newSlider(128, 1815, 1, GetSpecialSettingFallback("tts_index_tts", "max_mel_tokens", 1500.0).(float64), 0)
	segmentTokens, segmentTokensLabel, segmentTokensControl := newSlider(20, 500, 1, GetSpecialSettingFallback("tts_index_tts", "max_text_tokens_per_segment", 120.0).(float64), 0)
	streamingSegmentLength, streamingSegmentLengthLabel, streamingSegmentLengthControl := newSlider(40, 1000, 10, GetSpecialSettingFallback("tts_index_tts", "streaming_segment_goal_length", 120.0).(float64), 0)
	segmentPause, segmentPauseLabel, segmentPauseControl := newSlider(1, 5000, 10, GetSpecialSettingFallback("tts_index_tts", "pause_between_segments_ms", 200.0).(float64), 0)
	voiceChangePause, voiceChangePauseLabel, voiceChangePauseControl := newSlider(0, 5000, 10, GetSpecialSettingFallback("tts_index_tts", "pause_between_voice_change_ms", 400.0).(float64), 0)

	doSample := widget.NewCheck(lang.L("Enable"), nil)
	doSample.SetChecked(GetSpecialSettingFallback("tts_index_tts", "do_sample", true).(bool))
	textNormalization := widget.NewCheck(lang.L("Enable"), nil)
	textNormalization.SetChecked(GetSpecialSettingFallback("tts_index_tts", "text_normalization", true).(bool))
	emotionEnabled := widget.NewCheck(lang.L("Enable"), nil)
	emotionEnabled.SetChecked(GetSpecialSettingFallback("tts_index_tts", "emotion_enabled", false).(bool))
	emotionRandom := widget.NewCheck(lang.L("Enable"), nil)
	emotionRandom.SetChecked(GetSpecialSettingFallback("tts_index_tts", "emotion_random_reference", false).(bool))

	emotionHappy, emotionHappyLabel, emotionHappyControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_happy", 0.0).(float64), 2)
	emotionAngry, emotionAngryLabel, emotionAngryControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_angry", 0.0).(float64), 2)
	emotionSad, emotionSadLabel, emotionSadControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_sad", 0.0).(float64), 2)
	emotionAfraid, emotionAfraidLabel, emotionAfraidControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_afraid", 0.0).(float64), 2)
	emotionDisgusted, emotionDisgustedLabel, emotionDisgustedControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_disgusted", 0.0).(float64), 2)
	emotionMelancholic, emotionMelancholicLabel, emotionMelancholicControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_melancholic", 0.0).(float64), 2)
	emotionSurprised, emotionSurprisedLabel, emotionSurprisedControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_surprised", 0.0).(float64), 2)
	emotionCalm, emotionCalmLabel, emotionCalmControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_calm", 0.0).(float64), 2)
	emotionStrength, emotionStrengthLabel, emotionStrengthControl := newSlider(0, 1, 0.05, GetSpecialSettingFallback("tts_index_tts", "emotion_strength", 1.0).(float64), 2)

	update := func() {
		seed, err := strconv.Atoi(seedInput.Text)
		if err != nil {
			seed = -1
		}
		UpdateSpecialTTSSettings("tts_index_tts", "language", languageSelect.GetSelected().Value)
		UpdateTTSPrecision("tts_index_tts", precisionSelect.GetSelected().Value)
		UpdateSpecialTTSSettings("tts_index_tts", "seed", seed)
		UpdateSpecialTTSSettings("tts_index_tts", "duration_factor", duration.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "num_beams", int(beams.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "do_sample", doSample.Checked)
		UpdateSpecialTTSSettings("tts_index_tts", "temperature", temperature.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "top_p", topP.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "top_k", int(topK.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "repetition_penalty", repetition.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "max_mel_tokens", int(maxMel.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "max_text_tokens_per_segment", int(segmentTokens.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "streaming_segment_goal_length", int(streamingSegmentLength.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "pause_between_segments_ms", int(segmentPause.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "pause_between_voice_change_ms", int(voiceChangePause.Value))
		UpdateSpecialTTSSettings("tts_index_tts", "text_normalization", textNormalization.Checked)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_enabled", emotionEnabled.Checked)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_happy", emotionHappy.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_angry", emotionAngry.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_sad", emotionSad.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_afraid", emotionAfraid.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_disgusted", emotionDisgusted.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_melancholic", emotionMelancholic.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_surprised", emotionSurprised.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_calm", emotionCalm.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_strength", emotionStrength.Value)
		UpdateSpecialTTSSettings("tts_index_tts", "emotion_random_reference", emotionRandom.Checked)
	}

	bindSlider := func(slider *widget.Slider, label *widget.Label, decimals int) {
		slider.OnChanged = func(value float64) {
			label.SetText(strconv.FormatFloat(value, 'f', decimals, 64))
			update()
		}
	}
	for _, binding := range []struct {
		slider   *widget.Slider
		label    *widget.Label
		decimals int
	}{
		{duration, durationLabel, 2}, {beams, beamsLabel, 0},
		{temperature, temperatureLabel, 2}, {topP, topPLabel, 2},
		{topK, topKLabel, 0}, {repetition, repetitionLabel, 1},
		{maxMel, maxMelLabel, 0}, {segmentTokens, segmentTokensLabel, 0},
		{streamingSegmentLength, streamingSegmentLengthLabel, 0},
		{segmentPause, segmentPauseLabel, 0}, {voiceChangePause, voiceChangePauseLabel, 0},
		{emotionHappy, emotionHappyLabel, 2}, {emotionAngry, emotionAngryLabel, 2},
		{emotionSad, emotionSadLabel, 2}, {emotionAfraid, emotionAfraidLabel, 2},
		{emotionDisgusted, emotionDisgustedLabel, 2}, {emotionMelancholic, emotionMelancholicLabel, 2},
		{emotionSurprised, emotionSurprisedLabel, 2}, {emotionCalm, emotionCalmLabel, 2},
		{emotionStrength, emotionStrengthLabel, 2},
	} {
		bindSlider(binding.slider, binding.label, binding.decimals)
	}
	languageSelect.OnChanged = func(CustomWidget.TextValueOption) { update() }
	precisionSelect.OnChanged = func(CustomWidget.TextValueOption) { update() }
	seedInput.OnChanged = func(string) { update() }
	for _, check := range []*widget.Check{doSample, textNormalization, emotionEnabled, emotionRandom} {
		check.OnChanged = func(bool) { update() }
	}
	Fields.TtsModelSelectionChanged = func(displayValue string) {
		selectedLanguage := "auto"
		if selected := languageSelect.GetSelected(); selected != nil {
			selectedLanguage = selected.Value
		}
		previousLanguage := selectedLanguage
		options := indexTTSLanguageOptions(displayValue)
		languageSelect.SetValueOptions(options)
		if !containsIndexTTSLanguage(options, selectedLanguage) {
			selectedLanguage = "auto"
		}
		onChanged := languageSelect.OnChanged
		languageSelect.OnChanged = nil
		languageSelect.SetSelected(selectedLanguage)
		languageSelect.OnChanged = onChanged
		if selectedLanguage != previousLanguage {
			UpdateSpecialTTSSettings("tts_index_tts", "language", selectedLanguage)
		}
	}
	voiceSwitchingInfo := widget.NewButtonWithIcon(lang.L("Voice Switching"), theme.InfoIcon(), func() {
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			dialog.ShowInformation(
				lang.L("Voice Switching"),
				lang.L("Put [voice_name] at the start of a line to switch to the matching voice file. Untagged text and [main] use the selected voice."),
				windows[0],
			)
		}
	})
	voiceSwitchingInfo.Importance = widget.LowImportance

	return container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		widget.NewAccordion(
			widget.NewAccordionItem(lang.L("General"), container.NewVBox(
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Language")+":"), languageSelect,
						widget.NewLabel(lang.L("Precision")+":"), precisionSelect,
						widget.NewLabel(lang.L("Seed")+":"), seedInput,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Duration Factor (higher is slower)")+":"), durationControl,
						widget.NewLabel(lang.L("Beams")+":"), beamsControl,
						widget.NewLabel(lang.L("Text Normalization")+":"), textNormalization,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Streaming Segment Target (characters)")+":"), streamingSegmentLengthControl,
					widget.NewLabel(lang.L("Pause Between Segments (ms)")+":"), segmentPauseControl,
					widget.NewLabel(lang.L("Pause Between Voice Changes (ms)")+":"), voiceChangePauseControl,
				),
				container.NewHBox(layout.NewSpacer(), voiceSwitchingInfo),
			)),
			widget.NewAccordionItem(lang.L("Sampling"), container.New(layout.NewFormLayout(),
				widget.NewLabel(lang.L("Sampling")+":"), doSample,
				widget.NewLabel(lang.L("Temperature")+":"), temperatureControl,
				widget.NewLabel("Top P:"), topPControl,
				widget.NewLabel("Top K:"), topKControl,
				widget.NewLabel(lang.L("Repetition Penalty")+":"), repetitionControl,
				widget.NewLabel(lang.L("Max Mel Tokens")+":"), maxMelControl,
				widget.NewLabel(lang.L("Max Text Tokens Per Segment")+":"), segmentTokensControl,
			)),
			widget.NewAccordionItem(lang.L("Emotion"), container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Emotion Control")+":"), emotionEnabled,
					widget.NewLabel(lang.L("Random Emotion Prototype")+":"), emotionRandom,
				),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Happy")+":"), emotionHappyControl,
						widget.NewLabel(lang.L("Sad")+":"), emotionSadControl,
						widget.NewLabel(lang.L("Disgusted")+":"), emotionDisgustedControl,
						widget.NewLabel(lang.L("Surprised")+":"), emotionSurprisedControl,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Angry")+":"), emotionAngryControl,
						widget.NewLabel(lang.L("Afraid")+":"), emotionAfraidControl,
						widget.NewLabel(lang.L("Melancholic")+":"), emotionMelancholicControl,
						widget.NewLabel(lang.L("Calm")+":"), emotionCalmControl,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Strength")+":"), emotionStrengthControl,
				),
			)),
		),
	), languageSelect
}

func BuildIndexTTSSpecialSettings() fyne.CanvasObject {
	advancedSettings, _ := buildIndexTTSSpecialSettings()
	return advancedSettings
}
