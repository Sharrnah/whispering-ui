package Utilities

import (
	"fmt"
	"math"
)

const pcm16FullScale = 32768.0

// EnergyThresholdAmplitude converts the backend's signed-PCM16 energy value
// into a full-scale amplitude. The backend keeps using the original integer;
// this conversion is only for UI meters and labels.
func EnergyThresholdAmplitude(energy float64) float64 {
	if math.IsNaN(energy) || energy <= 0 {
		return 0
	}
	return energy / pcm16FullScale
}

// EnergyThresholdDBFS converts the backend's signed-PCM16 energy value to
// decibels relative to full scale. Zero is returned as negative infinity
// because the application uses zero to disable automatic volume triggering.
func EnergyThresholdDBFS(energy float64) float64 {
	amplitude := EnergyThresholdAmplitude(energy)
	if amplitude <= 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(amplitude)
}

// FormatEnergyThresholdDBFS formats an energy value for display while letting
// the caller supply a localized label for the disabled (zero) state.
func FormatEnergyThresholdDBFS(energy float64, disabledLabel string) string {
	decibels := EnergyThresholdDBFS(energy)
	if math.IsInf(decibels, -1) {
		return disabledLabel
	}
	return fmt.Sprintf("%.1f dBFS", decibels)
}
