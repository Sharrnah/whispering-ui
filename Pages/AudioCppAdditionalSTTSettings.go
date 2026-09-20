package Pages

import "fyne.io/fyne/v2/lang"

func additionalAudioCppSTTSchema(model string) (audioCppSTTSchema, bool) {
	family := ""
	switch model {
	case "Canary-180M-Flash-GGUF":
		family = "canary_asr"
	case "Cohere-Transcribe-GGUF":
		family = "cohere_asr"
	case "Moonshine-Streaming-Tiny-GGUF", "Moonshine-Streaming-Small-GGUF", "Moonshine-Streaming-Medium-GGUF":
		family = "moonshine_asr"
	case "Niagara-19M-Batch-English-GGUF", "Niagara-38M-Batch-English-GGUF":
		family = "niagara_asr"
	case "MOSS-Transcribe-Diarize-GGUF":
		family = "moss_transcribe_diarize"
	case "VibeVoice-ASR-Streaming-7B-GGUF":
		family = "vibevoice_asr_streaming"
	default:
		return audioCppSTTSchema{}, false
	}
	schema := audioCppSTTSchema{family: family, title: model}
	if family == "canary_asr" || family == "cohere_asr" {
		schema.groups = []audioCppSTTGroup{{title: lang.L("Settings"), columns: 1, fields: []audioCppSTTField{audioCppSTTCheckbox("pnc", lang.L("Punctuation and capitalization"), true)}}}
	}
	return schema, true
}
