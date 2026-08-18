package Messages

import (
	"testing"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
)

func TestTTSModelDisplayValuePreservesCanonicalGroupAndModel(t *testing.T) {
	original := Fields.TtsModelSelectionValues
	t.Cleanup(func() { Fields.TtsModelSelectionValues = original })

	display := "Qwen3-TTS-12Hz-1.7B-VoiceDesign (Voice design)"
	canonical := []string{"Voice design", "Qwen3-TTS-12Hz-1.7B-VoiceDesign"}
	Fields.TtsModelSelectionValues = map[string][]string{display: canonical}

	if got := ttsModelDisplayValue(canonical); got != display {
		t.Fatalf("canonical Qwen selection resolved to %q, want %q", got, display)
	}
}

func TestTTSModelDisplayValueRejectsIncompleteLegacyValue(t *testing.T) {
	if got := ttsModelDisplayValue([]string{"Qwen3-TTS-12Hz-0.6B-Base"}); got != "" {
		t.Fatalf("incomplete model selection resolved unexpectedly: %q", got)
	}
}

func TestHasSelectableTTSVoice(t *testing.T) {
	if hasSelectableTTSVoice(nil) {
		t.Fatal("nil voice options should not be selectable")
	}
	if hasSelectableTTSVoice([]CustomWidget.TextValueOption{{Text: "", Value: ""}}) {
		t.Fatal("the empty placeholder should not be selectable")
	}
	if !hasSelectableTTSVoice([]CustomWidget.TextValueOption{
		{Text: "Vivian", Value: "Vivian"},
		{Text: "Ryan", Value: "Ryan"},
	}) {
		t.Fatal("a multi-voice Qwen preset list should be selectable")
	}
}
