package Pages

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestNormalizedAudioMeterLevel(t *testing.T) {
	tests := []struct {
		name      string
		amplitude float64
		want      float64
	}{
		{name: "silence", amplitude: 0, want: 0},
		{name: "below floor", amplitude: 0.0001, want: 0},
		{name: "minus six decibels", amplitude: 0.5, want: 90},
		{name: "full scale", amplitude: 1, want: 100},
		{name: "clamped above full scale", amplitude: 2, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizedAudioMeterLevel(tt.amplitude)
			if math.Abs(got-tt.want) > 0.05 {
				t.Fatalf("normalizedAudioMeterLevel() = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}

func TestS32AudioMeterLevel(t *testing.T) {
	tests := []struct {
		name    string
		samples []int32
		want    float64
	}{
		{name: "silence", samples: []int32{0, 0}, want: 0},
		{name: "full scale", samples: []int32{math.MaxInt32, math.MinInt32}, want: 100},
		{name: "minus six decibels", samples: []int32{math.MaxInt32 / 2, math.MinInt32 / 2}, want: 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.samples)*4)
			for index, sample := range tt.samples {
				binary.LittleEndian.PutUint32(data[index*4:], uint32(sample))
			}

			got := s32AudioMeterLevel(data)
			if math.Abs(got-tt.want) > 0.05 {
				t.Fatalf("s32AudioMeterLevel() = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}

func TestS32AudioPeakMeterLevel(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, uint32(math.MaxInt32))
	binary.LittleEndian.PutUint32(data[4:], 0)
	got := s32AudioPeakMeterLevel(data)
	if math.Abs(got-100) > 0.05 {
		t.Fatalf("s32AudioPeakMeterLevel() = %.3f, want 100", got)
	}
}

func TestSpeechTriggerMeterMatchesSelectedRecorder(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, uint32(math.MaxInt32))
	binary.LittleEndian.PutUint32(data[4:], 0)

	peakLevel := s32SpeechTriggerMeterLevel(data, true)
	rmsLevel := s32SpeechTriggerMeterLevel(data, false)
	if peakLevel <= rmsLevel {
		t.Fatalf("VAD peak level %.3f should exceed legacy RMS level %.3f", peakLevel, rmsLevel)
	}
	if math.Abs(peakLevel-100) > 0.05 {
		t.Fatalf("VAD peak level = %.3f, want 100", peakLevel)
	}
}

func TestEnergyMeterPositionUsesBackendPeakScale(t *testing.T) {
	if got := energyMeterPosition(0); got != 0 {
		t.Fatalf("disabled energy marker position = %.3f, want 0", got)
	}
	if got := energyMeterPosition(32767); math.Abs(float64(got)-1) > 0.001 {
		t.Fatalf("full-scale energy marker position = %.3f, want 1", got)
	}
	low := energyMeterPosition(300)
	high := energyMeterPosition(1000)
	if low <= 0 || low >= high || high >= 1 {
		t.Fatalf("energy marker positions are not ordered: 300=%0.3f, 1000=%0.3f", low, high)
	}
}

func TestEnergyMarkerMatchesBackendPCM16TriggerBoundary(t *testing.T) {
	const energy = 300
	data := make([]byte, 4)
	// Miniaudio supplies S32 samples. Shifting a PCM16 threshold by 16 bits
	// produces the same normalized amplitude seen by the Python PCM16 backend.
	binary.LittleEndian.PutUint32(data, uint32(int32(energy<<16)))

	visibleLevel := s32AudioPeakMeterLevel(data) / audioMeterMax
	markerPosition := float64(energyMeterPosition(energy))
	if math.Abs(visibleLevel-markerPosition) > 0.000001 {
		t.Fatalf("meter/marker boundary = %.9f/%.9f", visibleLevel, markerPosition)
	}
}

func TestF32AudioMeterLevel(t *testing.T) {
	tests := []struct {
		name    string
		samples []float32
		want    float64
	}{
		{name: "silence", samples: []float32{0, 0}, want: 0},
		{name: "full scale", samples: []float32{1, -1}, want: 100},
		{name: "minus six decibels", samples: []float32{0.5, -0.5}, want: 90},
		{name: "invalid samples ignored", samples: []float32{float32(math.NaN()), 1}, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.samples)*4)
			for index, sample := range tt.samples {
				binary.LittleEndian.PutUint32(data[index*4:], math.Float32bits(sample))
			}

			got := f32AudioMeterLevel(data)
			if math.Abs(got-tt.want) > 0.05 {
				t.Fatalf("f32AudioMeterLevel() = %.3f, want %.3f", got, tt.want)
			}
		})
	}
}
