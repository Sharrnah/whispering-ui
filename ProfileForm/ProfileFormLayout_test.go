package ProfileForm

import "testing"

func TestDenoisedTriggerSharesNoiseFilterRow(t *testing.T) {
	denoiseRows := 0
	for _, row := range BuildFullProfileLayout() {
		if row.CustomKey == "DenoiseRow" {
			denoiseRows++
		}
		for _, controlName := range row.ControlNames {
			if controlName == "DenoiseBeforeTrigger" {
				t.Fatal("denoised trigger is still rendered as a separate profile row")
			}
		}
	}
	if denoiseRows != 1 {
		t.Fatalf("DenoiseRow count = %d, want 1", denoiseRows)
	}
}
