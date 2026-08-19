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

// BuildQwen3TTSSpecialSettings exposes the controls shared by Qwen3-TTS Base,
// CustomVoice, and VoiceDesign checkpoints. Capabilities that do not apply to
// the selected checkpoint are safely ignored by the Python adapter.
func BuildQwen3TTSSpecialSettings() fyne.CanvasObject {
	const settingsGroup = "tts_qwen3_tts"

	languageSelect := CustomWidget.NewTextValueSelect("qwen3_tts_language", []CustomWidget.TextValueOption{
		{Text: "Auto", Value: "auto"},
		{Text: "Chinese", Value: "zh"},
		{Text: "English", Value: "en"},
		{Text: "Japanese", Value: "ja"},
		{Text: "Korean", Value: "ko"},
		{Text: "German", Value: "de"},
		{Text: "French", Value: "fr"},
		{Text: "Russian", Value: "ru"},
		{Text: "Portuguese", Value: "pt"},
		{Text: "Spanish", Value: "es"},
		{Text: "Italian", Value: "it"},
	}, nil, 0)
	languageSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "language", "auto").(string))

	precisionSelect := CustomWidget.NewTextValueSelect("qwen3_tts_precision", []CustomWidget.TextValueOption{
		{Text: "Auto (recommended)", Value: "auto"},
		{Text: "BFloat16", Value: "bfloat16"},
		{Text: "Float16 request (safe BF16/FP32 fallback)", Value: "float16"},
		{Text: "Float32", Value: "float32"},
	}, nil, 0)
	precisionSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "precision", "auto").(string))

	attentionSelect := CustomWidget.NewTextValueSelect("qwen3_tts_attention", []CustomWidget.TextValueOption{
		{Text: "Auto (FlashAttention 2 when supported)", Value: "auto"},
		{Text: "FlashAttention 2", Value: "flash_attention_2"},
		{Text: "PyTorch SDPA", Value: "sdpa"},
		{Text: "Eager", Value: "eager"},
	}, nil, 0)
	attentionSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "attention", "auto").(string))

	cloneModeSelect := CustomWidget.NewTextValueSelect("qwen3_tts_clone_mode", []CustomWidget.TextValueOption{
		{Text: "Auto (ICL with transcript, otherwise speaker embedding)", Value: "auto"},
		{Text: "ICL (highest fidelity; exact transcript required)", Value: "icl"},
		{Text: "Speaker embedding only (no transcript)", Value: "x_vector"},
	}, nil, 0)
	cloneModeSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "clone_mode", "auto").(string))

	textModeSelect := CustomWidget.NewTextValueSelect("qwen3_tts_text_mode", []CustomWidget.TextValueOption{
		{Text: "Automatic (official checkpoint defaults)", Value: "auto"},
		{Text: "Full-text conditioning", Value: "full_text"},
		{Text: "Incremental-text simulation", Value: "streaming_simulation"},
	}, nil, 0)
	textModeSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "model_text_mode", "auto").(string))

	streamingModeSelect := CustomWidget.NewTextValueSelect("qwen3_tts_streaming_mode", []CustomWidget.TextValueOption{
		{Text: "Live codec frames (best continuity)", Value: "codec"},
		{Text: "Parallel lookahead batches (fastest; may have seams)", Value: "lookahead"},
		{Text: "Completed text segments (compatibility fallback)", Value: "segment"},
	}, nil, 0)
	streamingModeSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "streaming_mode", "codec").(string))

	streamingBufferModeSelect := CustomWidget.NewTextValueSelect("qwen3_tts_streaming_buffer_mode", []CustomWidget.TextValueOption{
		{Text: "Adaptive (smooth playback; recommended)", Value: "adaptive"},
		{Text: "Fixed shared minimum (lowest latency; may underrun)", Value: "fixed"},
	}, nil, 0)
	streamingBufferModeSelect.SetSelected(GetSpecialSettingFallback(settingsGroup, "streaming_buffer_mode", "adaptive").(string))

	instructionInput := widget.NewMultiLineEntry()
	instructionInput.SetMinRowsVisible(3)
	instructionInput.SetPlaceHolder(lang.L("Describe the voice, accent, emotion, speaking style, or delivery"))
	instructionInput.SetText(GetSpecialSettingFallback(settingsGroup, "voice_instruction", "").(string))

	referenceTextInput := widget.NewMultiLineEntry()
	referenceTextInput.SetMinRowsVisible(3)
	referenceTextInput.SetPlaceHolder(lang.L("Exact transcript of the selected reference audio (optional in Auto mode)"))
	referenceTextInput.SetText(GetSpecialSettingFallback(settingsGroup, "reference_text", "").(string))

	seedInput := widget.NewEntry()
	seedInput.SetText(strconv.Itoa(GetSpecialSettingFallback(settingsGroup, "seed", -1).(int)))
	maxTokensInput := widget.NewEntry()
	maxTokensInput.SetText(strconv.Itoa(GetSpecialSettingFallback(settingsGroup, "max_new_tokens", 2048).(int)))

	applyProsody := widget.NewCheck(lang.L("Append the shared rate and pitch controls to the instruction"), nil)
	applyProsody.SetChecked(GetSpecialSettingFallback(settingsGroup, "apply_prosody_to_instruction", true).(bool))
	doSample := widget.NewCheck(lang.L("Enable"), nil)
	doSample.SetChecked(GetSpecialSettingFallback(settingsGroup, "do_sample", true).(bool))
	subtalkerDoSample := widget.NewCheck(lang.L("Enable"), nil)
	subtalkerDoSample.SetChecked(GetSpecialSettingFallback(settingsGroup, "subtalker_do_sample", true).(bool))

	newSlider := func(min, max, step, value float64, decimals int) (*widget.Slider, *widget.Label, fyne.CanvasObject) {
		slider := widget.NewSlider(min, max)
		slider.Step = step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(strconv.FormatFloat(value, 'f', decimals, 64))
		valueLabel.Alignment = fyne.TextAlignTrailing
		return slider, valueLabel, container.NewBorder(nil, nil, nil, valueLabel, slider)
	}

	temperature, temperatureLabel, temperatureControl := newSlider(0.01, 2.0, 0.01, GetSpecialSettingFallback(settingsGroup, "temperature", 0.9).(float64), 2)
	topP, topPLabel, topPControl := newSlider(0.01, 1.0, 0.01, GetSpecialSettingFallback(settingsGroup, "top_p", 1.0).(float64), 2)
	topK, topKLabel, topKControl := newSlider(0, 500, 1, GetSpecialSettingFallback(settingsGroup, "top_k", 50.0).(float64), 0)
	repetition, repetitionLabel, repetitionControl := newSlider(0.1, 5.0, 0.01, GetSpecialSettingFallback(settingsGroup, "repetition_penalty", 1.05).(float64), 2)
	subtalkerTemperature, subtalkerTemperatureLabel, subtalkerTemperatureControl := newSlider(0.01, 2.0, 0.01, GetSpecialSettingFallback(settingsGroup, "subtalker_temperature", 0.9).(float64), 2)
	subtalkerTopP, subtalkerTopPLabel, subtalkerTopPControl := newSlider(0.01, 1.0, 0.01, GetSpecialSettingFallback(settingsGroup, "subtalker_top_p", 1.0).(float64), 2)
	subtalkerTopK, subtalkerTopKLabel, subtalkerTopKControl := newSlider(0, 500, 1, GetSpecialSettingFallback(settingsGroup, "subtalker_top_k", 50.0).(float64), 0)
	codecFrames, codecFramesLabel, codecFramesControl := newSlider(1, 50, 1, GetSpecialSettingFallback(settingsGroup, "streaming_codec_frames", 2.0).(float64), 0)
	decoderContextFrames, decoderContextFramesLabel, decoderContextFramesControl := newSlider(0, 100, 1, GetSpecialSettingFallback(settingsGroup, "streaming_decoder_context_frames", 25.0).(float64), 0)
	bufferSafety, bufferSafetyLabel, bufferSafetyControl := newSlider(0, 5000, 50, GetSpecialSettingFallback(settingsGroup, "streaming_buffer_safety_ms", 500.0).(float64), 0)
	lookaheadBatchSize, lookaheadBatchSizeLabel, lookaheadBatchSizeControl := newSlider(2, 4, 1, GetSpecialSettingFallback(settingsGroup, "streaming_lookahead_batch_size", 3.0).(float64), 0)
	lookaheadCharacters, lookaheadCharactersLabel, lookaheadCharactersControl := newSlider(30, 500, 1, GetSpecialSettingFallback(settingsGroup, "streaming_lookahead_characters", 80.0).(float64), 0)
	lookaheadCodecFrames, lookaheadCodecFramesLabel, lookaheadCodecFramesControl := newSlider(1, 50, 1, GetSpecialSettingFallback(settingsGroup, "streaming_lookahead_codec_frames", 6.0).(float64), 0)
	lookaheadPause, lookaheadPauseLabel, lookaheadPauseControl := newSlider(0, 1000, 10, GetSpecialSettingFallback(settingsGroup, "streaming_lookahead_pause_ms", 0.0).(float64), 0)
	segmentCharacters, segmentCharactersLabel, segmentCharactersControl := newSlider(20, 1000, 1, GetSpecialSettingFallback(settingsGroup, "streaming_segment_characters", 180.0).(float64), 0)
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
		UpdateSpecialTTSSettings(settingsGroup, "language", languageSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "precision", precisionSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "attention", attentionSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "clone_mode", cloneModeSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "model_text_mode", textModeSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "voice_instruction", instructionInput.Text)
		UpdateSpecialTTSSettings(settingsGroup, "reference_text", referenceTextInput.Text)
		UpdateSpecialTTSSettings(settingsGroup, "apply_prosody_to_instruction", applyProsody.Checked)
		UpdateSpecialTTSSettings(settingsGroup, "seed", parseInt(seedInput.Text, -1))
		UpdateSpecialTTSSettings(settingsGroup, "do_sample", doSample.Checked)
		UpdateSpecialTTSSettings(settingsGroup, "temperature", temperature.Value)
		UpdateSpecialTTSSettings(settingsGroup, "top_p", topP.Value)
		UpdateSpecialTTSSettings(settingsGroup, "top_k", int(topK.Value))
		UpdateSpecialTTSSettings(settingsGroup, "repetition_penalty", repetition.Value)
		UpdateSpecialTTSSettings(settingsGroup, "subtalker_do_sample", subtalkerDoSample.Checked)
		UpdateSpecialTTSSettings(settingsGroup, "subtalker_temperature", subtalkerTemperature.Value)
		UpdateSpecialTTSSettings(settingsGroup, "subtalker_top_p", subtalkerTopP.Value)
		UpdateSpecialTTSSettings(settingsGroup, "subtalker_top_k", int(subtalkerTopK.Value))
		UpdateSpecialTTSSettings(settingsGroup, "max_new_tokens", parseInt(maxTokensInput.Text, 2048))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_mode", streamingModeSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "streaming_buffer_mode", streamingBufferModeSelect.GetSelected().Value)
		UpdateSpecialTTSSettings(settingsGroup, "streaming_codec_frames", int(codecFrames.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_decoder_context_frames", int(decoderContextFrames.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_buffer_safety_ms", int(bufferSafety.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_lookahead_batch_size", int(lookaheadBatchSize.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_lookahead_characters", int(lookaheadCharacters.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_lookahead_codec_frames", int(lookaheadCodecFrames.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_lookahead_pause_ms", int(lookaheadPause.Value))
		UpdateSpecialTTSSettings(settingsGroup, "streaming_segment_characters", int(segmentCharacters.Value))
		UpdateSpecialTTSSettings(settingsGroup, "pause_between_segments_ms", int(segmentPause.Value))
	}

	for _, selector := range []*CustomWidget.TextValueSelect{languageSelect, precisionSelect, attentionSelect, cloneModeSelect, textModeSelect, streamingModeSelect, streamingBufferModeSelect} {
		selector.OnChanged = func(CustomWidget.TextValueOption) { update() }
	}
	for _, entry := range []*widget.Entry{instructionInput, referenceTextInput, seedInput, maxTokensInput} {
		entry.OnChanged = func(string) { update() }
	}
	for _, check := range []*widget.Check{applyProsody, doSample, subtalkerDoSample} {
		check.OnChanged = func(bool) { update() }
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
		{temperature, temperatureLabel, 2}, {topP, topPLabel, 2}, {topK, topKLabel, 0},
		{repetition, repetitionLabel, 2}, {subtalkerTemperature, subtalkerTemperatureLabel, 2},
		{subtalkerTopP, subtalkerTopPLabel, 2}, {subtalkerTopK, subtalkerTopKLabel, 0},
		{codecFrames, codecFramesLabel, 0}, {decoderContextFrames, decoderContextFramesLabel, 0},
		{bufferSafety, bufferSafetyLabel, 0},
		{lookaheadBatchSize, lookaheadBatchSizeLabel, 0}, {lookaheadCharacters, lookaheadCharactersLabel, 0},
		{lookaheadCodecFrames, lookaheadCodecFramesLabel, 0},
		{lookaheadPause, lookaheadPauseLabel, 0},
		{segmentCharacters, segmentCharactersLabel, 0}, {segmentPause, segmentPauseLabel, 0},
	} {
		bindSlider(binding.slider, binding.label, binding.decimals)
	}

	resetButton := widget.NewButton(lang.L("Reset"), func() {
		resetting = true
		languageSelect.SetSelected("auto")
		precisionSelect.SetSelected("auto")
		attentionSelect.SetSelected("auto")
		cloneModeSelect.SetSelected("auto")
		textModeSelect.SetSelected("auto")
		streamingModeSelect.SetSelected("codec")
		streamingBufferModeSelect.SetSelected("adaptive")
		instructionInput.SetText("")
		referenceTextInput.SetText("")
		applyProsody.SetChecked(true)
		seedInput.SetText("-1")
		doSample.SetChecked(true)
		temperature.SetValue(0.9)
		topP.SetValue(1.0)
		topK.SetValue(50)
		repetition.SetValue(1.05)
		subtalkerDoSample.SetChecked(true)
		subtalkerTemperature.SetValue(0.9)
		subtalkerTopP.SetValue(1.0)
		subtalkerTopK.SetValue(50)
		maxTokensInput.SetText("2048")
		codecFrames.SetValue(2)
		decoderContextFrames.SetValue(25)
		bufferSafety.SetValue(500)
		lookaheadBatchSize.SetValue(3)
		lookaheadCharacters.SetValue(80)
		lookaheadCodecFrames.SetValue(6)
		lookaheadPause.SetValue(0)
		segmentCharacters.SetValue(180)
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
					widget.NewLabel(lang.L("Language")+":"), languageSelect,
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Reset to defaults")+":"), resetButton,
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Precision")+":"), precisionSelect,
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Attention")+":"), attentionSelect,
				),
			)),
			widget.NewAccordionItem(lang.L("Voice Cloning and Design"), container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Clone Mode")+":"), cloneModeSelect,
					widget.NewLabel(lang.L("Reference Transcript")+":"), referenceTextInput,
					widget.NewLabel(lang.L("Voice Instruction / Design")+":"), instructionInput,
					widget.NewLabel(lang.L("Shared Prosody Controls")+":"), applyProsody,
					widget.NewLabel(lang.L("Reuse Generated Voice")+":"), saveReferenceButton,
				),
			)),
			widget.NewAccordionItem(lang.L("Generation"), container.NewVBox(
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Model Text Mode")+":"), textModeSelect,
				),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Seed")+":"), seedInput,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Max new tokens")+":"), maxTokensInput,
					),
				),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Sampling")+":"), doSample,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Sub-talker Sampling")+":"), subtalkerDoSample,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Temperature")+":"), temperatureControl,
					widget.NewLabel("Top P:"), topPControl,
					widget.NewLabel("Top K:"), topKControl,
					widget.NewLabel(lang.L("Repetition Penalty")+":"), repetitionControl,
					widget.NewLabel(lang.L("Sub-talker Temperature")+":"), subtalkerTemperatureControl,
					widget.NewLabel("Sub-talker Top P:"), subtalkerTopPControl,
					widget.NewLabel("Sub-talker Top K:"), subtalkerTopKControl,
				),
			)),
			widget.NewAccordionItem(lang.L("Streaming"), container.NewVBox(
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Streaming Mode")+":"), streamingModeSelect,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Playback Buffer")+":"), streamingBufferModeSelect,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Adaptive Safety Margin (ms)")+":"), bufferSafetyControl,
				),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Codec Frames per Packet (~80 ms each)")+":"), codecFramesControl,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Decoder Left Context (frames)")+":"), decoderContextFramesControl,
					),
				),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Lookahead Batch Size")+":"), lookaheadBatchSizeControl,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Lookahead Codec Frames per Packet")+":"), lookaheadCodecFramesControl,
					),
				),
				container.New(layout.NewFormLayout(),
					widget.NewLabel(lang.L("Lookahead Target Characters")+":"), lookaheadCharactersControl,
					widget.NewLabel(lang.L("Lookahead Join Pause (ms)")+":"), lookaheadPauseControl,
					widget.NewLabel(lang.L("Fallback Segment Characters")+":"), segmentCharactersControl,
					widget.NewLabel(lang.L("Fallback Segment Pause (ms)")+":"), segmentPauseControl,
				),
			)),
		),
	)
}
