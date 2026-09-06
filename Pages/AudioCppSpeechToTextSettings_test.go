package Pages

import (
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func audioCppSTTFields(schema audioCppSTTSchema) []audioCppSTTField {
	var fields []audioCppSTTField
	for _, group := range schema.groups {
		fields = append(fields, group.fields...)
	}
	return fields
}

func audioCppSTTFieldKeys(schema audioCppSTTSchema) []string {
	fields := audioCppSTTFields(schema)
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, field.key)
	}
	return keys
}

func TestAudioCppSTTSchemasExposeCuratedModelFeatures(t *testing.T) {
	tests := []struct {
		model  string
		family string
		keys   []string
	}{
		{"Qwen3-ASR-0.6B-GGUF", "qwen3_asr", []string{"mode", "audio_chunk_mode", "audio_chunk_seconds"}},
		{"Qwen3-ASR-1.7B-GGUF", "qwen3_asr", []string{"mode", "audio_chunk_mode", "audio_chunk_seconds"}},
		{"Nemotron-3.5-ASR-Streaming-0.6B-GGUF", "nemotron_asr", []string{"mode", "lookahead_tokens", "keep_language_tags"}},
		{"VibeVoice-ASR-GGUF", "vibevoice_asr", []string{"audio_chunk_mode", "audio_chunk_seconds"}},
		{"Voxtral-Mini-4B-Realtime-2602-GGUF", "voxtral_realtime", []string{"mode"}},
		{"Audio8-ASR-0.1B-GGUF", "audio8_asr", []string{}},
		{"Kroko-ASR-English-64L-GGUF", "kroko_asr", []string{"mode", "decoding_method", "num_beams", "hotwords", "hotwords_score", "enable_endpoint", "rule1_min_trailing_silence_sec", "rule2_min_trailing_silence_sec", "rule3_min_utterance_length_sec"}},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			schema, ok := audioCppSTTSchemaForModel(test.model)
			if !ok || schema.family != test.family {
				t.Fatalf("schema family = %q/%v, want %q/true", schema.family, ok, test.family)
			}
			if got := audioCppSTTFieldKeys(schema); !reflect.DeepEqual(got, test.keys) {
				t.Fatalf("visible settings = %#v, want %#v", got, test.keys)
			}
		})
	}
}

func TestAudioCppSTTHidesImplementationAndDuplicatePrecisionSettings(t *testing.T) {
	models := []string{
		"Qwen3-ASR-0.6B-GGUF", "Nemotron-3.5-ASR-Streaming-0.6B-GGUF", "VibeVoice-ASR-GGUF",
		"Voxtral-Mini-4B-Realtime-2602-GGUF", "Audio8-ASR-0.1B-GGUF", "Kroko-ASR-English-64L-GGUF",
	}
	for _, model := range models {
		schema, _ := audioCppSTTSchemaForModel(model)
		for _, field := range audioCppSTTFields(schema) {
			for _, fragment := range []string{"weight_type", "arena", "context_mb", "cache_steps", "vad_model_path", "max_tokens", "max_new_tokens", "seed"} {
				if strings.Contains(field.key, fragment) {
					t.Fatalf("%s exposes implementation setting %q", model, field.key)
				}
			}
		}
	}
}

func TestAudioCppSTTUsesSlidersAndCompactTwoColumnGroups(t *testing.T) {
	models := []string{
		"Qwen3-ASR-0.6B-GGUF", "Nemotron-3.5-ASR-Streaming-0.6B-GGUF", "VibeVoice-ASR-GGUF",
		"Voxtral-Mini-4B-Realtime-2602-GGUF", "Kroko-ASR-English-64L-GGUF",
	}
	for _, model := range models {
		schema, _ := audioCppSTTSchemaForModel(model)
		for _, group := range schema.groups {
			rows := len(group.fields)
			if group.columns > 1 {
				rows = (rows + group.columns - 1) / group.columns
			}
			if rows > 2 {
				t.Fatalf("%s group %q uses %d visible rows; want at most 2", model, group.title, rows)
			}
			for _, field := range group.fields {
				switch field.fallback.(type) {
				case int, float64:
					if field.kind != audioCppSTTSlider {
						t.Fatalf("%s numeric setting %q does not use a slider", model, field.key)
					}
				}
			}
		}
	}
}

func TestOfflineOnlyAudioCppSTTModelsDoNotShowDisabledModeSelectors(t *testing.T) {
	for _, model := range []string{"VibeVoice-ASR-GGUF", "Audio8-ASR-0.1B-GGUF"} {
		schema, _ := audioCppSTTSchemaForModel(model)
		for _, field := range audioCppSTTFields(schema) {
			if field.key == "mode" {
				t.Fatalf("%s exposes a mode selector even though only offline mode is supported", model)
			}
		}
	}
}

func TestAudioCppSTTPanelContainsOnlyControlsAndReset(t *testing.T) {
	panel, ok := buildAudioCppSTTSettingsForModel("Qwen3-ASR-0.6B-GGUF", nil).(*fyne.Container)
	if !ok {
		t.Fatal("audio.cpp STT panel is not a container")
	}
	if len(panel.Objects) != 2 {
		t.Fatalf("audio.cpp STT panel has %d top-level rows, want only accordion and reset", len(panel.Objects))
	}

	noSettings, ok := buildAudioCppSTTSettingsForModel("Audio8-ASR-0.1B-GGUF", nil).(*fyne.Container)
	if !ok || len(noSettings.Objects) != 0 {
		t.Fatalf("Audio8 no-settings panel = %#v, want an empty container", noSettings)
	}
}
