package Settings

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAdditionalAudioRouteAutoLanguageIsExplicitlyEncoded(t *testing.T) {
	route := AdditionalAudioRoute{Current_language: ""}

	encodedYAML, err := yaml.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encodedYAML), `current_language: ""`) {
		t.Fatalf("Auto language was omitted from YAML:\n%s", encodedYAML)
	}

	encodedJSON, err := json.Marshal(route)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encodedJSON), `"current_language":""`) {
		t.Fatalf("Auto language was omitted from JSON:\n%s", encodedJSON)
	}
}

func TestNilMainAudioPluginListEncodesLegacyAllPluginsAsNull(t *testing.T) {
	encoded, err := yaml.Marshal(Conf{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), "main_audio_plugins: null") {
		t.Fatalf("nil plugin routing must encode the legacy all-plugin mode, got:\n%s", encoded)
	}
	if !strings.Contains(string(encoded), "additional_audio_routes: []") {
		t.Fatalf("an empty profile must clear routes during profile merges, got:\n%s", encoded)
	}
}

func TestExplicitEmptyMainAudioPluginListSurvivesYamlRoundTrip(t *testing.T) {
	empty := []string{}
	configuration := Conf{
		Main_audio_plugins: &empty,
		Additional_audio_routes: []AdditionalAudioRoute{{
			ID:                                   "game",
			Name:                                 "Game audio",
			Enabled:                              true,
			Audio_api:                            "WASAPI",
			Audio_input_process:                  "game.exe",
			Audio_input_process_id:               42,
			Stt_enabled:                          true,
			Realtime_frequency_time:              0.4,
			Silence_cutting_enabled:              false,
			Denoise_audio:                        "noise_reduce",
			Denoise_strength:                     0.65,
			Vad_smart_turn_enabled:               true,
			Vad_smart_turn_min_length:            1.5,
			Vad_smart_turn_probability_threshold: 0.7,
			Vad_smart_turn_pause_length:          0.25,
			Websocket_enabled:                    true,
			Osc_enabled:                          true,
			Osc_typing_indicator:                 true,
			Osc_chat_notification:                true,
			Osc_chat_prefix:                      "[game] ",
			Plugins:                              []string{"SubtitlePlugin"},
		}},
	}

	encoded, err := yaml.Marshal(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), "main_audio_plugins: []") {
		t.Fatalf("explicit empty allowlist was not encoded, got:\n%s", encoded)
	}

	var decoded Conf
	if err := yaml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Main_audio_plugins == nil || len(*decoded.Main_audio_plugins) != 0 {
		t.Fatalf("empty allowlist did not survive: %#v", decoded.Main_audio_plugins)
	}
	if len(decoded.Additional_audio_routes) != 1 {
		t.Fatalf("route count = %d, want 1", len(decoded.Additional_audio_routes))
	}
	route := decoded.Additional_audio_routes[0]
	if route.Audio_input_process != "game.exe" || route.Audio_input_process_id != 42 {
		t.Fatalf("application capture did not survive: %#v", route)
	}
	if len(route.Plugins) != 1 || route.Plugins[0] != "SubtitlePlugin" {
		t.Fatalf("plugin routing did not survive: %#v", route.Plugins)
	}
	if route.Realtime_frequency_time != 0.4 ||
		route.Denoise_audio != "noise_reduce" ||
		route.Denoise_strength != 0.65 ||
		!route.Vad_smart_turn_enabled ||
		route.Vad_smart_turn_probability_threshold != 0.7 {
		t.Fatalf("audio-processing settings did not survive: %#v", route)
	}
	if !route.Osc_enabled || !route.Osc_typing_indicator || !route.Osc_chat_notification || route.Osc_chat_prefix != "[game] " {
		t.Fatalf("OSC route settings did not survive: %#v", route)
	}
}

func TestLegacyProfileClearsRoutesAndExplicitPluginRoutingFromPreviousProfile(t *testing.T) {
	empty := []string{}
	previous := Conf{
		Additional_audio_routes: []AdditionalAudioRoute{{ID: "game", Name: "Game audio"}},
		Main_audio_plugins:      &empty,
	}
	legacyProfile := Conf{}

	merged := MergeSettings(previous, legacyProfile)
	if len(merged.Additional_audio_routes) != 0 {
		t.Fatalf("routes leaked across profiles: %#v", merged.Additional_audio_routes)
	}
	if merged.Main_audio_plugins != nil {
		t.Fatalf("plugin allowlist leaked across profiles: %#v", merged.Main_audio_plugins)
	}
}
