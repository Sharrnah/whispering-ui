package Pages

import (
	"testing"

	"github.com/gen2brain/malgo"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities"
)

func TestApplicationAudioOptionIsInsertedAfterDefault(t *testing.T) {
	options := []CustomWidget.TextValueOption{
		{Text: "Default", Value: "-1"},
		{Text: "Microphone A", Value: "1"},
		{Text: "Microphone B", Value: "2"},
	}

	result := appendApplicationAudioInputOption(options, malgo.BackendWasapi)
	if len(result) != 4 {
		t.Fatalf("option count = %d, want 4", len(result))
	}
	if result[0].Value != "-1" || result[1].Value != Utilities.AudioApplicationOptionValue {
		t.Fatalf("first options = %#v, want Default followed by Application Audio", result[:2])
	}
	if result[2].Value != "1" || result[3].Value != "2" {
		t.Fatalf("device ordering changed: %#v", result)
	}
}

func TestApplicationAudioOptionHandlesCachedDefaultSuffix(t *testing.T) {
	options := []CustomWidget.TextValueOption{
		{Text: "Default#|WASAPI", Value: "-1#|WASAPI"},
		{Text: "Microphone", Value: "1#|WASAPI"},
	}

	result := appendApplicationAudioInputOption(options, malgo.BackendWasapi)
	if result[1].Value != Utilities.AudioApplicationOptionValue {
		t.Fatalf("Application Audio index = %#v, want index 1", result)
	}
}

func TestBuildLiveAudioInputSwitchRequestForDevice(t *testing.T) {
	input := &CustomWidget.TextValueOption{Text: "USB microphone", Value: "4"}
	request, ok := buildLiveAudioInputSwitchRequest("request-1", "WASAPI", input, nil)
	if !ok {
		t.Fatal("device request was rejected")
	}
	if request.AudioInputDevice != "USB microphone" || request.AudioInputProcess != "" || request.AudioInputProcessID != 0 {
		t.Fatalf("device request = %#v", request)
	}
}

func TestBuildLiveAudioInputSwitchRequestForApplication(t *testing.T) {
	input := &CustomWidget.TextValueOption{Text: "Application Audio", Value: Utilities.AudioApplicationOptionValue}
	application := &CustomWidget.TextValueOption{
		Text:  "player.exe - Music (PID 42)",
		Value: Utilities.FormatAudioProcessOptionValue(42, "player.exe"),
	}
	request, ok := buildLiveAudioInputSwitchRequest("request-2", "WASAPI", input, application)
	if !ok {
		t.Fatal("application request was rejected")
	}
	if request.AudioInputDevice != application.Text || request.AudioInputProcess != "player.exe" || request.AudioInputProcessID != 42 {
		t.Fatalf("application request = %#v", request)
	}
}

func TestBuildLiveAudioOutputSwitchRequest(t *testing.T) {
	output := &CustomWidget.TextValueOption{Text: "USB speakers", Value: "7"}
	request, ok := buildLiveAudioOutputSwitchRequest("request-3", "WASAPI", output)
	if !ok {
		t.Fatal("output request was rejected")
	}
	if request.RequestID != "request-3" || request.AudioAPI != "WASAPI" || request.AudioOutputDevice != "USB speakers" {
		t.Fatalf("output request = %#v", request)
	}
}
