package SpecialTextToSpeechSettings

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func BuildZonos2SpecialSettings() fyne.CanvasObject {
	seedInput := widget.NewEntry()
	seedInput.SetText(fmt.Sprintf("%d", GetSpecialSettingFallback("tts_zonos2", "seed", -1).(int)))
	attention := widget.NewSelect([]string{"auto", "flash_attention", "sdpa"}, nil)
	attention.SetSelected(GetSpecialSettingFallback("tts_zonos2", "attention", "auto").(string))
	if attention.Selected == "" {
		attention.SetSelected("auto")
	}

	newSlider := func(min, max, step, value float64, decimals int) (*widget.Slider, *widget.Label, fyne.CanvasObject) {
		slider := widget.NewSlider(min, max)
		slider.Step = step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(strconv.FormatFloat(value, 'f', decimals, 64))
		valueLabel.Alignment = fyne.TextAlignTrailing
		return slider, valueLabel, container.NewBorder(nil, nil, nil, valueLabel, slider)
	}

	maxTokens, maxTokensLabel, maxTokensControl := newSlider(32, 6000, 32, GetSpecialSettingFallback("tts_zonos2", "max_new_tokens", 1024.0).(float64), 0)
	temperature, temperatureLabel, temperatureControl := newSlider(0, 2, 0.01, GetSpecialSettingFallback("tts_zonos2", "temperature", 1.15).(float64), 2)
	topK, topKLabel, topKControl := newSlider(0, 1026, 1, GetSpecialSettingFallback("tts_zonos2", "top_k", 106.0).(float64), 0)
	topP, topPLabel, topPControl := newSlider(0, 1, 0.01, GetSpecialSettingFallback("tts_zonos2", "top_p", 0.0).(float64), 2)
	minP, minPLabel, minPControl := newSlider(0, 1, 0.01, GetSpecialSettingFallback("tts_zonos2", "min_p", 0.18).(float64), 2)
	repetitionWindow, repetitionWindowLabel, repetitionWindowControl := newSlider(0, 512, 1, GetSpecialSettingFallback("tts_zonos2", "repetition_window", 50.0).(float64), 0)
	repetitionPenalty, repetitionPenaltyLabel, repetitionPenaltyControl := newSlider(1, 2, 0.01, GetSpecialSettingFallback("tts_zonos2", "repetition_penalty", 1.2).(float64), 2)
	repetitionCodebooks, repetitionCodebooksLabel, repetitionCodebooksControl := newSlider(-1, 9, 1, GetSpecialSettingFallback("tts_zonos2", "repetition_codebooks", 8.0).(float64), 0)
	fadeOut, fadeOutLabel, fadeOutControl := newSlider(0, 2000, 10, GetSpecialSettingFallback("tts_zonos2", "fade_out_ms", 0.0).(float64), 0)

	emotionHappy, emotionHappyLabel, emotionHappyControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_happy", 0.0).(float64), 2)
	emotionSad, emotionSadLabel, emotionSadControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_sad", 0.0).(float64), 2)
	emotionAngry, emotionAngryLabel, emotionAngryControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_angry", 0.0).(float64), 2)
	emotionSurprised, emotionSurprisedLabel, emotionSurprisedControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_surprised", 0.0).(float64), 2)
	emotionValence, emotionValenceLabel, emotionValenceControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_valence", 0.0).(float64), 2)
	emotionArousal, emotionArousalLabel, emotionArousalControl := newSlider(-1, 1, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_arousal", 0.0).(float64), 2)
	emotionStrength, emotionStrengthLabel, emotionStrengthControl := newSlider(0, 3, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_strength", 1.0).(float64), 2)
	emotionCFG, emotionCFGLabel, emotionCFGControl := newSlider(1, 3, 0.05, GetSpecialSettingFallback("tts_zonos2", "emotion_cfg_scale", 1.0).(float64), 2)

	speakingRateOptions := []string{
		"Default", "0: 0-8", "1: 8-11", "2: 11-14", "3: 14-17",
		"4: 17-21", "5: 21-28", "6: 28-40", "7: 40+",
	}
	speakingRate := widget.NewSelect(speakingRateOptions, nil)
	rateValue := GetSpecialSettingFallback("tts_zonos2", "speaking_rate", -1).(int)
	if rateValue >= 0 && rateValue < 8 {
		speakingRate.SetSelected(speakingRateOptions[rateValue+1])
	} else {
		speakingRate.SetSelected(speakingRateOptions[0])
	}

	qualitySelect := func(options []string, key string, fallback int) *widget.Select {
		selectWidget := widget.NewSelect(options, nil)
		value := GetSpecialSettingFallback("tts_zonos2", key, fallback).(int)
		if value >= 0 && value+1 < len(options) {
			selectWidget.SetSelected(options[value+1])
		} else {
			selectWidget.SetSelected(options[0])
		}
		return selectWidget
	}
	loudness := qualitySelect([]string{"Default", "0: below -50", "1: -50--45.5", "2: -45.5--41", "3: -41--36.5", "4: -36.5--32", "5: -32--27.5", "6: -27.5--23", "7: -23--18.5", "8: -18.5--14", "9: -14--9.5", "10: -9.5--5", "11: -5+ LUFS"}, "loudness_lufs", -1)
	snr := qualitySelect([]string{"Default", "0: below 0", "1: 0-6", "2: 6-12", "3: 12-18", "4: 18-24", "5: 24-30", "6: 30-36", "7: 36-42", "8: 42-48", "9: 48-54", "10: 54-60", "11: 60+ dB"}, "estimated_snr", -1)
	maximumPause := qualitySelect([]string{"Default", "0: 0-0.5", "1: 0.5-1", "2: 1-1.5", "3: 1.5-2", "4: 2-2.5", "5: 2.5-3", "6: 3-3.5", "7: 3.5-4", "8: 4-4.5", "9: 4.5-5", "10: 5-5.5", "11: 5.5-6 s"}, "maximum_pause", -1)
	bandlimit := qualitySelect([]string{"Default", "0: 495-3433", "1: 3433-6371", "2: 6371-9310", "3: 9310-12248", "4: 12248-15186", "5: 15186-18124", "6: 18124-21062", "7: 21062-24000 Hz"}, "estimated_bandlimit_hz", -1)
	leadingSilence := qualitySelect([]string{"Default", "0: 0-0.05", "1: 0.05-0.1", "2: 0.1-0.25", "3: 0.25-0.5", "4: 0.5-1", "5: 1-2", "6: 2-4", "7: 4+ s"}, "leading_silence", -1)
	trailingSilence := qualitySelect([]string{"Default", "0: 0-0.05", "1: 0.05-0.1", "2: 0.1-0.25", "3: 0.25-0.5", "4: 0.5-1", "5: 1-2", "6: 2-4", "7: 4+ s"}, "trailing_silence", 3)

	qualityEnabled := widget.NewCheck(lang.L("Enable"), nil)
	qualityEnabled.SetChecked(GetSpecialSettingFallback("tts_zonos2", "quality_enabled", true).(bool))
	cleanBackground := widget.NewCheck(lang.L("Enable"), nil)
	cleanBackground.SetChecked(GetSpecialSettingFallback("tts_zonos2", "clean_speaker_background", false).(bool))
	accurateMode := widget.NewCheck(lang.L("Enable"), nil)
	accurateMode.SetChecked(GetSpecialSettingFallback("tts_zonos2", "accurate_mode", true).(bool))
	emotionEnabled := widget.NewCheck(lang.L("Enable"), nil)
	emotionEnabled.SetChecked(GetSpecialSettingFallback("tts_zonos2", "emotion_enabled", false).(bool))

	selectedBucket := func(selected string) int {
		if selected == "Default" || selected == "" {
			return -1
		}
		value, err := strconv.Atoi(strings.SplitN(selected, ":", 2)[0])
		if err != nil {
			return -1
		}
		return value
	}

	update := func() {
		seed, err := strconv.Atoi(seedInput.Text)
		if err != nil {
			seed = -1
		}
		UpdateSpecialTTSSettings("tts_zonos2", "seed", seed)
		UpdateSpecialTTSSettings("tts_zonos2", "attention", attention.Selected)
		UpdateSpecialTTSSettings("tts_zonos2", "max_new_tokens", int(maxTokens.Value))
		UpdateSpecialTTSSettings("tts_zonos2", "temperature", temperature.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "top_k", int(topK.Value))
		UpdateSpecialTTSSettings("tts_zonos2", "top_p", topP.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "min_p", minP.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "repetition_window", int(repetitionWindow.Value))
		UpdateSpecialTTSSettings("tts_zonos2", "repetition_penalty", repetitionPenalty.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "repetition_codebooks", int(repetitionCodebooks.Value))
		UpdateSpecialTTSSettings("tts_zonos2", "fade_out_ms", fadeOut.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "speaking_rate", selectedBucket(speakingRate.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "quality_enabled", qualityEnabled.Checked)
		UpdateSpecialTTSSettings("tts_zonos2", "loudness_lufs", selectedBucket(loudness.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "estimated_snr", selectedBucket(snr.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "maximum_pause", selectedBucket(maximumPause.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "estimated_bandlimit_hz", selectedBucket(bandlimit.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "leading_silence", selectedBucket(leadingSilence.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "trailing_silence", selectedBucket(trailingSilence.Selected))
		UpdateSpecialTTSSettings("tts_zonos2", "clean_speaker_background", cleanBackground.Checked)
		UpdateSpecialTTSSettings("tts_zonos2", "accurate_mode", accurateMode.Checked)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_enabled", emotionEnabled.Checked)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_happy", emotionHappy.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_sad", emotionSad.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_angry", emotionAngry.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_surprised", emotionSurprised.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_valence", emotionValence.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_arousal", emotionArousal.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_strength", emotionStrength.Value)
		UpdateSpecialTTSSettings("tts_zonos2", "emotion_cfg_scale", emotionCFG.Value)
	}

	bindSlider := func(slider *widget.Slider, label *widget.Label, decimals int) {
		slider.OnChanged = func(value float64) {
			label.SetText(strconv.FormatFloat(value, 'f', decimals, 64))
			update()
		}
	}
	bindSlider(maxTokens, maxTokensLabel, 0)
	bindSlider(temperature, temperatureLabel, 2)
	bindSlider(topK, topKLabel, 0)
	bindSlider(topP, topPLabel, 2)
	bindSlider(minP, minPLabel, 2)
	bindSlider(repetitionWindow, repetitionWindowLabel, 0)
	bindSlider(repetitionPenalty, repetitionPenaltyLabel, 2)
	bindSlider(repetitionCodebooks, repetitionCodebooksLabel, 0)
	bindSlider(fadeOut, fadeOutLabel, 0)
	bindSlider(emotionHappy, emotionHappyLabel, 2)
	bindSlider(emotionSad, emotionSadLabel, 2)
	bindSlider(emotionAngry, emotionAngryLabel, 2)
	bindSlider(emotionSurprised, emotionSurprisedLabel, 2)
	bindSlider(emotionValence, emotionValenceLabel, 2)
	bindSlider(emotionArousal, emotionArousalLabel, 2)
	bindSlider(emotionStrength, emotionStrengthLabel, 2)
	bindSlider(emotionCFG, emotionCFGLabel, 2)

	seedInput.OnChanged = func(string) { update() }
	for _, selectWidget := range []*widget.Select{attention, speakingRate, loudness, snr, maximumPause, bandlimit, leadingSilence, trailingSilence} {
		selectWidget.OnChanged = func(string) { update() }
	}
	for _, check := range []*widget.Check{qualityEnabled, cleanBackground, accurateMode, emotionEnabled} {
		check.OnChanged = func(bool) { update() }
	}

	return container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Seed")+":"), seedInput,
			widget.NewLabel(lang.L("Attention")+":"), attention,
			widget.NewLabel(lang.L("Speaking Rate")+":"), speakingRate,
			widget.NewLabel(lang.L("Fade Out (ms)")+":"), fadeOutControl,
			widget.NewLabel(lang.L("Clean Speaker Background")+":"), cleanBackground,
			widget.NewLabel(lang.L("Accurate Mode")+":"), accurateMode,
		),
		widget.NewAccordion(
			widget.NewAccordionItem(lang.L("Emotion"),
				container.NewVBox(
					widget.NewLabel(lang.L("Requires a selected voice. CFG above 1.0 uses approximately twice the inference compute.")),
					container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Emotion Control")+":"), emotionEnabled),
					container.NewGridWithColumns(2,
						container.New(layout.NewFormLayout(),
							widget.NewLabel(lang.L("Happy")+":"), emotionHappyControl,
							widget.NewLabel(lang.L("Sad")+":"), emotionSadControl,
							widget.NewLabel(lang.L("Valence")+":"), emotionValenceControl,
							widget.NewLabel(lang.L("Strength")+":"), emotionStrengthControl,
						),
						container.New(layout.NewFormLayout(),
							widget.NewLabel(lang.L("Angry")+":"), emotionAngryControl,
							widget.NewLabel(lang.L("Surprised")+":"), emotionSurprisedControl,
							widget.NewLabel(lang.L("Arousal")+":"), emotionArousalControl,
							widget.NewLabel("CFG:"), emotionCFGControl,
						),
					),
				),
			),
			widget.NewAccordionItem(lang.L("Quality Conditioning"),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Quality Conditioning")+":"), qualityEnabled,
					widget.NewLabel(lang.L("Loudness")+":"), loudness,
					widget.NewLabel(lang.L("Estimated SNR")+":"), snr,
					widget.NewLabel(lang.L("Maximum Pause")+":"), maximumPause,
					widget.NewLabel(lang.L("Bandlimit")+":"), bandlimit,
					widget.NewLabel(lang.L("Leading Silence")+":"), leadingSilence,
					widget.NewLabel(lang.L("Trailing Silence")+":"), trailingSilence,
				),
			),
			widget.NewAccordionItem(lang.L("Sampling"),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Maximum New Tokens")+":"), maxTokensControl,
					widget.NewLabel(lang.L("Temperature")+":"), temperatureControl,
					widget.NewLabel("Top K:"), topKControl,
					widget.NewLabel("Top P:"), topPControl,
					widget.NewLabel("Min P:"), minPControl,
					widget.NewLabel(lang.L("Repetition Window")+":"), repetitionWindowControl,
					widget.NewLabel(lang.L("Repetition Penalty")+":"), repetitionPenaltyControl,
					widget.NewLabel(lang.L("Repetition Codebooks")+":"), repetitionCodebooksControl,
				),
			),
		),
	)
}
