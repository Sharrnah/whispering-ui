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
	// NLLB200CT2 models
	{"NLLB200CT2_small", 3087.0},
	{"NLLB200CT2_medium", 6069.0},
	{"NLLB200CT2_large", 13803.0},
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
