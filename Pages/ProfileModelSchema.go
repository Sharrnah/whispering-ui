package Pages

import (
	"whispering-tiger-ui/CustomWidget"

	"fyne.io/fyne/v2/lang"
)

// Schema for available options per model type and task, to keep logic centralized
// and reusable across UI and programmatic updates.

// Convenience alias
type TVO = CustomWidget.TextValueOption

// Multi-modal models used across tasks
func MultiModalModels() map[string]bool {
	return map[string]bool{
		"seamless_m4t": true,
		"phi4":         true,
		"voxtral":      true,
	}
}

// Common device options
func DefaultDeviceOptions() []TVO {
	return []TVO{{Text: "CUDA", Value: "cuda"}, {Text: "CPU", Value: "cpu"}}
}

// STT options
func STTModelOptions(modelType string) (options []TVO, defaultIndex int, enableSize bool) {
	enableSize = true
	switch modelType {
	case "faster_whisper":
		return []TVO{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large V1", Value: "large-v1"},
			{Text: "Large V2", Value: "large-v2"},
			{Text: "Large V3", Value: "large-v3"},
			{Text: "Large V3 Turbo", Value: "large-v3-turbo"},
			{Text: "Medium Distilled (English)", Value: "medium-distilled.en"},
			{Text: "Large V2 Distilled (English)", Value: "large-distilled-v2.en"},
			{Text: "Large V3 Distilled (English)", Value: "large-distilled-v3.en"},
			{Text: "Large V3.5 Distilled (English)", Value: "large-distilled-v3.5.en"},
			{Text: "Crisper", Value: "crisper"},
			{Text: "Small (European finetune)", Value: "small.eu"},
			{Text: "Medium (European finetune)", Value: "medium.eu"},
			{Text: "Small (German finetune)", Value: "small.de"},
			{Text: "Medium (German finetune)", Value: "medium.de"},
			{Text: "Large V2 (German finetune)", Value: "large-v2.de2"},
			{Text: "Large V3 Distilled (German finetune)", Value: "large-distilled-v3.de"},
			{Text: "Small (German-Swiss finetune)", Value: "small.de-swiss"},
			{Text: "Medium (Mix-Japanese-v2 finetune)", Value: "medium.mix-jpv2"},
			{Text: "Large V2 (Mix-Japanese finetune)", Value: "large-v2.mix-jp"},
			{Text: "Small (Japanese finetune)", Value: "small.jp"},
			{Text: "Medium (Japanese finetune)", Value: "medium.jp"},
			{Text: "Large V2 (Japanese finetune)", Value: "large-v2.jp"},
			{Text: "Medium (Korean finetune)", Value: "medium.ko"},
			{Text: "Large V2 (Korean finetune)", Value: "large-v2.ko"},
			{Text: "Small (Chinese finetune)", Value: "small.zh"},
			{Text: "Medium (Chinese finetune)", Value: "medium.zh"},
			{Text: "Large V2 (Chinese finetune)", Value: "large-v2.zh"},
			{Text: "Custom (Place in '.cache/whisper/custom-ct2' directory)", Value: "custom"},
		}, 0, true
	case "original_whisper":
		return []TVO{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large V1", Value: "large-v1"},
			{Text: "Large V2", Value: "large-v2"},
			{Text: "Large V3", Value: "large-v3"},
			{Text: "Large V3 Turbo", Value: "large-v3-turbo"},
			{Text: "Custom (Place in '.cache/whisper/custom' directory)", Value: "custom"},
		}, 0, true
	case "transformer_whisper":
		return []TVO{
			{Text: "Tiny", Value: "tiny"},
			{Text: "Tiny (English only)", Value: "tiny.en"},
			{Text: "Base", Value: "base"},
			{Text: "Base (English only)", Value: "base.en"},
			{Text: "Small", Value: "small"},
			{Text: "Small (English only)", Value: "small.en"},
			{Text: "Medium", Value: "medium"},
			{Text: "Medium (English only)", Value: "medium.en"},
			{Text: "Large V1", Value: "large-v1"},
			{Text: "Large V2", Value: "large-v2"},
			{Text: "Large V3", Value: "large-v3"},
			{Text: "Large V3 Turbo", Value: "large-v3-turbo"},
			{Text: "Custom (Place in '.cache/whisper-transformer/custom' directory)", Value: "custom"},
		}, 0, true
	case "medusa_whisper":
		return []TVO{{Text: "V1", Value: "v1"}}, 0, true
	case "seamless_m4t":
		return []TVO{{Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}, {Text: "Large V2", Value: "large-v2"}}, 1, true
	case "mms":
		return []TVO{{Text: "1b-fl102 (102 languages)", Value: "mms-1b-fl102"}, {Text: "1b-l1107 (1107 languages)", Value: "mms-1b-l1107"}, {Text: "1b-all (1162 languages)", Value: "1b-all"}}, 1, true
	case "nemo_canary":
		return []TVO{{Text: "Nemo Canary 1b", Value: "canary-1b"}, {Text: "Nemo Canary 180m flash", Value: "canary-180m-flash"}, {Text: "Nemo Canary 1b flash", Value: "canary-1b-flash"}, {Text: "Parakeet TDT 0.6B V2 (English)", Value: "parakeet-tdt-0_6b-v2"}}, 0, true
	case "phi4":
		return []TVO{{Text: "Large", Value: "large"}}, 0, true
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
		return []TVO{
			{Text: "float32 " + lang.L("precision"), Value: "float32"},
			{Text: "float16 " + lang.L("precision"), Value: "float16"},
			{Text: "int16 " + lang.L("precision"), Value: "int16"},
			{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
			{Text: "int8 " + lang.L("precision"), Value: "int8"},
			{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
			{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
		}, true
	case "original_whisper", "medusa_whisper":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}}, true
	case "transformer_whisper", "wav2vec_bert", "mms", "voxtral":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "8bit " + lang.L("precision"), Value: "8bit"}, {Text: "4bit " + lang.L("precision"), Value: "4bit"}}, true
	case "seamless_m4t":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"}, {Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "nemo_canary":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}}, false
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("precision"), Value: "bfloat16"}}, true
	case "speech_t5":
		return nil, false
	default:
		return nil, false
	}
}

// TXT options
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
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}}, true
	case "NLLB200_CT2", "M2M100":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "int16 " + lang.L("precision"), Value: "int16"}, {Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"}, {Text: "int8 " + lang.L("precision"), Value: "int8"}, {Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "seamless_m4t":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"}, {Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"}}, true
	case "voxtral":
		return []TVO{{Text: "float32 " + lang.L("precision"), Value: "float32"}, {Text: "float16 " + lang.L("precision"), Value: "float16"}, {Text: "8bit " + lang.L("precision"), Value: "8bit"}, {Text: "4bit " + lang.L("precision"), Value: "4bit"}}, true
	default:
		return nil, false
	}
}

func TXTDeviceOptions(modelType string) []TVO {
	// CTranslate2 models do not support DirectML here (UI only has CUDA/CPU)
	return DefaultDeviceOptions()
}

// TTS options
func TTSDeviceOptions(modelType string) []TVO {
	// Most models share CUDA/CPU; specific restrictions can be added here later
	return DefaultDeviceOptions()
}
