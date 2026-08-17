package Hardwareinfo

import (
	"math"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNormalizeModelSizePreservesNonWhisperDots(t *testing.T) {
	tests := []struct {
		modelType string
		size      string
		want      string
	}{
		{"qwen3_asr", "Qwen3-ASR-0.6B-hf", "qwen3-asr-0.6b-hf"},
		{"milmmt", "MiLMMT-46-12B-v1.0", "milmmt-46-12b-v1.0"},
		{"faster_whisper", "large-distilled-v3.5.en", "large-distilled"},
		{"original_whisper", "medium.en", "medium"},
	}
	for _, test := range tests {
		if got := normalizeModelSize(test.modelType, test.size); got != test.want {
			t.Errorf("normalizeModelSize(%q, %q) = %q, want %q", test.modelType, test.size, got, test.want)
		}
	}
}

func TestKnownModelMemoryMatchesVersionedAndMixedCaseModels(t *testing.T) {
	tests := []struct {
		option ProfileAIModelOption
		want   float64
	}{
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "qwen3_asr", AIModelSize: "Qwen3-ASR-0.6B-hf", Precision: Float16}, 2100},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "qwen3_asr", AIModelSize: "Qwen3-ASR-1.7B-hf", Precision: Float16}, 5000},
		{ProfileAIModelOption{AIModel: "TxtTranslator", AIModelType: "milmmt", AIModelSize: "MiLMMT-46-1B-v1.0", Precision: Float16}, 2600},
		{ProfileAIModelOption{AIModel: "TxtTranslator", AIModelType: "milmmt", AIModelSize: "MiLMMT-46-4B-v1.0", Precision: Bit8}, 5000},
		{ProfileAIModelOption{AIModel: "TxtTranslator", AIModelType: "milmmt", AIModelSize: "MiLMMT-46-12B-v1.0", Precision: Float16}, 28000},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "voxtral", AIModelSize: "voxtral-mini-3b-2507", Precision: Float32}, 18852},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "seamless_m4t", AIModelSize: "large-v2", Precision: Float32}, 10518},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "speech_t5", AIModelSize: "stale-disabled-selection", Precision: Float32}, 927},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "wav2vec_bert", AIModelSize: "stale-disabled-selection", Precision: Float32}, 2989},
	}
	for _, test := range tests {
		got, found := knownModelMemory(test.option)
		if !found {
			t.Fatalf("no estimate found for %#v", test.option)
		}
		if math.Abs(got-test.want) > 0.001 {
			t.Errorf("estimate for %#v = %v, want %v", test.option, got, test.want)
		}
	}
}

func TestEveryEstimateTableEntryCanBeResolved(t *testing.T) {
	for _, model := range Models {
		option := ProfileAIModelOption{
			AIModel:     model.BaseName,
			AIModelType: model.ModelType,
			AIModelSize: model.ModelSize,
			Precision:   Float32,
		}
		got, found := knownModelMemory(option)
		if !found {
			t.Errorf("estimate table entry cannot be resolved: %#v", model)
			continue
		}
		if math.Abs(got-model.Float32PrecisionMemoryUsage) > 0.001 {
			t.Errorf("estimate table entry %#v resolves to %v", model, got)
		}
	}
}

func TestCalculateMemoryConsumptionCalculatesFirstCompleteSnapshot(t *testing.T) {
	testApp := fyneTest.NewApp()
	t.Cleanup(testApp.Quit)

	originalOptions := AllProfileAIModelOptions
	originalCalculating := calculating
	t.Cleanup(func() {
		AllProfileAIModelOptions = originalOptions
		calculating = originalCalculating
	})
	AllProfileAIModelOptions = nil
	calculating = false

	cpuBar := widget.NewProgressBar()
	gpuBar := widget.NewProgressBar()
	option := ProfileAIModelOption{
		AIModel:     "Whisper",
		AIModelType: "qwen3_asr",
		AIModelSize: "Qwen3-ASR-0.6B-hf",
		Precision:   Float16,
		Device:      "cuda",
	}
	option.CalculateMemoryConsumption(cpuBar, gpuBar, 32000)

	if len(AllProfileAIModelOptions) != 1 {
		t.Fatalf("expected one estimator entry, got %d", len(AllProfileAIModelOptions))
	}
	if got := AllProfileAIModelOptions[0].MemoryConsumption; math.Abs(got-2100) > 0.001 {
		t.Fatalf("first complete snapshot estimate = %v, want 2100", got)
	}

	ProfileAIModelOption{}.CalculateMemoryConsumption(cpuBar, gpuBar, 32000)
	if len(AllProfileAIModelOptions) != 1 {
		t.Fatalf("refresh-only call created a synthetic entry: %#v", AllProfileAIModelOptions)
	}
}

func TestPrecisionMemoryFactor(t *testing.T) {
	tests := map[string]float64{
		"float32":  Float32,
		"bfloat16": Float16,
		"int8":     Bit8,
		"8bit":     Bit8,
		"4bit":     Bit4,
		"auto":     Float32,
	}
	for precision, want := range tests {
		if got := PrecisionMemoryFactor(precision); got != want {
			t.Errorf("PrecisionMemoryFactor(%q) = %v, want %v", precision, got, want)
		}
	}
}
