package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2/lang"
	"strings"
)

func additionalAudioCppTTSSchema(model string) (audioCppTTSSchema, bool) {
	schema := audioCppTTSSchema{title: model}
	fields := []audioCppTTSField{}
	switch model {
	case "Kokoro-82M-GGUF":
		schema.family = "kokoro_tts"
		field := audioCppTTSLanguage("language", "auto", []string{"auto", "en-us", "en-gb", "es", "fr-fr", "hi", "it", "ja", "pt-br", "zh"}, false)
		field.label = lang.L("Language")
		fields = append(fields, field)
	case "Breeze-TTS-2-GGUF":
		schema.family = "breeze_tts"
		fields = append(fields, audioCppTTSParagraph("instruction", lang.L("Instruction"), "Speak clearly and naturally."), audioCppTTSParagraph("reference_text", lang.L("Reference text"), ""))
	case "CosyVoice3-GGUF":
		schema.family = "cosyvoice3"
		fields = append(fields, audioCppTTSChoice("template_name", lang.L("Template"), "cross_lingual", audioCppTTSOptions("cross_lingual", "zero_shot", "instruct")...), audioCppTTSParagraph("reference_text", lang.L("Reference text"), ""), audioCppTTSParagraph("instruction", lang.L("Instruction"), ""), audioCppTTSIntSlider("num_inference_steps", lang.L("Quality steps"), 10, 1, 50, 1))
	case "Audio8-TTS-Preview-0.6B-GGUF":
		schema.family = "audio8_tts"
		fields = append(fields, audioCppTTSParagraph("reference_text", lang.L("Reference text"), ""), audioCppTTSFloatSlider("temperature", lang.L("Creativity"), 0.7, 0.1, 2, 0.05, 2))
	case "Chatterbox-Turbo-GGUF":
		schema.family = "chatterbox_turbo"
	default:
		if !strings.HasPrefix(model, "sanoTTS-") || !strings.HasSuffix(model, "-GGUF") {
			return audioCppTTSSchema{}, false
		}
		schema.family = "sanotts"
		fields = append(fields, audioCppTTSFloatSlider("speaking_rate", lang.L("Speed"), 1, 0.5, 2, 0.05, 2))
	}
	if len(fields) > 0 {
		schema.groups = []audioCppTTSGroup{{title: lang.L("Settings"), columns: 2, fields: fields}}
	}
	return schema, true
}
