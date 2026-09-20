package Settings

import "testing"

func TestLegacyAudioCppTTSPrecisionDefaultsToOriginalModelDtype(t *testing.T) {
	if precision := legacyTTSPrecision("audio_cpp", nil); precision != "orig" {
		t.Fatalf("audio.cpp precision = %q, want orig", precision)
	}
}
