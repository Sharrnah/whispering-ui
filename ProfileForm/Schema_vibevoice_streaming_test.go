package ProfileForm

import (
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"testing"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
)

func TestVibeVoiceStreamingSelectors(t *testing.T) {
	types := optionValues(STTTypeOptions())
	if !types["vibevoice_asr"] || types["vibevoice_asr_streaming"] {
		t.Fatal("all VibeVoice checkpoints must share one type selector")
	}
	models, selected, enabled := STTModelOptions("vibevoice_asr")
	if !enabled || selected != 0 || len(models) != 4 || models[0].Value != "VibeVoice-ASR-HF" {
		t.Fatalf("unexpected VibeVoice selectors: %#v", models)
	}
	for _, model := range models[1:] {
		precisions, enabled := STTPrecisionOptionsForModel("vibevoice_asr", model.Value)
		if !enabled || len(precisions) != 2 || precisions[0].Value != "bfloat16" || precisions[1].Value != "float32" {
			t.Fatalf("unexpected streaming precisions for %s: %#v", model.Value, precisions)
		}
	}
	precisions, _ := STTPrecisionOptionsForModel("vibevoice_asr", "VibeVoice-ASR-HF")
	if !optionValues(precisions)["4bit"] || !optionValues(precisions)["float16"] {
		t.Fatal("the original model must retain its precision choices")
	}
}

func TestVibeVoiceProfileRestoreAndModelSwitch(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatal(err)
	}
	for _, input := range []struct{ kind, model, precision, wantModel string }{
		{"vibevoice_asr", "VibeVoice-ASR-HF", "4bit", "VibeVoice-ASR-HF"},
		{"vibevoice_asr", "VibeVoice-ASR-Streaming-1.5B", "bfloat16", "VibeVoice-ASR-Streaming-1.5B"},
		{"vibevoice_asr", "VibeVoice-ASR-Streaming-7B", "float32", "VibeVoice-ASR-Streaming-7B"},
		{"vibevoice_asr_streaming", "VibeVoice-ASR-Streaming-7B", "bfloat16", "VibeVoice-ASR-Streaming-7B"},
		{"vibevoice_asr_streaming", "custom", "bfloat16", "custom-streaming"},
	} {
		controls := &AllProfileControls{}
		coord := &Coordinator{Controls: controls}
		engine := NewFormEngine(controls, coord)
		NewProfileBuilder().BuildAll(engine, nil, nil, nil)
		controls.STTModelSize.OnChanged = func(CustomWidget.TextValueOption) { coord.RefreshSTTPrecisionForModel() }
		conf := Settings.Conf{Stt_type: input.kind, Model: input.model, Whisper_precision: input.precision}
		engine.LoadFromSettings(&conf)
		if selectedValue(controls.STTType) != "vibevoice_asr" || selectedValue(controls.STTModelSize) != input.wantModel || selectedValue(controls.STTPrecision) != input.precision {
			t.Fatalf("restore %v: got %s / %s / %s", input, selectedValue(controls.STTType), selectedValue(controls.STTModelSize), selectedValue(controls.STTPrecision))
		}
		if conf.Stt_type != input.kind || conf.Model != input.model {
			t.Fatal("opening the form changed the caller's snapshot")
		}
		engine.SaveToSettings(&conf)
		if conf.Stt_type != "vibevoice_asr" || conf.Model != input.wantModel {
			t.Fatalf("saved profile must use the unified selection: %v", conf)
		}
		controls.STTModelSize.SetSelected("VibeVoice-ASR-HF")
		controls.STTPrecision.SetSelected("4bit")
		controls.STTModelSize.SetSelected("VibeVoice-ASR-Streaming-1.5B")
		if selectedValue(controls.STTPrecision) != "bfloat16" {
			t.Fatal("switching to streaming must replace unsupported quantization")
		}
		controls.STTModelSize.SetSelected("VibeVoice-ASR-HF")
		if !optionValues(controls.STTPrecision.Options)["4bit"] {
			t.Fatal("switching back must restore the original precision options")
		}
	}
}
