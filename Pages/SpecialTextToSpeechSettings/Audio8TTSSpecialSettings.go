package SpecialTextToSpeechSettings

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/SendMessageChannel"
)

// BuildAudio8TTSSpecialSettings exposes the generation and exact-transcript
// cloning controls for the Audio8 TTS Preview checkpoint.
func BuildAudio8TTSSpecialSettings() fyne.CanvasObject {
	const settingsGroup = "tts_audio8"

	precisionSelect := CustomWidget.NewTextValueSelect("audio8_tts_precision", []CustomWidget.TextValueOption{
		{Text: lang.L("Automatic (BF16 on supported CUDA)"), Value: "auto"},
		{Text: lang.L("BFloat16"), Value: "bfloat16"},
		{Text: lang.L("Float32"), Value: "float32"},
	}, nil, 0)
	precisionSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "precision", "auto").(string))
	attentionSelect := CustomWidget.NewTextValueSelect("audio8_tts_attention", []CustomWidget.TextValueOption{
		{Text: lang.L("Auto (FlashAttention 2 when supported)"), Value: "auto"},
		{Text: lang.L("FlashAttention 2"), Value: "flash_attention_2"},
		{Text: lang.L("Compatible PyTorch attention"), Value: "sdpa"},
	}, nil, 0)
	attentionSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "attention", "auto").(string))

	cloneModeSelect := CustomWidget.NewTextValueSelect("audio8_tts_clone_mode", []CustomWidget.TextValueOption{
		{Text: lang.L("Auto (clone with exact transcript)"), Value: "auto"},
		{Text: lang.L("Required (exact transcript)"), Value: "required"},
		{Text: lang.L("Disabled (unconditioned voice)"), Value: "disabled"},
	}, nil, 0)
	cloneModeSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "clone_mode", "auto").(string))

	referenceTextInput := widget.NewMultiLineEntry()
	referenceTextInput.SetMinRowsVisible(3)
	referenceTextInput.SetPlaceHolder(lang.L("Exact transcript of the selected reference audio (optional in Auto mode)"))
	referenceTextInput.SetText(GetSpecialSettingFallback(settingsGroup, "reference_text", "").(string))

	seedInput := widget.NewEntry()
	seedInput.SetText(strconv.Itoa(GetSpecialSettingFallback(settingsGroup, "seed", -1).(int)))
	maxTokensInput := widget.NewEntry()
	maxTokensInput.SetText(strconv.Itoa(GetSpecialSettingFallback(settingsGroup, "max_new_tokens", 512).(int)))

	doSample := widget.NewCheck(lang.L("Enable"), nil)
	doSample.SetChecked(GetSpecialSettingFallback(settingsGroup, "do_sample", true).(bool))

	newSlider := func(min, max, step, value float64, decimals int) (*widget.Slider, *widget.Label, fyne.CanvasObject) {
		slider := widget.NewSlider(min, max)
		slider.Step = step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(strconv.FormatFloat(value, 'f', decimals, 64))
		valueLabel.Alignment = fyne.TextAlignTrailing
		return slider, valueLabel, container.NewBorder(nil, nil, nil, valueLabel, slider)
	}

	temperature, temperatureLabel, temperatureControl := newSlider(0.01, 2.0, 0.01, GetSpecialSettingFallback(settingsGroup, "temperature", 0.8).(float64), 2)
	topP, topPLabel, topPControl := newSlider(0.01, 1.0, 0.01, GetSpecialSettingFallback(settingsGroup, "top_p", 0.95).(float64), 2)
	topK, topKLabel, topKControl := newSlider(1, 500, 1, GetSpecialSettingFallback(settingsGroup, "top_k", 50.0).(float64), 0)
	segmentCharacters, segmentCharactersLabel, segmentCharactersControl := newSlider(20, 150, 1, GetSpecialSettingFallback(settingsGroup, "streaming_segment_characters", 140.0).(float64), 0)
	segmentPause, segmentPauseLabel, segmentPauseControl := newSlider(0, 5000, 10, GetSpecialSettingFallback(settingsGroup, "pause_between_segments_ms", 120.0).(float64), 0)

	parseInt := func(value string, fallback int) int {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return parsed
	}

	resetting := false
	update := func() {
		if resetting {
			return
		}
		UpdateSpecialTTSSettings(settingsGroup, "precision", precisionSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "attention", attentionSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "clone_mode", cloneModeSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "reference_text", referenceTextInput.Text)
		UpdateSpecialTTSSettings(settingsGroup, "seed", parseInt(seedInput.Text, -1))
		UpdateSpecialTTSSettings(settingsGroup, "do_sample", doSample.Checked)
		UpdateSpecialTTSSettings(settingsGroup, "temperature", temperature.Value)
		UpdateSpecialTTSSettings(settingsGroup, "top_p", topP.Value)
		UpdateSpecialTTSSettings(settingsGroup, "top_k", int(topK.Value))
		UpdateSpecialTTSSettings(settingsGroup, "max_new_tokens", parseInt(maxTokensInput.Text, 512))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_segment_characters", int(segmentCharacters.Value))
		UpdateSpecialTTSSettings(settingsGroup, "pause_between_segments_ms", int(segmentPause.Value))
	}

	precisionSelect.OnChanged = func(CustomWidget.TextValueOption) { update() }
	attentionSelect.OnChanged = func(CustomWidget.TextValueOption) { update() }
	cloneModeSelect.OnChanged = func(CustomWidget.TextValueOption) { update() }
	referenceTextInput.OnChanged = func(string) { update() }
	seedInput.OnChanged = func(string) { update() }
	maxTokensInput.OnChanged = func(string) { update() }
	doSample.OnChanged = func(bool) { update() }

	bindSlider := func(slider *widget.Slider, label *widget.Label, decimals int) {
		slider.OnChanged = func(value float64) {
			label.SetText(strconv.FormatFloat(value, 'f', decimals, 64))
			update()
		}
	}
	bindSlider(temperature, temperatureLabel, 2)
	bindSlider(topP, topPLabel, 2)
	bindSlider(topK, topKLabel, 0)
	bindSlider(segmentCharacters, segmentCharactersLabel, 0)
	bindSlider(segmentPause, segmentPauseLabel, 0)

	resetButton := widget.NewButton(lang.L("Reset"), func() {
		resetting = true
		precisionSelect.SetSelected("auto")
		attentionSelect.SetSelected("auto")
		cloneModeSelect.SetSelected("auto")
		referenceTextInput.SetText("")
		seedInput.SetText("-1")
		doSample.SetChecked(true)
		temperature.SetValue(0.8)
		topP.SetValue(0.95)
		topK.SetValue(50)
		maxTokensInput.SetText("512")
		segmentCharacters.SetValue(140)
		segmentPause.SetValue(120)
		resetting = false
		update()
	})

	saveReferenceButton := widget.NewButtonWithIcon(lang.L("Save last generation as clone reference"), theme.DocumentSaveIcon(), func() {
		SendMessageChannel.SendMessageStruct{Type: "tts_voice_save_req"}.SendMessage()
	})

	return container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		widget.NewAccordion(
			widget.NewAccordionItem(lang.L("General"), container.NewGridWithColumns(2,
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Precision")+":"), precisionSelect,
					widget.NewLabel(lang.L("Attention")+":"), attentionSelect,
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Reset to defaults")+":"), resetButton,
				),
			)),
			widget.NewAccordionItem(lang.L("Voice Cloning"), container.New(layout.NewFormLayout(),
				widget.NewLabel(lang.L("Clone Mode")+":"), cloneModeSelect,
				widget.NewLabel(lang.L("Reference Transcript")+":"), referenceTextInput,
				widget.NewLabel(lang.L("Reuse Generated Voice")+":"), saveReferenceButton,
			)),
			widget.NewAccordionItem(lang.L("Generation"), container.NewVBox(
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Seed")+":"), seedInput,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Max new tokens")+":"), maxTokensInput,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Sampling")+":"), doSample,
					widget.NewLabel(lang.L("Temperature")+":"), temperatureControl,
					widget.NewLabel(lang.L("Top P")+":"), topPControl,
					widget.NewLabel(lang.L("Top K")+":"), topKControl,
				),
			)),
			widget.NewAccordionItem(lang.L("Streaming"), container.New(layout.NewFormLayout(),
				widget.NewLabel(lang.L("Fallback Segment Characters")+":"), segmentCharactersControl,
				widget.NewLabel(lang.L("Fallback Segment Pause (ms)")+":"), segmentPauseControl,
			)),
		),
	)
}
