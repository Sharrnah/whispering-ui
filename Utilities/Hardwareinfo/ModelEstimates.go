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
	// Speech T5
	{"Whispert5_tiny", 927.0},
	{"Whispert5_base", 927.0},
	{"Whispert5_small", 927.0},
	{"Whispert5_medium", 927.0},
	{"Whispert5_large", 927.0},
	// Seamless M4T
	{"Whisperm4t_medium", 6250.0},
	{"Whisperm4t_large", 10518.0},
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
	Float32 = 4
	Float16 = 2
	Int32   = 4
	Int16   = 2
	Int8    = 1
)

func EstimateMemoryUsage(float32MemoryUsage float64, targetType int) float64 {
	return (float32MemoryUsage / float64(Float32)) * float64(targetType)
}
