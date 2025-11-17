package Hardwareinfo

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type AIModel struct {
	BaseName                    string
	ModelType                   string
	ModelSize                   string
	Float32PrecisionMemoryUsage float64
}

// Models is a list of AIModel which is a struct that contains the name of the model and the memory usage in MB for Float32
var Models = []AIModel{
	// Official Whisper models
	{"Whisper", "O", "tiny", 1676.0},
	{"Whisper", "O", "base", 1932.0},
	{"Whisper", "O", "small", 3432.0},
	{"Whisper", "O", "medium", 7634.0},
	{"Whisper", "O", "large", 13702.0},
	// Duplicate for original_whisper Type mapping
	{"Whisper", "original_whisper", "tiny", 1676.0},
	{"Whisper", "original_whisper", "base", 1932.0},
	{"Whisper", "original_whisper", "small", 3432.0},
	{"Whisper", "original_whisper", "medium", 7634.0},
	{"Whisper", "original_whisper", "large", 13702.0},
	// WhisperCT2 models
	{"Whisper", "faster_whisper", "tiny", 1054.0},
	{"Whisper", "faster_whisper", "base", 1185.0},
	{"Whisper", "faster_whisper", "small", 1873.0},
	{"Whisper", "faster_whisper", "medium", 3905.0},
	{"Whisper", "faster_whisper", "large", 6985.0},
	{"Whisper", "faster_whisper", "medium-distilled", 1898.0},
	{"Whisper", "faster_whisper", "large-distilled", 3339.0},
	// Transformer Whisper models (angenommen ähnlich Original)
	{"Whisper", "transformer_whisper", "tiny", 1676.0},
	{"Whisper", "transformer_whisper", "base", 1932.0},
	{"Whisper", "transformer_whisper", "small", 3432.0},
	{"Whisper", "transformer_whisper", "medium", 7634.0},
	{"Whisper", "transformer_whisper", "large", 13702.0},
	// Speech T5
	{"Whisper", "speech_t5", "tiny", 927.0},
	{"Whisper", "speech_t5", "base", 927.0},
	{"Whisper", "speech_t5", "small", 927.0},
	{"Whisper", "speech_t5", "medium", 927.0},
	{"Whisper", "speech_t5", "large", 927.0},
	// Seamless M4T
	{"Whisper", "seamless_m4t", "medium", 6250.0},
	{"Whisper", "seamless_m4t", "large", 10518.0},
	// wav2vec bert2.0 models
	{"Whisper", "wav2vec_bert", "tiny", 2989.0},
	{"Whisper", "wav2vec_bert", "base", 2989.0},
	{"Whisper", "wav2vec_bert", "small", 2989.0},
	{"Whisper", "wav2vec_bert", "medium", 2989.0},
	{"Whisper", "wav2vec_bert", "large", 2989.0},
	// NeMo Canary models
	{"Whisper", "nemo_canary", "canary-1b", 4509.0},
	{"Whisper", "nemo_canary", "canary-180m-flash", 702.0},
	{"Whisper", "nemo_canary", "canary-1b-flash", 4000.0},
	{"Whisper", "nemo_canary", "parakeet-tdt-0_6b-v2", 2828.0},
	{"Whisper", "nemo_canary", "parakeet-tdt-0_6b-v3", 2928.0},
	// MMS models
	{"Whisper", "mms", "1b-all", 4646.0},
	{"Whisper", "mms", "mms-1b-fl102", 4544.0},
	{"Whisper", "mms", "mms-1b-l1107", 4623.0},
	// Phi-4 model
	{"Whisper", "phi4", "", 22531.0},
	// Voxtral model
	{"Whisper", "voxtral", "", 18852.0},
	// NLLB200CT2 models
	{"TxtTranslator", "NLLB200_CT2", "small", 3087.0},
	{"TxtTranslator", "NLLB200_CT2", "medium", 6069.0},
	{"TxtTranslator", "NLLB200_CT2", "large", 13803.0},
	// NLLB200 models
	{"TxtTranslator", "NLLB200", "small", 3657.0},
	{"TxtTranslator", "NLLB200", "medium", 6620.0},
	{"TxtTranslator", "NLLB200", "large", 14837.0},
	// M2M100 models
	{"TxtTranslator", "M2M100", "small", 2197.0},
	{"TxtTranslator", "M2M100", "large", 5211.0},
	// Seamless M4T models
	{"TxtTranslator", "seamless_m4t", "medium", 6250.0},
	{"TxtTranslator", "seamless_m4t", "large", 10518.0},
	{"TxtTranslator", "seamless_m4t", "large-v2", 10518.0},
	// Phi-4 model
	{"TxtTranslator", "phi4", "", 22531.0},
	// Voxtral model
	{"TxtTranslator", "voxtral", "", 18852.0},
	// TTS types
	{"ttsType", "silero", "", 1533.0},
	{"ttsType", "f5_e2", "", 1200.0},
	{"ttsType", "zonos", "", 3030.0},
	{"ttsType", "kokoro", "", 312.0},
	{"ttsType", "chatterbox", "", 3470.0},
	// OCR types
	{"ocrType", "easyocr", "", 520.0},
	{"ocrType", "got_ocr_20", "", 1559.0},
	{"ocrType", "phi4", "", 22531.0},
}

const (
	Float32 = 4.0
	Float16 = 2.0
	Int32   = 4.0
	Int16   = 2.0
	Int8    = 1.0

	Bit8 = 1.0
	Bit4 = 0.5
)

func EstimateMemoryUsage(float32MemoryUsage float64, targetType float64) float64 {
	return (float32MemoryUsage / float64(Float32)) * float64(targetType)
}

type ProfileAIModelOption struct {
	AIModel           string
	AIModelType       string
	AIModelSize       string
	Precision         float64
	Device            string
	MemoryConsumption float64
}

var AllProfileAIModelOptions = make([]ProfileAIModelOption, 0)
var calculating bool

func (p ProfileAIModelOption) CalculateMemoryConsumption(CPUBar *widget.ProgressBar, GPUBar *widget.ProgressBar, totalGPUMemory int64) {
	if CPUBar == nil || GPUBar == nil {
		return
	}
	if calculating {
		return
	}
	calculating = true
	defer func() { calculating = false }()
	// Normalize incoming size to canonical form before using it
	if p.AIModelSize != "" {
		p.AIModelSize = normalizeModelSize(p.AIModelType, p.AIModelSize)
	}
	// Debug reduziert, um Log-Spam zu vermeiden
	addToList := true
	lastIndex := -1
	for index, profileAIModelOption := range AllProfileAIModelOptions {
		if profileAIModelOption.AIModel == p.AIModel {
			// update existing entry
			if p.Device != "" {
				AllProfileAIModelOptions[index].Device = p.Device
			}
			if p.AIModelType != "" {
				AllProfileAIModelOptions[index].AIModelType = p.AIModelType
			}
			if p.AIModelSize != "" {
				AllProfileAIModelOptions[index].AIModelSize = p.AIModelSize
			}
			if p.Precision != 0 {
				AllProfileAIModelOptions[index].Precision = p.Precision
			}
			AllProfileAIModelOptions[index].MemoryConsumption = p.MemoryConsumption
			addToList = false
			lastIndex = index
			break
		}
	}
	if lastIndex > -1 && len(AllProfileAIModelOptions) >= lastIndex+1 {
		// normalize size for matching known models
		currentType := AllProfileAIModelOptions[lastIndex].AIModelType
		currentSize := AllProfileAIModelOptions[lastIndex].AIModelSize
		normSize := normalizeModelSize(currentType, currentSize)
		for _, model := range Models {
			if model.BaseName == AllProfileAIModelOptions[lastIndex].AIModel &&
				model.ModelType == AllProfileAIModelOptions[lastIndex].AIModelType &&
				(model.ModelSize == "" || model.ModelSize == normSize) {
				AllProfileAIModelOptions[lastIndex].AIModelSize = model.ModelSize

				finalMemoryUsage := EstimateMemoryUsage(model.Float32PrecisionMemoryUsage, AllProfileAIModelOptions[lastIndex].Precision)
				// full model match -> aktualisiere Memory
				AllProfileAIModelOptions[lastIndex].MemoryConsumption = finalMemoryUsage
			}
		}
	}

	if addToList {
		// add new entry (normalized)
		AllProfileAIModelOptions = append(AllProfileAIModelOptions, p)
	}

	// Deduplicate the list.
	// Für multimodale Modelle (seamless_m4t, phi4, voxtral) zählen wir nur einmal pro ModelType
	// über alle Aufgaben hinweg, damit die gemeinsam genutzten Gewichte nicht mehrfach addiert werden.
	uniqueOptions := make(map[string]ProfileAIModelOption)
	//isMultiModal := func(t string) bool {
	//	switch t {
	//	case "seamless_m4t", "phi4", "voxtral":
	//		return true
	//	default:
	//		return false
	//	}
	//}
	for _, option := range AllProfileAIModelOptions {
		key := option.AIModelType + "|" + option.AIModelSize
		//key := ""
		//if isMultiModal(option.AIModelType) {
		//	// nur pro Typ, größe beachten falls vorhanden
		//	key = option.AIModelType + "|" + option.AIModelSize
		//} else {
		//	// generell nur ModelType|ModelSize (ohne AIModel) wie gewünscht
		//	key = option.AIModelType + "|" + option.AIModelSize
		//}
		if _, exists := uniqueOptions[key]; !exists {
			uniqueOptions[key] = option
		}
	}

	// Reset progress bars
	fyne.Do(func() {
		GPUBar.Value = 0.0
		CPUBar.Value = 0.0
		// Setze Max auf bekannte Gesamt-GPU-Menge, wenn verfügbar.
		// Falls totalGPUMemory==0, belasse bestehenden Max-Wert (z.B. bereits beim Start gesetzt).
		if totalGPUMemory > 0 {
			GPUBar.Max = float64(totalGPUMemory)
		}

		// Sum up the memory consumption for each unique option
		for _, profileAIModelOption := range uniqueOptions {
			deviceLower := strings.ToLower(profileAIModelOption.Device)
			if strings.HasPrefix(deviceLower, "cuda") ||
				strings.HasPrefix(deviceLower, "direct-ml") {
				// Wenn Gesamtwert unbekannt UND Max aktuell 0 ist, skaliere Max dynamisch
				if totalGPUMemory == 0 && GPUBar.Max == 0 {
					GPUBar.Max = GPUBar.Value + profileAIModelOption.MemoryConsumption
				}
				GPUBar.Value += profileAIModelOption.MemoryConsumption
			} else if strings.HasPrefix(deviceLower, "cpu") {
				CPUBar.Value += profileAIModelOption.MemoryConsumption
			}
		}
		CPUBar.Refresh()
		GPUBar.Refresh()
	})
}

// normalizeModelSize reduziert Variantenbezeichner auf kanonische Größen
// z.B. tiny.en -> tiny, large-v2 -> large, medium-distilled -> medium
func normalizeModelSize(modelType, size string) string {
	if size == "" {
		return ""
	}
	s := strings.ToLower(size)
	// entferne Sprach-/Region-Suffixe wie .en, .de, .jp, .eu etc.
	if idx := strings.Index(s, "."); idx > 0 {
		s = s[:idx]
	}
	// distilled-Varianten erkennen
	if strings.Contains(size, "distilled") {
		if strings.Contains(s, "large") {
			return "large-distilled"
		}
		if strings.Contains(s, "medium") {
			return "medium-distilled"
		}
	}
	// Whisper-Familie und Abwandlungen auf Grundgrößen mappen
	if strings.HasPrefix(s, "tiny") {
		return "tiny"
	}
	if strings.HasPrefix(s, "base") {
		return "base"
	}
	if strings.HasPrefix(s, "small") {
		return "small"
	}
	if strings.HasPrefix(s, "medium") {
		return "medium"
	}
	if strings.HasPrefix(s, "large") {
		return "large"
	}
	return s
}
