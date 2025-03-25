package Hardwareinfo

import (
	"fyne.io/fyne/v2/widget"
	"strings"
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
	// WhisperCT2 models
	{"Whisper", "faster_whisper", "tiny", 1054.0},
	{"Whisper", "faster_whisper", "base", 1185.0},
	{"Whisper", "faster_whisper", "small", 1873.0},
	{"Whisper", "faster_whisper", "medium", 3905.0},
	{"Whisper", "faster_whisper", "large", 6985.0},
	{"Whisper", "faster_whisper", "medium-distilled", 1898.0},
	{"Whisper", "faster_whisper", "large-distilled", 3339.0},
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
	{"Whisper", "nemo_canary", "tiny", 4509.0},
	{"Whisper", "nemo_canary", "base", 4509.0},
	{"Whisper", "nemo_canary", "small", 4509.0},
	{"Whisper", "nemo_canary", "medium", 4509.0},
	{"Whisper", "nemo_canary", "large", 4509.0},
	// MMS models
	{"Whisper", "mms", "1b-all", 4646.0},
	{"Whisper", "mms", "mms-1b-fl102", 4544.0},
	{"Whisper", "mms", "mms-1b-l1107", 4623.0},
	// Phi-4 model
	{"Whisper", "phi4", "", 22531.0},
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
	// TTS types
	{"ttsType", "silero", "", 1533.0},
	{"ttsType", "f5_e2", "", 1200.0},
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

func (p ProfileAIModelOption) CalculateMemoryConsumption(CPUBar *widget.ProgressBar, GPUBar *widget.ProgressBar, totalGPUMemory int64) {
	println("Calculating memory usage for target type:", p.AIModel+"#"+p.AIModelType+"#"+p.AIModelSize)
	addToList := true
	lastIndex := -1
	for index, profileAIModelOption := range AllProfileAIModelOptions {
		if profileAIModelOption.AIModel == p.AIModel {
			println("Device updated...")
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
		for _, model := range Models {
			if model.BaseName == AllProfileAIModelOptions[lastIndex].AIModel &&
				model.ModelType == AllProfileAIModelOptions[lastIndex].AIModelType &&
				model.ModelSize == AllProfileAIModelOptions[lastIndex].AIModelSize {
				finalMemoryUsage := EstimateMemoryUsage(model.Float32PrecisionMemoryUsage, AllProfileAIModelOptions[lastIndex].Precision)
				println("Full model match found:")
				println("BaseName: " + model.BaseName + ", ModelType: " + model.ModelType + ", ModelSize: " + model.ModelSize)
				println("finalMemoryUsage:")
				println(int(finalMemoryUsage))
				AllProfileAIModelOptions[lastIndex].MemoryConsumption = finalMemoryUsage
			}
		}
	}

	if addToList {
		println("Device added...")
		AllProfileAIModelOptions = append(AllProfileAIModelOptions, p)
	}

	// Deduplicate the list based on AIModel, AIModelType, and AIModelSize
	uniqueOptions := make(map[string]ProfileAIModelOption)
	for _, option := range AllProfileAIModelOptions {
		//key := option.AIModel + "|" + option.AIModelType + "|" + option.AIModelSize
		key := option.AIModelType + "|" + option.AIModelSize
		// Only add if key does not exist
		if _, exists := uniqueOptions[key]; !exists {
			uniqueOptions[key] = option
		}
	}

	// Reset progress bars
	GPUBar.Value = 0.0
	CPUBar.Value = 0.0

	// Sum up the memory consumption for each unique option
	//for _, profileAIModelOption := range AllProfileAIModelOptions {
	for _, profileAIModelOption := range uniqueOptions {
		println(profileAIModelOption.AIModel, profileAIModelOption.MemoryConsumption)
		deviceLower := strings.ToLower(profileAIModelOption.Device)
		println(profileAIModelOption.AIModel, profileAIModelOption.MemoryConsumption)
		if strings.HasPrefix(deviceLower, "cuda") ||
			strings.HasPrefix(deviceLower, "direct-ml") {
			println("CUDA MEMORY:")
			println(int(profileAIModelOption.MemoryConsumption))
			if totalGPUMemory == 0 {
				GPUBar.Max = GPUBar.Value + profileAIModelOption.MemoryConsumption
			}
			GPUBar.Value += profileAIModelOption.MemoryConsumption
		} else if strings.HasPrefix(deviceLower, "cpu") {
			println("CPU MEMORY:")
			println(int(profileAIModelOption.MemoryConsumption))
			CPUBar.Value += profileAIModelOption.MemoryConsumption
		}
	}
	CPUBar.Refresh()
	GPUBar.Refresh()
}
