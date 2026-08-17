package ProfileForm

import "testing"

func TestQwen3ASROptions(t *testing.T) {
	typeFound := false
	for _, option := range STTTypeOptions() {
		if option.Value == "qwen3_asr" {
			typeFound = true
			break
		}
	}
	if !typeFound {
		t.Fatal("Qwen3-ASR is missing from the STT type options")
	}

	models, defaultIndex, enabled := STTModelOptions("qwen3_asr")
	if !enabled || defaultIndex != 0 || len(models) != 3 {
		t.Fatalf("unexpected Qwen3-ASR model options: enabled=%v default=%d count=%d", enabled, defaultIndex, len(models))
	}
	if models[0].Value != "Qwen3-ASR-0.6B-hf" || models[1].Value != "Qwen3-ASR-1.7B-hf" {
		t.Fatalf("unexpected Qwen3-ASR model values: %#v", models)
	}

	precisions, enabled := STTPrecisionOptions("qwen3_asr")
	if !enabled {
		t.Fatal("Qwen3-ASR precision selection should be enabled")
	}
	wanted := map[string]bool{"float32": false, "float16": false, "bfloat16": false, "8bit": false, "4bit": false}
	for _, precision := range precisions {
		if _, ok := wanted[precision.Value]; ok {
			wanted[precision.Value] = true
		}
	}
	for precision, found := range wanted {
		if !found {
			t.Errorf("Qwen3-ASR precision %q is missing", precision)
		}
	}
}
