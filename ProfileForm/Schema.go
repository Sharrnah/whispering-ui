package ProfileForm

import (
	"whispering-tiger-ui/CustomWidget"

	"fyne.io/fyne/v2/lang"
)

type TVO = CustomWidget.TextValueOption

func MultiModalModels() map[string]bool {
	return map[string]bool{
		"seamless_m4t": true,
		"phi4":         true,
		"voxtral":      true,
	}
}

func DefaultDeviceOptions() []TVO {
	return []TVO{{Text: "CPU", Value: "cpu"}, {Text: "CUDA", Value: "cuda"}}
}

// Generic precision lists used for initial population; Coordinator will narrow them by type later
func GenericWhisperPrecisionOptions() []TVO {
	return []TVO{
		{Text: "float32 " + lang.L("Precision"), Value: "float32"},
		{Text: "float16 " + lang.L("Precision"), Value: "float16"},
		{Text: "int16 " + lang.L("Precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("Precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
		{Text: "8bit " + lang.L("Precision"), Value: "8bit"},
		{Text: "4bit " + lang.L("Precision"), Value: "4bit"},
	}
}

func GenericTextPrecisionOptions() []TVO {
	return []TVO{
		{Text: "float32 " + lang.L("Precision"), Value: "float32"},
		{Text: "float16 " + lang.L("Precision"), Value: "float16"},
		{Text: "int16 " + lang.L("Precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("Precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
	}
}

func GenericOcrPrecisionOptions() []TVO {
	return []TVO{
		{Text: "float32 " + lang.L("Precision"), Value: "float32"},
		{Text: "float16 " + lang.L("Precision"), Value: "float16"},
		{Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"},
	}
}

// Base type-option providers for initial selects
func STTTypeOptions() []TVO {
	return []TVO{{Text: "Faster Whisper", Value: "faster_whisper"}, {Text: "Original Whisper", Value: "original_whisper"}, {Text: "Transformer Whisper", Value: "transformer_whisper"}, {Text: "Seamless M4T", Value: "seamless_m4t"}, {Text: "MMS", Value: "mms"}, {Text: "Speech T5 (English only)", Value: "speech_t5"}, {Text: "Wav2Vec Bert 2.0", Value: "wav2vec_bert"}, {Text: "NeMo Canary", Value: "nemo_canary"}, {Text: "Phi-4", Value: "phi4"}, {Text: "Voxtral", Value: "voxtral"}, {Text: lang.L("Disabled"), Value: ""}}
}

func TXTTypeOptions() []TVO {
	return []TVO{{Text: "Faster NLLB200 (200 languages)", Value: "NLLB200_CT2"}, {Text: "Original NLLB200 (200 languages)", Value: "NLLB200"}, {Text: "M2M100 (100 languages)", Value: "M2M100"}, {Text: "Seamless M4T (101 languages)", Value: "seamless_m4t"}, {Text: "Phi-4 (23 languages)", Value: "phi4"}, {Text: "Voxtral (13 languages)", Value: "voxtral"}, {Text: lang.L("Disabled"), Value: ""}}
}

func TTSTypeOptions() []TVO {
	//return []TVO{{Text: "Silero", Value: "silero"}, {Text: "F5/E2", Value: "f5_e2"}, {Text: "Zonos", Value: "zonos"}, {Text: "Kokoro", Value: "kokoro"}, {Text: "Orpheus", Value: "orpheus"}, {Text: "Parler", Value: "parler"}, {Text: lang.L("Disabled"), Value: ""}}
	return []TVO{{Text: "Silero", Value: "silero"}, {Text: "F5/E2", Value: "f5_e2"}, {Text: "Zonos", Value: "zonos"}, {Text: "Kokoro", Value: "kokoro"}, {Text: "Orpheus", Value: "orpheus"}, {Text: "Chatterbox", Value: "chatterbox"}, {Text: lang.L("Disabled"), Value: ""}}
}

func OcrTypeOptions() []TVO {
	return []TVO{{Text: "Easy OCR", Value: "easyocr"}, {Text: "GOT OCR 2.0", Value: "got_ocr_20"}, {Text: "Phi-4", Value: "phi4"}, {Text: lang.L("Disabled"), Value: ""}}
}

func STTModelOptions(modelType string) (options []TVO, defaultIndex int, enableSize bool) {
	enableSize = true
	switch modelType {
	case "faster_whisper":
		return []TVO{{Text: "Tiny", Value: "tiny"}, {Text: "Tiny (English only)", Value: "tiny.en"}, {Text: "Base", Value: "base"}, {Text: "Base (English only)", Value: "base.en"}, {Text: "Small", Value: "small"}, {Text: "Small (English only)", Value: "small.en"}, {Text: "Medium", Value: "medium"}, {Text: "Medium (English only)", Value: "medium.en"}, {Text: "Large V1", Value: "large-v1"}, {Text: "Large V2", Value: "large-v2"}, {Text: "Large V3", Value: "large-v3"}, {Text: "Large V3 Turbo", Value: "large-v3-turbo"}, {Text: "Medium Distilled (English)", Value: "medium-distilled.en"}, {Text: "Large V2 Distilled (English)", Value: "large-distilled-v2.en"}, {Text: "Large V3 Distilled (English)", Value: "large-distilled-v3.en"}, {Text: "Large V3.5 Distilled (English)", Value: "large-distilled-v3.5.en"}, {Text: "Crisper", Value: "crisper"}, {Text: "Small (European finetune)", Value: "small.eu"}, {Text: "Medium (European finetune)", Value: "medium.eu"}, {Text: "Small (German finetune)", Value: "small.de"}, {Text: "Medium (German finetune)", Value: "medium.de"}, {Text: "Large V2 (German finetune)", Value: "large-v2.de2"}, {Text: "Large V3 Distilled (German finetune)", Value: "large-distilled-v3.de"}, {Text: "Small (German-Swiss finetune)", Value: "small.de-swiss"}, {Text: "Medium (Mix-Japanese-v2 finetune)", Value: "medium.mix-jpv2"}, {Text: "Large V2 (Mix-Japanese finetune)", Value: "large-v2.mix-jp"}, {Text: "Small (Japanese finetune)", Value: "small.jp"}, {Text: "Medium (Japanese finetune)", Value: "medium.jp"}, {Text: "Large V2 (Japanese finetune)", Value: "large-v2.jp"}, {Text: "Medium (Korean finetune)", Value: "medium.ko"}, {Text: "Large V2 (Korean finetune)", Value: "large-v2.ko"}, {Text: "Small (Chinese finetune)", Value: "small.zh"}, {Text: "Medium (Chinese finetune)", Value: "medium.zh"}, {Text: "Large V2 (Chinese finetune)", Value: "large-v2.zh"}, {Text: "Custom (Place in '.cache/whisper/custom-ct2' directory)", Value: "custom"}}, 0, true
	case "original_whisper":
		return []TVO{{Text: "Tiny", Value: "tiny"}, {Text: "Tiny (English only)", Value: "tiny.en"}, {Text: "Base", Value: "base"}, {Text: "Base (English only)", Value: "base.en"}, {Text: "Small", Value: "small"}, {Text: "Small (English only)", Value: "small.en"}, {Text: "Medium", Value: "medium"}, {Text: "Medium (English only)", Value: "medium.en"}, {Text: "Large V1", Value: "large-v1"}, {Text: "Large V2", Value: "large-v2"}, {Text: "Large V3", Value: "large-v3"}, {Text: "Large V3 Turbo", Value: "large-v3-turbo"}, {Text: "Custom (Place in '.cache/whisper/custom' directory)", Value: "custom"}}, 0, true
	case "transformer_whisper":
		return []TVO{{Text: "Tiny", Value: "tiny"}, {Text: "Tiny (English only)", Value: "tiny.en"}, {Text: "Base", Value: "base"}, {Text: "Base (English only)", Value: "base.en"}, {Text: "Small", Value: "small"}, {Text: "Small (English only)", Value: "small.en"}, {Text: "Medium", Value: "medium"}, {Text: "Medium (English only)", Value: "medium.en"}, {Text: "Large V1", Value: "large-v1"}, {Text: "Large V2", Value: "large-v2"}, {Text: "Large V3", Value: "large-v3"}, {Text: "Large V3 Turbo", Value: "large-v3-turbo"}, {Text: "Custom (Place in '.cache/whisper-transformer/custom' directory)", Value: "custom"}}, 0, true
	case "medusa_whisper":
		return []TVO{{Text: "V1", Value: "v1"}}, 0, true
	case "seamless_m4t":
		return []TVO{{Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}, {Text: "Large V2", Value: "large-v2"}}, 1, true
	case "mms":
		return []TVO{{Text: "1b-fl102 (102 languages)", Value: "mms-1b-fl102"}, {Text: "1b-l1107 (1107 languages)", Value: "mms-1b-l1107"}, {Text: "1b-all (1162 languages)", Value: "1b-all"}}, 1, true
	case "nemo_canary":
		return []TVO{{Text: "Nemo Canary 1b", Value: "canary-1b"}, {Text: "Nemo Canary 180m flash", Value: "canary-180m-flash"}, {Text: "Nemo Canary 1b flash", Value: "canary-1b-flash"}, {Text: "Parakeet TDT 0.6B V2 (English)", Value: "parakeet-tdt-0_6b-v2"}, {Text: "Parakeet TDT 0.6B V3 (Multilingual)", Value: "parakeet-tdt-0_6b-v3"}}, 0, true
	case "phi4":
		// Phi-4 has a fixed model size; provide option for display but disable the selector in Coordinator
		return []TVO{{Text: "Large", Value: "large"}}, 0, false
	case "voxtral":
		return []TVO{{Text: "Voxtral-Mini-3B-2507", Value: "Voxtral-Mini-3B-2507"}}, 0, true
	case "speech_t5":
		return nil, 0, false
	default:
		return nil, 0, false
	}
}

func STTPrecisionOptions(modelType string) (options []TVO, enablePrecision bool) {
	switch modelType {
	case "faster_whisper":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "int16 " + lang.L("Precision"), Value: "int16"}, {Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"}, {Text: "int8 " + lang.L("Precision"), Value: "int8"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "original_whisper", "medusa_whisper":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}}, true
	case "transformer_whisper", "wav2vec_bert", "mms", "voxtral":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "8bit " + lang.L("Precision"), Value: "8bit"}, {Text: "4bit " + lang.L("Precision"), Value: "4bit"}}, true
	case "seamless_m4t":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "nemo_canary":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}}, false
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"}}, true
	case "speech_t5":
		return nil, false
	default:
		return nil, false
	}
}

func TXTSizeOptions(modelType string) (options []TVO, defaultIndex int, enableSize bool) {
	enableSize = true
	switch modelType {
	case "M2M100":
		return []TVO{{Text: "Small", Value: "small"}, {Text: "Large", Value: "large"}}, 0, true
	case "NLLB200_CT2", "NLLB200":
		return []TVO{{Text: "Small", Value: "small"}, {Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}}, 0, true
	case "seamless_m4t":
		return []TVO{{Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}, {Text: "Large V2", Value: "large-v2"}}, 0, true
	case "phi4":
		return []TVO{{Text: "Large", Value: "large"}}, 0, true
	case "voxtral":
		return []TVO{{Text: "Voxtral-Mini-3B-2507", Value: "Voxtral-Mini-3B-2507"}}, 0, true
	default:
		return nil, 0, false
	}
}

func TXTPrecisionOptions(modelType string) (options []TVO, enablePrecision bool) {
	switch modelType {
	case "NLLB200":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}}, true
	case "NLLB200_CT2", "M2M100":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "int16 " + lang.L("Precision"), Value: "int16"}, {Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"}, {Text: "int8 " + lang.L("Precision"), Value: "int8"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "seamless_m4t":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}}, true
	case "voxtral":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "8bit " + lang.L("Precision"), Value: "8bit"}, {Text: "4bit " + lang.L("Precision"), Value: "4bit"}}, true
	default:
		return nil, false
	}
}

// OCRPrecisionOptions returns precision lists for OCR types and whether the precision selector should be enabled.
// Mirrors the pattern of STTPrecisionOptions/TXTPrecisionOptions for consistency.
func OCRPrecisionOptions(modelType string) (options []TVO, enablePrecision bool) {
	switch modelType {
	case "easyocr":
		// EasyOCR is CPU-only in our app and does not expose precision tuning
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}}, false
	case "got_ocr_20":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"}}, true
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"}}, true
	default:
		return nil, false
	}
}

// OCRDeviceOptions returns allowed AI device options by OCR model type.
// For easyocr we only allow CPU; others default to CPU/CUDA like the rest of the app.
func OCRDeviceOptions(modelType string) []TVO {
	switch modelType {
	case "easyocr":
		return []TVO{{Text: "CPU", Value: "cpu"}}
	case "got_ocr_20", "phi4":
		return DefaultDeviceOptions()
	default:
		return DefaultDeviceOptions()
	}
}
