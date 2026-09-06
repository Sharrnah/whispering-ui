package SpecialTextToSpeechSettings

import (
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func audioCppTTSFields(schema audioCppTTSSchema) []audioCppTTSField {
	var fields []audioCppTTSField
	for _, group := range schema.groups {
		fields = append(fields, group.fields...)
	}
	return fields
}

func audioCppTTSFieldKeys(schema audioCppTTSSchema) []string {
	fields := audioCppTTSFields(schema)
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		keys = append(keys, field.key)
	}
	return keys
}

func TestAudioCppTTSSchemasExposeCuratedModelFeatures(t *testing.T) {
	tests := []struct {
		model  string
		family string
		keys   []string
	}{
		{"Supertonic-3-GGUF", "supertonic", []string{"language", "num_inference_steps"}},
		{"Confucius4-TTS-GGUF", "confucius4_tts", []string{"language", "num_beams", "num_inference_steps", "guidance_scale", "temperature"}},
		{"DotTTS-SOAR-GGUF", "dots_tts", []string{"language", "template_name", "reference_text", "instruction", "num_inference_steps", "guidance_scale", "speaker_scale", "sampler_mode"}},
		{"DotTTS-MeanFlow-GGUF", "dots_tts", []string{"language", "template_name", "reference_text", "instruction", "num_inference_steps", "guidance_scale", "speaker_scale", "sampler_mode"}},
		{"DotTTS-Edit-GGUF", "dots_tts_edit", []string{"source_audio", "source_text", "target_text", "instruction", "language", "use_xvector", "num_inference_steps", "guidance_scale", "speaker_scale", "sampler_mode"}},
		{"IndexTTS2-GGUF", "index_tts2", []string{"language", "duration_factor", "emotion_audio", "emotion_text", "use_emotion_text", "use_random_emotion", "emotion_alpha", "emotion_happy", "emotion_angry", "emotion_sad", "emotion_afraid", "emotion_disgusted", "emotion_melancholic", "emotion_surprised", "emotion_calm"}},
		{"IndexTTS2.5-GGUF", "index_tts2", []string{"language", "duration_factor", "emotion_audio", "emotion_text", "use_emotion_text", "use_random_emotion", "emotion_alpha", "emotion_happy", "emotion_angry", "emotion_sad", "emotion_afraid", "emotion_disgusted", "emotion_melancholic", "emotion_surprised", "emotion_calm"}},
		{"MagpieTTS-Multilingual-357M-GGUF", "magpie_tts", []string{"language", "temperature", "top_k", "guidance_scale"}},
		{"OmniVoice-GGUF", "omnivoice", []string{"language", "denoise", "reference_text", "voice_instruction", "num_inference_steps", "guidance_scale", "speed", "duration"}},
		{"VoxCPM1-0.5B-GGUF", "voxcpm1", []string{"reference_text", "num_inference_steps", "guidance_scale"}},
		{"VoxCPM2-GGUF", "voxcpm2", []string{"reference_text", "voice_instruction", "num_inference_steps", "guidance_scale"}},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			schema, ok := audioCppTTSSchemaForModel(test.model)
			if !ok || schema.family != test.family {
				t.Fatalf("schema family = %q/%v, want %q/true", schema.family, ok, test.family)
			}
			if got := audioCppTTSFieldKeys(schema); !reflect.DeepEqual(got, test.keys) {
				t.Fatalf("visible settings = %#v, want %#v", got, test.keys)
			}
		})
	}
}

func TestAudioCppTTSHidesImplementationAndDuplicatePrecisionSettings(t *testing.T) {
	models := []string{
		"Supertonic-3-GGUF", "Confucius4-TTS-GGUF", "DotTTS-SOAR-GGUF", "DotTTS-Edit-GGUF",
		"IndexTTS2.5-GGUF", "MagpieTTS-Multilingual-357M-GGUF", "OmniVoice-GGUF", "VoxCPM2-GGUF",
	}
	for _, model := range models {
		schema, _ := audioCppTTSSchemaForModel(model)
		for _, field := range audioCppTTSFields(schema) {
			for _, fragment := range []string{"weight_type", "arena", "context_mb", "cache_slots", "max_tokens", "text_chunk", "seed"} {
				if strings.Contains(field.key, fragment) {
					t.Fatalf("%s exposes implementation setting %q", model, field.key)
				}
			}
		}
	}
}

func TestAudioCppTTSUsesPurposeBuiltWidgetsAndCompactGroups(t *testing.T) {
	models := []string{
		"Supertonic-3-GGUF", "Confucius4-TTS-GGUF", "DotTTS-SOAR-GGUF", "DotTTS-Edit-GGUF",
		"IndexTTS2.5-GGUF", "MagpieTTS-Multilingual-357M-GGUF", "OmniVoice-GGUF", "VoxCPM2-GGUF",
	}
	for _, model := range models {
		schema, _ := audioCppTTSSchemaForModel(model)
		for _, group := range schema.groups {
			rows := len(group.fields)
			if group.columns > 1 {
				rows = (rows + group.columns - 1) / group.columns
			}
			if rows > 4 {
				t.Fatalf("%s group %q uses %d visible rows; want at most 4", model, group.title, rows)
			}
			for _, field := range group.fields {
				if field.key == "language" && field.kind != audioCppTTSSearchSelect {
					t.Fatalf("%s language uses kind %v, want searchable select", model, field.kind)
				}
				switch field.fallback.(type) {
				case int, float64:
					if field.kind != audioCppTTSSlider {
						t.Fatalf("%s numeric setting %q does not use a slider", model, field.key)
					}
				}
				if (field.key == "source_audio" || field.key == "emotion_audio") && field.kind != audioCppTTSAudioFile {
					t.Fatalf("%s audio path %q does not use a file picker", model, field.key)
				}
			}
		}
	}
}

func TestAudioCppOmniVoiceLanguageAllowsFilteredCustomValues(t *testing.T) {
	schema, _ := audioCppTTSSchemaForModel("OmniVoice-GGUF")
	for _, field := range audioCppTTSFields(schema) {
		if field.key != "language" {
			continue
		}
		if field.kind != audioCppTTSSearchSelect || !field.allowCustom || len(field.options) < 30 {
			t.Fatalf("OmniVoice language field = %#v, want searchable custom field with useful suggestions", field)
		}
		if value, ok := audioCppTTSOptionValue(field.options, "German (de)"); !ok || value != "de" {
			t.Fatalf("German display value resolved to %q/%v, want de/true", value, ok)
		}
		return
	}
	t.Fatal("OmniVoice language field missing")
}

func TestAudioCppVoiceSelectionIsNotDuplicatedInAdvancedSettings(t *testing.T) {
	for _, model := range []string{"Supertonic-3-GGUF", "MagpieTTS-Multilingual-357M-GGUF"} {
		schema, _ := audioCppTTSSchemaForModel(model)
		for _, field := range audioCppTTSFields(schema) {
			if field.key == "voice_id" || field.key == "speaking_rate" {
				t.Fatalf("%s duplicates normal profile control %q", model, field.key)
			}
		}
	}
}

func TestAudioCppTTSPanelContainsOnlyControlsAndReset(t *testing.T) {
	panel, ok := buildAudioCppTTSSettingsForModel("Supertonic-3-GGUF", nil).(*fyne.Container)
	if !ok {
		t.Fatal("audio.cpp TTS panel is not a container")
	}
	if len(panel.Objects) != 2 {
		t.Fatalf("audio.cpp TTS panel has %d top-level rows, want only accordion and reset", len(panel.Objects))
	}
}
