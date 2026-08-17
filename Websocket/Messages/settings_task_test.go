package Messages

import (
	"testing"

	"whispering-tiger-ui/CustomWidget"
)

func TestResolveWhisperTaskSelectionKeepsCombinedTaskExact(t *testing.T) {
	options := []CustomWidget.TextValueOption{
		{Text: "transcribe", Value: "transcribe"},
		{Text: "translate", Value: "translate"},
		{Text: "transcribe & translate", Value: "transcribe_translate"},
	}
	if got := resolveWhisperTaskSelection("transcribe_translate", options); got != "transcribe_translate" {
		t.Fatalf("combined task was not retained: %q", got)
	}
}

func TestResolveWhisperTaskSelectionDoesNotGuessUnavailableCombinedTask(t *testing.T) {
	options := []CustomWidget.TextValueOption{
		{Text: "transcribe", Value: "transcribe"},
		{Text: "translate", Value: "translate"},
	}
	if got := resolveWhisperTaskSelection("transcribe_translate", options); got != "" {
		t.Fatalf("unavailable combined task should not be reduced to another task: %q", got)
	}
}
