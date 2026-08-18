package Websocket

import (
	"reflect"
	"testing"

	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Websocket/Messages"
)

func TestHandleSendMessagePreservesCanonicalTTSModelSelection(t *testing.T) {
	want := []string{"Voice cloning", "Qwen3-TTS-12Hz-0.6B-Base"}
	message := SendMessageChannel.SendMessageStruct{
		Type:  "setting_change",
		Name:  "tts_model",
		Value: want,
	}

	HandleSendMessage(&message)

	if got, ok := message.Value.([]string); !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("canonical TTS model changed to %#v, want %#v", message.Value, want)
	}
}

func TestNormalizeTTSModelSelectionAcceptsDecodedJSONArray(t *testing.T) {
	want := []string{"Voice design", "Qwen3-TTS-12Hz-1.7B-VoiceDesign"}
	got, ok := normalizeTTSModelSelection([]interface{}{want[0], want[1]})
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded JSON selection normalized to %#v, %v; want %#v, true", got, ok, want)
	}
}

func TestNormalizeTTSModelSelectionConvertsLegacyDisplayValue(t *testing.T) {
	originalValues := Fields.TtsModelSelectionValues
	originalLanguages := Messages.TtsLanguages
	t.Cleanup(func() {
		Fields.TtsModelSelectionValues = originalValues
		Messages.TtsLanguages = originalLanguages
	})

	display := "Qwen3-TTS-12Hz-0.6B-Base (Voice cloning)"
	want := []string{"Voice cloning", "Qwen3-TTS-12Hz-0.6B-Base"}
	Fields.TtsModelSelectionValues = map[string][]string{display: want}

	got, ok := normalizeTTSModelSelection(display)
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy display selection normalized to %#v, %v; want %#v, true", got, ok, want)
	}
}

func TestNormalizeTTSModelSelectionRejectsIncompleteArray(t *testing.T) {
	if got, ok := normalizeTTSModelSelection([]string{"Voice cloning"}); ok || got != nil {
		t.Fatalf("incomplete selection normalized unexpectedly: %#v, %v", got, ok)
	}
}
