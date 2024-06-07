package Hardwareinfo

type AIModel struct {
	Name                        string
	Float32PrecisionMemoryUsage float64
}

// Models is a list of AIModel which is a struct that contains the name of the model and the memory usage in MB for Float32
var Models = []AIModel{
	// Official Whisper models
	{"WhisperO_tiny", 1676.0},
	{"WhisperO_base", 1932.0},
	{"WhisperO_small", 3432.0},
	{"WhisperO_medium", 7634.0},
	{"WhisperO_large", 13702.0},
	// WhisperCT2 models
	{"WhisperCT2_tiny", 1054.0},
	{"WhisperCT2_base", 1185.0},
	{"WhisperCT2_small", 1873.0},
	{"WhisperCT2_medium", 3905.0},
	{"WhisperCT2_large", 6985.0},
	{"WhisperCT2_medium-distilled", 1898.0},
	{"WhisperCT2_large-distilled", 3339.0},
	// Speech T5
	{"Whispert5_tiny", 927.0},
	{"Whispert5_base", 927.0},
	{"Whispert5_small", 927.0},
	{"Whispert5_medium", 927.0},
	{"Whispert5_large", 927.0},
	// Seamless M4T
	{"Whisperm4t_medium", 6250.0},
	{"Whisperm4t_large", 10518.0},
	// wav2vec bert2.0 models
	{"Whisperwav2vec-bert_tiny", 2989.0},
	{"Whisperwav2vec-bert_base", 2989.0},
	{"Whisperwav2vec-bert_small", 2989.0},
	{"Whisperwav2vec-bert_medium", 2989.0},
	{"Whisperwav2vec-bert_large", 2989.0},
	// NeMo Canary models
	{"Whispernemo-canary_tiny", 4509.0},
	{"Whispernemo-canary_base", 4509.0},
	{"Whispernemo-canary_small", 4509.0},
	{"Whispernemo-canary_medium", 4509.0},
	{"Whispernemo-canary_large", 4509.0},
	// MMS models
	{"Whispermms_1b-all", 4646.0},
	// NLLB200CT2 models
	{"TxtTranslatorNLLB200_CT2_small", 3087.0},
	{"TxtTranslatorNLLB200_CT2_medium", 6069.0},
	{"TxtTranslatorNLLB200_CT2_large", 13803.0},
	// NLLB200 models
	{"TxtTranslatorNLLB200_small", 3657.0},
	{"TxtTranslatorNLLB200_medium", 6620.0},
	{"TxtTranslatorNLLB200_large", 14837.0},
	// M2M100 models
	{"TxtTranslatorM2M100_small", 2197.0},
	{"TxtTranslatorM2M100_large", 5211.0},
	// Seamless M4T models
	{"TxtTranslatorSeamless_M4T_medium", 6250.0},
	{"TxtTranslatorSeamless_M4T_large", 10518.0},
	// Silero TTS
	{"SileroO_", 1533.0},
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
