package Hardwareinfo

import (
	"math"
	"testing"
)

func TestAudioCppMemoryEstimatesAndGGUFPrecisions(t *testing.T) {
	tests := []struct {
		option ProfileAIModelOption
		want   float64
	}{
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "audio_cpp", AIModelSize: "Qwen3-ASR-0.6B-GGUF", Precision: PrecisionMemoryFactor("q8_0")}, 1250},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "audio_cpp", AIModelSize: "Qwen3-ASR-1.7B-GGUF", Precision: PrecisionMemoryFactor("f16")}, 5400},
		{ProfileAIModelOption{AIModel: "ttsType", AIModelType: "audio_cpp", AIModelSize: "Supertonic-3-GGUF", Precision: Float32}, 900},
		{ProfileAIModelOption{AIModel: "Whisper", AIModelType: "audio_cpp", AIModelSize: "Voxtral-Mini-4B-Realtime-2602-GGUF", Precision: PrecisionMemoryFactor("q4_k")}, 3000},
		{ProfileAIModelOption{AIModel: "ttsType", AIModelType: "audio_cpp", AIModelSize: "OmniVoice-GGUF", Precision: PrecisionMemoryFactor("q8_0")}, 2000},
		{ProfileAIModelOption{AIModel: "ttsType", AIModelType: "audio_cpp", AIModelSize: "OmniVoice-GGUF", Precision: PrecisionMemoryFactor("f16")}, 4000},
		{ProfileAIModelOption{AIModel: "ttsType", AIModelType: "audio_cpp", AIModelSize: "OmniVoice-GGUF", Precision: PrecisionMemoryFactor("bf16")}, 4000},
	}
	for _, test := range tests {
		got, found := knownModelMemory(test.option)
		if !found {
			t.Fatalf("no audio.cpp estimate found for %#v", test.option)
		}
		if math.Abs(got-test.want) > 0.001 {
			t.Errorf("estimate for %#v = %v, want %v", test.option, got, test.want)
		}
	}
}
