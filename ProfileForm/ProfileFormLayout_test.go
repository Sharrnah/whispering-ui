package ProfileForm

import (
	"reflect"
	"testing"

	"fyne.io/fyne/v2/lang"
)

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

func TestAIProfileRowsIncludeGPUSelectorsAndTTSPrecision(t *testing.T) {
	want := map[string][]string{
		lang.L("A.I. Device for Speech-to-Text"):   {"STTDevice", "STTGPU"},
		lang.L("A.I. Device for Text-Translation"): {"TxtDevice", "TxtGPU"},
		lang.L("A.I. Device for Text-to-Speech"):   {"TTSDevice", "TTSGPU", "TTSPrecision"},
		lang.L("A.I. Device for Image-to-Text"):    {"OCRDevice", "OCRGPU", "OCRPrecision"},
	}
	for _, row := range BuildFullProfileLayout() {
		controls, exists := want[row.Label]
		if !exists {
			continue
		}
		if !reflect.DeepEqual(row.ControlNames, controls) {
			t.Fatalf("row %q controls = %#v, want %#v", row.Label, row.ControlNames, controls)
		}
		delete(want, row.Label)
	}
	if len(want) != 0 {
		t.Fatalf("missing AI device rows: %#v", want)
	}
}
