package ProfileForm

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/Hardwareinfo"
)

func TestBuildProfileMemoryOptionCapturesCompleteSelection(t *testing.T) {
	testApp := fyneTest.NewApp()
	t.Cleanup(testApp.Quit)

	size := CustomWidget.NewTextValueSelect("model", []TVO{
		{Text: "0.6B", Value: "Qwen3-ASR-0.6B-hf"},
		{Text: "1.7B", Value: "Qwen3-ASR-1.7B-hf"},
	}, nil, 0)
	precision := CustomWidget.NewTextValueSelect("precision", []TVO{
		{Text: "float32", Value: "float32"},
		{Text: "bfloat16", Value: "bfloat16"},
	}, nil, 0)
	device := CustomWidget.NewTextValueSelect("device", []TVO{
		{Text: "CPU", Value: "cpu"},
		{Text: "CUDA", Value: "cuda"},
	}, nil, 0)
	size.SetSelected("Qwen3-ASR-1.7B-hf")
	precision.SetSelected("bfloat16")
	device.SetSelected("cuda")

	got := BuildProfileMemoryOption("Whisper", "qwen3_asr", size, precision, device)
	if got.AIModel != "Whisper" || got.AIModelType != "qwen3_asr" {
		t.Fatalf("unexpected model identity: %#v", got)
	}
	if got.AIModelSize != "Qwen3-ASR-1.7B-hf" {
		t.Fatalf("unexpected model size: %q", got.AIModelSize)
	}
	if got.Precision != Hardwareinfo.Float16 || got.Device != "cuda" {
		t.Fatalf("unexpected precision/device snapshot: %#v", got)
	}
}
