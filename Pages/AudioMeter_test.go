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
