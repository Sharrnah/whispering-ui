package ProfileForm

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/Hardwareinfo"

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

// AudioCppDeviceOptions exposes only managed audio.cpp backends. Vulkan is the
// portable GPU path for AMD and Intel (and also supports NVIDIA). HIP/ROCm is
// intentionally hidden because upstream does not publish a managed binary;
// the backend still accepts it for legacy/manual custom-runtime profiles.
func AudioCppDeviceOptions(_ string) []TVO {
	if runtime.GOOS == "linux" {
		// The managed Linux runtime archives currently provide CPU and Vulkan.
		return []TVO{
			{Text: lang.L("CPU (most compatible, slower)"), Value: "cpu"},
			{Text: lang.L("Vulkan (AMD / Intel recommended; NVIDIA supported)"), Value: "vulkan"},
		}
	}
	if runtime.GOOS == "darwin" {
		return []TVO{
			{Text: lang.L("CPU (most compatible, slower)"), Value: "cpu"},
			{Text: lang.L("Metal (Apple GPU)"), Value: "metal"},
		}
	}
	return []TVO{
		{Text: lang.L("CPU (most compatible, slower)"), Value: "cpu"},
		{Text: lang.L("CUDA (NVIDIA, recommended)"), Value: "cuda"},
		{Text: lang.L("Vulkan (AMD / Intel recommended; NVIDIA supported)"), Value: "vulkan"},
	}
}

func STTDeviceOptions(modelType string) []TVO {
	if modelType == "audio_cpp" {
		return AudioCppDeviceOptions(modelType)
	}
	return DefaultDeviceOptions()
}

func TTSDeviceOptions(modelType string) []TVO {
	if modelType == "audio_cpp" {
		return AudioCppDeviceOptions(modelType)
	}
	return DefaultDeviceOptions()
}

// DefaultGPUOptions keeps profile loading deterministic before asynchronous
// hardware discovery finishes. Detected adapter names replace these entries.
func DefaultGPUOptions() []TVO {
	options := make([]TVO, 16)
	for index := range options {
		options[index] = TVO{Text: fmt.Sprintf("%s %d", lang.L("GPU"), index), Value: fmt.Sprint(index)}
	}
	return options
}

func CUDADeviceOptions(devices []Hardwareinfo.CUDADeviceInfo) []TVO {
	options := make([]TVO, 0, len(devices))
	for _, device := range devices {
		label := fmt.Sprintf("%s %d", lang.L("GPU"), device.Index)
		if device.Name != "" {
			label += " - " + device.Name
		}
		options = append(options, TVO{Text: label, Value: fmt.Sprint(device.Index)})
	}
	return options
}

// AudioCppGPUOptions converts the runtime's backend-specific --list-devices
// output into selector entries. Named OS adapters are used only as an
// explicitly marked estimate before a runtime is available locally.
func AudioCppGPUOptions(devices []Hardwareinfo.AudioCppDeviceInfo, graphics []Hardwareinfo.GPUInfo) map[string][]TVO {
	options := make(map[string][]TVO)
	for _, device := range devices {
		backend := strings.ToLower(strings.TrimSpace(device.Backend))
		if backend == "rocm" {
			backend = "hip"
		}
		if backend != "vulkan" && backend != "hip" && backend != "metal" {
			continue
		}
		label := fmt.Sprintf("%s %d", lang.L("GPU"), device.Index)
		if device.Name != "" {
			label += " - " + device.Name
		}
		if device.Kind != "" {
			label += " [" + device.Kind + "]"
		}
		options[backend] = append(options[backend], TVO{Text: label, Value: fmt.Sprint(device.Index)})
	}
	for backend := range options {
		sort.Slice(options[backend], func(i, j int) bool {
			left, _ := strconv.Atoi(options[backend][i].Value)
			right, _ := strconv.Atoi(options[backend][j].Value)
			return left < right
		})
	}

	if len(options["vulkan"]) == 0 {
		for index, device := range graphics {
			label := fmt.Sprintf("%s %d - %s (%s)", lang.L("GPU"), index, device.AdapterName, lang.L("estimated OS order"))
			options["vulkan"] = append(options["vulkan"], TVO{Text: label, Value: fmt.Sprint(index)})
		}
	}
	if len(options["hip"]) == 0 {
		index := 0
		for _, device := range graphics {
			vendorAndName := strings.ToLower(device.VendorName + " " + device.AdapterName)
			if !strings.Contains(vendorAndName, "amd") && !strings.Contains(vendorAndName, "advanced micro devices") && !strings.Contains(vendorAndName, "radeon") {
				continue
			}
			label := fmt.Sprintf("%s %d - %s (%s)", lang.L("GPU"), index, device.AdapterName, lang.L("estimated OS order"))
			options["hip"] = append(options["hip"], TVO{Text: label, Value: fmt.Sprint(index)})
			index++
		}
	}
	if runtime.GOOS == "darwin" && len(options["metal"]) == 0 {
		for index, device := range graphics {
			label := fmt.Sprintf("%s %d - %s (%s)", lang.L("GPU"), index, device.AdapterName, lang.L("estimated OS order"))
			options["metal"] = append(options["metal"], TVO{Text: label, Value: fmt.Sprint(index)})
		}
	}
	return options
}

func GenericTTSPrecisionOptions() []TVO {
	return []TVO{
		{Text: lang.L("Automatic (BF16 on supported CUDA)"), Value: "auto"},
		{Text: "float32 " + lang.L("Precision"), Value: "float32"},
		{Text: "float16 " + lang.L("Precision"), Value: "float16"},
		{Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "8bit " + lang.L("Precision"), Value: "8bit"},
	}
}

func TTSPrecisionOptions(modelType string) (options []TVO, enablePrecision bool) {
	switch modelType {
	case "silero", "f5_e2", "kokoro":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}}, false
	case "zonos", "zonos2", "maya1":
		return []TVO{{Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"}}, false
	case "orpheus":
		return []TVO{{Text: "8bit " + lang.L("Precision"), Value: "8bit"}}, false
	case "chatterbox":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}}, true
	case "index_tts":
		return []TVO{{Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "float32 " + lang.L("Precision"), Value: "float32"}}, true
	case "qwen3_tts":
		return []TVO{{Text: lang.L("Automatic (BF16 on supported CUDA)"), Value: "auto"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: lang.L("Float16 request (safe BF16/FP32 fallback)"), Value: "float16"}, {Text: "float32 " + lang.L("Precision"), Value: "float32"}}, true
	case "audio8_tts":
		return []TVO{{Text: lang.L("Automatic (BF16 on supported CUDA)"), Value: "auto"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "float32 " + lang.L("Precision"), Value: "float32"}}, true
	case "audio_cpp":
		return TTSPrecisionOptionsForModel(modelType, "Supertonic-3-GGUF")
	default:
		return nil, false
	}
}

// TTSPrecisionOptionsForModel narrows audio.cpp to the GGUF variants that
// actually exist for the selected package. Other TTS engines keep their
// type-level precision contract.
func TTSPrecisionOptionsForModel(modelType, modelName string) (options []TVO, enablePrecision bool) {
	if modelType != "audio_cpp" {
		return TTSPrecisionOptions(modelType)
	}
	q8 := TVO{Text: "Q8_0 GGUF — recommended balance", Value: "q8_0"}
	f16 := TVO{Text: "F16 GGUF — higher precision, more memory", Value: "f16"}
	bf16 := TVO{Text: "BF16 GGUF — higher precision, more memory", Value: "bf16"}
	orig := TVO{Text: "Original model dtype — highest fidelity", Value: "orig"}
	switch modelName {
	case "Confucius4-TTS-GGUF":
		return []TVO{orig}, false
	case "DotTTS-SOAR-GGUF", "DotTTS-MeanFlow-GGUF", "DotTTS-Edit-GGUF":
		return []TVO{q8, bf16}, true
	case "IndexTTS2-GGUF", "IndexTTS2.5-GGUF":
		return []TVO{q8, f16, orig}, true
	case "MagpieTTS-Multilingual-357M-GGUF":
		return []TVO{q8, orig}, true
	case "OmniVoice-GGUF":
		return []TVO{q8, f16, bf16}, true
	case "VoxCPM1-0.5B-GGUF":
		return []TVO{q8}, false
	case "VoxCPM2-GGUF":
		return []TVO{q8, bf16, orig}, true
	case "Supertonic-3-GGUF", "":
		return []TVO{orig, f16}, true
	default:
		return []TVO{q8}, false
	}
}

// TTSModelOptions contains only model lists that are part of the profile
// schema itself. Most Python TTS engines report their models dynamically after
// startup. audio.cpp publishes a fixed set of GGUF packages, so those can be
// selected before the Python backend starts.
func TTSModelOptions(modelType string) (options []TVO, defaultIndex int, enableModel bool) {
	switch modelType {
	case "audio_cpp":
		modelNames := []string{
			"Supertonic-3-GGUF",
			"Confucius4-TTS-GGUF",
			"DotTTS-SOAR-GGUF",
			"DotTTS-MeanFlow-GGUF",
			"DotTTS-Edit-GGUF",
			"IndexTTS2-GGUF",
			"IndexTTS2.5-GGUF",
			"MagpieTTS-Multilingual-357M-GGUF",
			"OmniVoice-GGUF",
			"VoxCPM1-0.5B-GGUF",
			"VoxCPM2-GGUF",
		}
		options = make([]TVO, 0, len(modelNames))
		for _, modelName := range modelNames {
			group, _ := TTSModelProfileGroup(modelType, modelName)
			options = append(options, TVO{
				Text:  TTSModelDisplayText(modelType, group, modelName),
				Value: modelName,
			})
		}
		return options, 0, true
	default:
		return nil, 0, false
	}
}

// TTSModelProfileGroup restores the two-part backend profile representation
// for models that can be selected directly in this form.
func TTSModelProfileGroup(modelType, modelName string) (string, bool) {
	if modelType != "audio_cpp" {
		return "", false
	}
	switch modelName {
	case "Supertonic-3-GGUF", "MagpieTTS-Multilingual-357M-GGUF":
		return "Preset voices", true
	case "Confucius4-TTS-GGUF", "DotTTS-SOAR-GGUF", "DotTTS-MeanFlow-GGUF":
		return "Voice cloning", true
	case "DotTTS-Edit-GGUF":
		return "Speech editing", true
	case "IndexTTS2-GGUF", "IndexTTS2.5-GGUF":
		return "Voice cloning and emotion", true
	case "OmniVoice-GGUF", "VoxCPM2-GGUF":
		return "Cloning and voice design", true
	case "VoxCPM1-0.5B-GGUF":
		return "Cloning and text to speech", true
	}
	return "", false
}

func TTSModelDisplayText(modelType, modelGroup, modelName string) string {
	if modelType == "audio_cpp" {
		descriptions := map[string]string{
			"Supertonic-3-GGUF":                "fastest; 31 languages; 10 preset voices",
			"Confucius4-TTS-GGUF":              "experimental multilingual voice cloning; very large",
			"DotTTS-SOAR-GGUF":                 "multilingual cloning and style control; SOAR baseline",
			"DotTTS-MeanFlow-GGUF":             "multilingual cloning and style control; MeanFlow",
			"DotTTS-Edit-GGUF":                 "edits an existing recording; not ordinary TTS",
			"IndexTTS2-GGUF":                   "Chinese/English cloning and emotion control",
			"IndexTTS2.5-GGUF":                 "multilingual cloning and emotion control",
			"MagpieTTS-Multilingual-357M-GGUF": "small; 13 locales; 5 preset voices",
			"OmniVoice-GGUF":                   "600+ languages; cloning and voice design",
			"VoxCPM1-0.5B-GGUF":                "compact cloning model; 16 kHz output",
			"VoxCPM2-GGUF":                     "2B cloning and voice design; 48 kHz output",
		}
		if description := descriptions[modelName]; description != "" {
			return modelName + " (" + description + ")"
		}
	}
	if modelType == "silero" || strings.TrimSpace(modelGroup) == "" {
		return modelName
	}
	return modelName + " (" + lang.L(modelGroup) + ")"
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
		{Text: "8bit " + lang.L("Precision"), Value: "8bit"},
		{Text: "4bit " + lang.L("Precision"), Value: "4bit"},
	}
}

func GenericOcrPrecisionOptions() []TVO {
	return []TVO{
		{Text: "float32 " + lang.L("Precision"), Value: "float32"},
		{Text: "float16 " + lang.L("Precision"), Value: "float16"},
		{Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"},
		{Text: "8bit " + lang.L("Precision"), Value: "8bit"},
		{Text: "4bit " + lang.L("Precision"), Value: "4bit"},
	}
}

// Base type-option providers for initial selects
func STTTypeOptions() []TVO {
	return []TVO{{Text: "Faster Whisper", Value: "faster_whisper"}, {Text: "Original Whisper", Value: "original_whisper"}, {Text: "Transformer Whisper", Value: "transformer_whisper"}, {Text: "Qwen3-ASR", Value: "qwen3_asr"}, {Text: lang.L("audio.cpp (native GGUF runtime)"), Value: "audio_cpp"}, {Text: "Seamless M4T", Value: "seamless_m4t"}, {Text: "MMS", Value: "mms"}, {Text: "Speech T5 (English only)", Value: "speech_t5"}, {Text: "Wav2Vec Bert 2.0", Value: "wav2vec_bert"}, {Text: "NeMo Canary", Value: "nemo_canary"},
		//{Text: "Phi-4", Value: "phi4"},
		{Text: "VibeVoice-ASR", Value: "vibevoice_asr"},
		{Text: "Voxtral", Value: "voxtral"}, {Text: "Higgs Audio", Value: "higgs_audio"}, {Text: lang.L("Disabled"), Value: ""}}
}

func TXTTypeOptions() []TVO {
	return []TVO{{Text: "Faster NLLB200 (200 languages)", Value: "NLLB200_CT2"}, {Text: "Original NLLB200 (200 languages)", Value: "NLLB200"}, {Text: "M2M100 (100 languages)", Value: "M2M100"}, {Text: "Seamless M4T (101 languages)", Value: "seamless_m4t"}, {Text: "Hunyuan MT (33 languages)", Value: "hunyuan_mt"}, {Text: "MiLMMT-46 (46 languages)", Value: "milmmt"},
		//{Text: "Phi-4 (23 languages)", Value: "phi4"},
		{Text: "Voxtral (13 languages)", Value: "voxtral"}, {Text: lang.L("Disabled"), Value: ""}}
}

func TTSTypeOptions() []TVO {
	//return []TVO{{Text: "Silero", Value: "silero"}, {Text: "F5/E2", Value: "f5_e2"}, {Text: "Zonos", Value: "zonos"}, {Text: "Kokoro", Value: "kokoro"}, {Text: "Orpheus", Value: "orpheus"}, {Text: "Parler", Value: "parler"}, {Text: lang.L("Disabled"), Value: ""}}
	return []TVO{{Text: "Silero", Value: "silero"}, {Text: "F5/E2 (Voice Cloning)", Value: "f5_e2"}, {Text: "Zonos", Value: "zonos"}, {Text: "ZONOS2 (BF16 / FP8) (Voice Cloning)", Value: "zonos2"}, {Text: "Kokoro", Value: "kokoro"}, {Text: "Orpheus", Value: "orpheus"}, {Text: "Chatterbox (Voice Cloning)", Value: "chatterbox"}, {Text: "IndexTTS 2.5 (Voice Cloning / Emotion)", Value: "index_tts"}, {Text: "Qwen3-TTS (Cloning / Voice Design)", Value: "qwen3_tts"}, {Text: lang.L("Audio8 TTS Preview (Voice Cloning)"), Value: "audio8_tts"}, {Text: lang.L("audio.cpp (native GGUF runtime)"), Value: "audio_cpp"}, {Text: "Maya1 (Emotion)", Value: "maya1"}, {Text: lang.L("Disabled"), Value: ""}}
}

func OcrTypeOptions() []TVO {
	return []TVO{{Text: "Easy OCR", Value: "easyocr"}, {Text: "GOT OCR 2.0", Value: "got_ocr_20"},
		//{Text: "Phi-4", Value: "phi4"},
		{Text: lang.L("Disabled"), Value: ""}}
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
	case "qwen3_asr":
		return []TVO{{Text: "Qwen3-ASR 0.6B (faster / lower memory)", Value: "Qwen3-ASR-0.6B-hf"}, {Text: "Qwen3-ASR 1.7B (best quality)", Value: "Qwen3-ASR-1.7B-hf"}, {Text: "Custom (Place in '.cache/qwen3-asr/custom' directory)", Value: "custom"}}, 0, true
	case "audio_cpp":
		return []TVO{
			{Text: "Qwen3-ASR 0.6B — recommended general multilingual model", Value: "Qwen3-ASR-0.6B-GGUF"},
			{Text: "Qwen3-ASR 1.7B — higher accuracy, more memory", Value: "Qwen3-ASR-1.7B-GGUF"},
			{Text: "Nemotron 3.5 ASR 0.6B — fastest streaming, 40 locales, timestamps", Value: "Nemotron-3.5-ASR-Streaming-0.6B-GGUF"},
			{Text: "VibeVoice-ASR — long meetings, speaker turns, very large", Value: "VibeVoice-ASR-GGUF"},
			{Text: "Voxtral Realtime 4B — low-latency multilingual; no timestamps", Value: "Voxtral-Mini-4B-Realtime-2602-GGUF"},
			{Text: "Audio8-ASR 0.1B — tiny local-only model; non-commercial license", Value: "Audio8-ASR-0.1B-GGUF"},
			{Text: "Kroko ASR 64L — tiny English streaming model with timestamps", Value: "Kroko-ASR-English-64L-GGUF"},
		}, 0, true
	case "medusa_whisper":
		return []TVO{{Text: "V1", Value: "v1"}}, 0, true
	case "seamless_m4t":
		return []TVO{{Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}, {Text: "Large V2", Value: "large-v2"}}, 1, true
	case "mms":
		return []TVO{{Text: "1b-fl102 (102 languages)", Value: "mms-1b-fl102"}, {Text: "1b-l1107 (1107 languages)", Value: "mms-1b-l1107"}, {Text: "1b-all (1162 languages)", Value: "1b-all"}}, 1, true
	case "nemo_canary":
		return []TVO{{Text: "Nemo Canary 1b v2", Value: "canary-1b-v2"}, {Text: "Nemo Canary 1b", Value: "canary-1b"}, {Text: "Nemo Canary 180m flash", Value: "canary-180m-flash"}, {Text: "Nemo Canary 1b flash", Value: "canary-1b-flash"}, {Text: "Parakeet TDT 0.6B V2 (English)", Value: "parakeet-tdt-0_6b-v2"}, {Text: "Parakeet TDT 0.6B V3 (Multilingual)", Value: "parakeet-tdt-0_6b-v3"}}, 0, true
	case "phi4":
		// Phi-4 has a fixed model size; provide option for display but disable the selector in Coordinator
		return []TVO{{Text: "Large", Value: "large"}}, 0, false
	case "voxtral":
		return []TVO{{Text: "Voxtral-Mini-3B-2507", Value: "Voxtral-Mini-3B-2507"}, {Text: "Voxtral-Mini-4B-Realtime-2602", Value: "Voxtral-Mini-4B-Realtime-2602"}}, 0, true
	case "speech_t5":
		return nil, 0, false
	case "vibevoice_asr":
		return nil, 0, false
	case "higgs_audio":
		return []TVO{{Text: "Higgs Audio v3", Value: "higgs-audio-v3-stt"}}, 0, true
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
	case "qwen3_asr":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "8bit " + lang.L("Precision"), Value: "8bit"}, {Text: "4bit " + lang.L("Precision"), Value: "4bit"}}, true
	case "audio_cpp":
		return STTPrecisionOptionsForModel(modelType, "Qwen3-ASR-0.6B-GGUF")
	case "seamless_m4t":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "int8_float16 " + lang.L("Precision"), Value: "int8_float16"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "int8_bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "int8_bfloat16"}}, true
	case "nemo_canary":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}}, false
	case "phi4":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "bfloat16 " + lang.L("Precision"), Value: "bfloat16"}}, true
	case "speech_t5":
		return nil, false
	case "vibevoice_asr":
		// 8bit only resulted in (unintelligible speech) output
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}, {Text: "4bit " + lang.L("Precision"), Value: "4bit"}}, true
	default:
		return nil, false
	}
}

// STTPrecisionOptionsForModel narrows audio.cpp to the downloadable (or, for
// Audio8, locally supplied) package variants for the selected ASR model.
func STTPrecisionOptionsForModel(modelType, modelName string) (options []TVO, enablePrecision bool) {
	if modelType != "audio_cpp" {
		return STTPrecisionOptions(modelType)
	}
	q4 := TVO{Text: "Q4_K GGUF — smallest and fastest", Value: "q4_k"}
	q8 := TVO{Text: "Q8_0 GGUF — recommended balance", Value: "q8_0"}
	f16 := TVO{Text: "F16 GGUF — higher precision, more memory", Value: "f16"}
	bf16 := TVO{Text: "BF16 GGUF — highest published precision", Value: "bf16"}
	switch modelName {
	case "Voxtral-Mini-4B-Realtime-2602-GGUF":
		return []TVO{q4, q8, bf16}, true
	case "Kroko-ASR-English-64L-GGUF":
		return []TVO{q8}, false
	case "Audio8-ASR-0.1B-GGUF":
		return []TVO{q8, f16}, true
	case "Qwen3-ASR-0.6B-GGUF", "Qwen3-ASR-1.7B-GGUF",
		"Nemotron-3.5-ASR-Streaming-0.6B-GGUF", "VibeVoice-ASR-GGUF", "":
		return []TVO{q8, f16}, true
	default:
		return []TVO{q8}, false
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
	case "hunyuan_mt":
		return []TVO{{Text: "Small", Value: "small"}, {Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}}, 0, true
	case "milmmt":
		return []TVO{{Text: "MiLMMT-46 1B v1.0 (faster / lower memory)", Value: "MiLMMT-46-1B-v1.0"}, {Text: "MiLMMT-46 4B v1.0 (balanced)", Value: "MiLMMT-46-4B-v1.0"}, {Text: "MiLMMT-46 12B v1.0 (best quality)", Value: "MiLMMT-46-12B-v1.0"}, {Text: "Custom (Place in '.cache/milmmt/custom' directory)", Value: "custom"}}, 0, true
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
	case "hunyuan_mt":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "float16 " + lang.L("Precision"), Value: "float16"}}, true
	case "milmmt":
		return []TVO{{Text: "float32 " + lang.L("Precision"), Value: "float32"}, {Text: "bfloat16 " + lang.L("Precision") + " (Compute >=8.0)", Value: "bfloat16"}, {Text: "8bit " + lang.L("Precision"), Value: "8bit"}}, true
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
		// EasyOCR chooses CPU/CUDA independently but does not expose precision tuning.
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
// All integrated OCR engines accept the shared CPU/CUDA selection.
func OCRDeviceOptions(modelType string) []TVO {
	switch modelType {
	case "easyocr":
		return DefaultDeviceOptions()
	case "got_ocr_20", "phi4":
		return DefaultDeviceOptions()
	default:
		return DefaultDeviceOptions()
	}
}
