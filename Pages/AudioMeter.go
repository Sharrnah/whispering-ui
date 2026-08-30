package Pages

import (
	"encoding/binary"
	"math"
)

const (
	audioMeterFloorDB = -60.0
	audioMeterMax     = 100.0
)

// audioMeterLevel converts a normalized RMS amplitude into a conventional
// decibel meter. The fixed -60 dBFS to 0 dBFS range keeps the bar comparable
// over time instead of rescaling it to whichever sample happened to be loudest.
func audioMeterLevel(sumSquares float64, sampleCount int) float64 {
	if sampleCount == 0 || sumSquares <= 0 {
		return 0
	}

	return normalizedAudioMeterLevel(math.Sqrt(sumSquares / float64(sampleCount)))
}

func normalizedAudioMeterLevel(amplitude float64) float64 {
	if amplitude <= 0 {
		return 0
	}
	if amplitude >= 1 {
		return audioMeterMax
	}

	decibels := 20 * math.Log10(amplitude)
	if decibels <= audioMeterFloorDB {
		return 0
	}

	return (decibels - audioMeterFloorDB) / -audioMeterFloorDB * audioMeterMax
}

func s32AudioMeterLevel(samples []byte) float64 {
	sampleCount := len(samples) / 4
	if sampleCount == 0 {
		return 0
	}

	var sumSquares float64
	for offset := 0; offset+4 <= len(samples); offset += 4 {
		sample := int32(binary.LittleEndian.Uint32(samples[offset : offset+4]))
		normalized := float64(sample) / float64(math.MaxInt32)
		if normalized < -1 {
			normalized = -1
		}
		sumSquares += normalized * normalized
	}

	return audioMeterLevel(sumSquares, sampleCount)
}

// s32AudioPeakMeterLevel matches the backend's speech-volume trigger, which
// compares the largest absolute int16 sample in each chunk instead of RMS.
func s32AudioPeakMeterLevel(samples []byte) float64 {
	var peak int64
	for offset := 0; offset+4 <= len(samples); offset += 4 {
		sample := int64(int32(binary.LittleEndian.Uint32(samples[offset : offset+4])))
		if sample < 0 {
			sample = -sample
		}
		if sample > peak {
			peak = sample
		}
	}
	if peak == 0 {
		return 0
	}
	return normalizedAudioMeterLevel(float64(peak) / float64(int64(math.MaxInt32)+1))
}

// s32SpeechTriggerMeterLevel follows the recorder selected in the profile:
// the VAD recorder uses peak amplitude, while speech_recognition's legacy
// non-VAD recorder applies the same energy value to an RMS measurement.
func s32SpeechTriggerMeterLevel(samples []byte, vadEnabled bool) float64 {
	if vadEnabled {
		return s32AudioPeakMeterLevel(samples)
	}
	return s32AudioMeterLevel(samples)
}

func f32AudioMeterLevel(samples []byte) float64 {
	var sumSquares float64
	validSampleCount := 0
	for offset := 0; offset+4 <= len(samples); offset += 4 {
		sample := float64(math.Float32frombits(binary.LittleEndian.Uint32(samples[offset : offset+4])))
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			continue
		}
		if sample > 1 {
			sample = 1
		} else if sample < -1 {
			sample = -1
		}
		sumSquares += sample * sample
		validSampleCount++
	}

	return audioMeterLevel(sumSquares, validSampleCount)
}
