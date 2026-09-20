package ProfileForm

import "fyne.io/fyne/v2/lang"

// Published GGUF packages verified against audio.cpp v0.8.1. Keep identifiers
// and precisions synchronized with Models/audio_cpp_catalog.py in the backend.
type audioCppModel struct {
	name       string
	group      string
	precisions []string
}

var audioCppAdditionalSTTModels = []audioCppModel{
	{"Canary-180M-Flash-GGUF", "", []string{"f32", "q8_0"}},
	{"Cohere-Transcribe-GGUF", "", []string{"bf16", "q8_0", "q4_0"}},
	{"Moonshine-Streaming-Tiny-GGUF", "", []string{"q8_0"}},
	{"Moonshine-Streaming-Small-GGUF", "", []string{"q8_0"}},
	{"Moonshine-Streaming-Medium-GGUF", "", []string{"q8_0"}},
	{"Niagara-19M-Batch-English-GGUF", "", []string{"f32"}},
	{"Niagara-38M-Batch-English-GGUF", "", []string{"f32"}},
	{"MOSS-Transcribe-Diarize-GGUF", "", []string{"bf16", "q8_0", "q4_k"}},
	{"VibeVoice-ASR-Streaming-7B-GGUF", "", []string{"q8_0", "bf16", "q4_k"}},
}

var audioCppAdditionalTTSModels = []audioCppModel{
	{"Breeze-TTS-2-GGUF", "Voice cloning", []string{"q8_0", "bf16"}},
	{"Chatterbox-Turbo-GGUF", "Preset voices", []string{"q8_0"}},
	{"CosyVoice3-GGUF", "Voice cloning", []string{"q8_0", "f32"}},
	{"Kokoro-82M-GGUF", "Preset voices", []string{"q8_0", "bf16"}},
	{"Audio8-TTS-Preview-0.6B-GGUF", "Voice cloning", []string{"q8_0"}},
	{"sanoTTS-heart-nano-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-heart-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-amy-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-hfc-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-kristin-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-vi-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-id-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-cs-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-de-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-es-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-fr-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-it-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-pt-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-ro-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-ru-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-tr-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-ne-GGUF", "Preset voices", []string{"orig"}},
	{"sanoTTS-hi-GGUF", "Preset voices", []string{"orig"}},
}

func additionalAudioCppModel(models []audioCppModel, name string) (audioCppModel, bool) {
	for _, model := range models {
		if model.name == name {
			return model, true
		}
	}
	return audioCppModel{}, false
}

func audioCppPublishedPrecisions(model audioCppModel) ([]TVO, bool) {
	options := make([]TVO, 0, len(model.precisions))
	for _, precision := range model.precisions {
		options = append(options, TVO{Text: precision + " GGUF", Value: precision})
	}
	return options, len(options) > 1
}

func audioCppAdditionalSTTOptions() []TVO {
	options := make([]TVO, 0, len(audioCppAdditionalSTTModels))
	for _, model := range audioCppAdditionalSTTModels {
		options = append(options, TVO{Text: lang.L(model.name) + " (" + audioCppAdditionalLanguages[model.name] + ")", Value: model.name})
	}
	return options
}

var audioCppAdditionalLanguages = map[string]string{
	"Canary-180M-Flash-GGUF":          "en, de, es, fr",
	"Cohere-Transcribe-GGUF":          "en, fr, de, es, it, pt, nl, pl, el, ar, ja, zh, vi, ko",
	"Moonshine-Streaming-Tiny-GGUF":   "en",
	"Moonshine-Streaming-Small-GGUF":  "en",
	"Moonshine-Streaming-Medium-GGUF": "en",
	"Niagara-19M-Batch-English-GGUF":  "en",
	"Niagara-38M-Batch-English-GGUF":  "en",
	"MOSS-Transcribe-Diarize-GGUF":    "auto",
	"VibeVoice-ASR-Streaming-7B-GGUF": "en, zh, es, pt, de, ja, ko, fr, ru, it",
	"Breeze-TTS-2-GGUF":               "zh, en",
	"Chatterbox-Turbo-GGUF":           "en",
	"CosyVoice3-GGUF":                 "zh, en, ja, ko, de, es, fr, it, ru, yue",
	"Kokoro-82M-GGUF":                 "en-us, en-gb, es, fr-fr, hi, it, ja, pt-br, zh",
	"Audio8-TTS-Preview-0.6B-GGUF":    "yue, zh, nl, en, fr, de, it, ja, ko, pl, es",
	"sanoTTS-heart-nano-GGUF":         "en",
	"sanoTTS-heart-GGUF":              "en",
	"sanoTTS-amy-GGUF":                "en",
	"sanoTTS-hfc-GGUF":                "en",
	"sanoTTS-kristin-GGUF":            "en",
	"sanoTTS-vi-GGUF":                 "vi",
	"sanoTTS-id-GGUF":                 "id",
	"sanoTTS-cs-GGUF":                 "cs",
	"sanoTTS-de-GGUF":                 "de",
	"sanoTTS-es-GGUF":                 "es",
	"sanoTTS-fr-GGUF":                 "fr",
	"sanoTTS-it-GGUF":                 "it",
	"sanoTTS-pt-GGUF":                 "pt",
	"sanoTTS-ro-GGUF":                 "ro",
	"sanoTTS-ru-GGUF":                 "ru",
	"sanoTTS-tr-GGUF":                 "tr",
	"sanoTTS-ne-GGUF":                 "ne",
	"sanoTTS-hi-GGUF":                 "hi",
}
