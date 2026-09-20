package Utilities

import (
	"math"
	"testing"
)

func TestEnergyThresholdDBFSUsesPCM16Scale(t *testing.T) {
	tests := []struct {
		energy float64
		want   float64
	}{
		{energy: 32768, want: 0},
		{energy: 16384, want: -6.020599913},
		{energy: 300, want: -40.7665736},
	}

	for _, test := range tests {
		got := EnergyThresholdDBFS(test.energy)
		if math.Abs(got-test.want) > 0.0001 {
			t.Fatalf("EnergyThresholdDBFS(%v) = %.6f, want %.6f", test.energy, got, test.want)
		}
	}
}

func TestFormatEnergyThresholdDBFSKeepsZeroAsDisabled(t *testing.T) {
	if got := FormatEnergyThresholdDBFS(0, "Disabled"); got != "Disabled" {
		t.Fatalf("zero threshold label = %q, want Disabled", got)
	}
	if got := FormatEnergyThresholdDBFS(300, "Disabled"); got != "-40.8 dBFS" {
		t.Fatalf("300 threshold label = %q, want -40.8 dBFS", got)
	}
}
