package SpecialTextToSpeechSettings

import "testing"

func TestQwen3TTSInstructionSupported(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"Qwen3-TTS-12Hz-1.7B-CustomVoice", true},
		{"Qwen3-TTS-12Hz-1.7B-VoiceDesign", true},
		{"Qwen3-TTS-12Hz-0.6B-CustomVoice", false},
		{"Qwen3-TTS-12Hz-0.6B-Base", false},
		{"Qwen3-TTS-12Hz-1.7B-Base", false},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			if got := qwen3TTSInstructionSupported(test.model); got != test.want {
				t.Fatalf("qwen3TTSInstructionSupported(%q) = %v, want %v", test.model, got, test.want)
			}
		})
	}
}
