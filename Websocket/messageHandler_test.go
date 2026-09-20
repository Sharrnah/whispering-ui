package Websocket

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	fynetest "fyne.io/fyne/v2/test"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
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

func TestAudioRoutesResultCommitsBackendConfigurationBeforeCallback(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	previousConfig := Settings.Config
	previousCallback := Fields.AudioRoutesUpdateResult
	t.Cleanup(func() {
		Settings.Config = previousConfig
		Fields.AudioRoutesUpdateResult = previousCallback
	})

	emptyMain := []string{}
	data, err := json.Marshal(map[string]interface{}{
		"request_id": "route-1",
		"success":    true,
		"routes": []Settings.AdditionalAudioRoute{{
			ID:      "game",
			Name:    "Game chat",
			Enabled: true,
		}},
		"main_audio_plugins": emptyMain,
	})
	if err != nil {
		t.Fatal(err)
	}

	completed := make(chan struct{}, 1)
	Fields.AudioRoutesUpdateResult = func(requestID string, success bool, errorMessage string) {
		if requestID != "route-1" || !success || errorMessage != "" {
			t.Errorf("unexpected callback: %q, %v, %q", requestID, success, errorMessage)
		}
		if len(Settings.Config.Additional_audio_routes) != 1 ||
			Settings.Config.Additional_audio_routes[0].ID != "game" {
			t.Errorf("backend routes were not committed before callback: %#v", Settings.Config.Additional_audio_routes)
		}
		if Settings.Config.Main_audio_plugins == nil || len(*Settings.Config.Main_audio_plugins) != 0 {
			t.Errorf("backend main plugin allowlist was not committed: %#v", Settings.Config.Main_audio_plugins)
		}
		completed <- struct{}{}
	}

	message := MessageStruct{Type: "audio_routes_update_result", Data: data}
	message.HandleReceiveMessage()
	select {
	case <-completed:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for audio route result callback")
	}
}
