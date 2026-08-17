package ProfileForm

import "testing"

func TestMiLMMTOptions(t *testing.T) {
	typeFound := false
	for _, option := range TXTTypeOptions() {
		if option.Value == "milmmt" {
			typeFound = true
			break
		}
	}
	if !typeFound {
		t.Fatal("MiLMMT is missing from the text translation type options")
	}

	models, defaultIndex, enabled := TXTSizeOptions("milmmt")
	if !enabled || defaultIndex != 0 || len(models) != 4 {
		t.Fatalf(
			"unexpected MiLMMT model options: enabled=%v default=%d count=%d",
			enabled,
			defaultIndex,
			len(models),
		)
	}
	wantedModels := []string{
		"MiLMMT-46-1B-v1.0",
		"MiLMMT-46-4B-v1.0",
		"MiLMMT-46-12B-v1.0",
		"custom",
	}
	for index, wanted := range wantedModels {
		if models[index].Value != wanted {
			t.Fatalf("unexpected MiLMMT model values: %#v", models)
		}
	}

	precisions, enabled := TXTPrecisionOptions("milmmt")
	if !enabled {
		t.Fatal("MiLMMT precision selection should be enabled")
	}
	if len(precisions) != 3 {
		t.Fatalf("unexpected MiLMMT precision count: %d", len(precisions))
	}
	wantedPrecisions := map[string]bool{
		"float32":  false,
		"bfloat16": false,
		"8bit":     false,
	}
	for _, precision := range precisions {
		if _, ok := wantedPrecisions[precision.Value]; ok {
			wantedPrecisions[precision.Value] = true
		}
	}
	for precision, found := range wantedPrecisions {
		if !found {
			t.Errorf("MiLMMT precision %q is missing", precision)
		}
	}
	for _, precision := range precisions {
		if precision.Value == "float16" || precision.Value == "4bit" {
			t.Errorf("MiLMMT precision %q must remain hidden because real inference returned empty output", precision.Value)
		}
	}
}
