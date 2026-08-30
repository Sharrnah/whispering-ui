package Fields

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestRealtimeResultDisplayCapsLongTextAtThreeLines(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)

	data := binding.NewString()
	label, scroll := newRealtimeResultDisplay(data)
	oneLineHeight := scroll.MinSize().Height
	scroll.Resize(fyne.NewSize(240, oneLineHeight))
	label.Resize(fyne.NewSize(240, oneLineHeight))

	if err := data.Set("first line\nsecond line"); err != nil {
		t.Fatal(err)
	}
	fyne.DoAndWait(func() {})

	twoLineProbe := widget.NewLabel("M\nM")
	twoLineProbe.TextStyle = label.TextStyle
	if got, want := scroll.MinSize().Height, twoLineProbe.MinSize().Height; got != want {
		t.Fatalf("two-line realtime viewport height = %v, want %v", got, want)
	}

	if err := data.Set(strings.Repeat("hallucination ", 500)); err != nil {
		t.Fatal(err)
	}
	fyne.DoAndWait(func() {})

	lineProbe := widget.NewLabel(strings.Repeat("M\n", realtimeResultMaxLines-1) + "M")
	lineProbe.TextStyle = label.TextStyle
	maxHeight := lineProbe.MinSize().Height
	if got := scroll.MinSize().Height; got != maxHeight {
		t.Fatalf("realtime viewport height = %v, want three-line cap %v", got, maxHeight)
	}
	if maxHeight <= oneLineHeight {
		t.Fatalf("three-line cap %v did not exceed one-line height %v", maxHeight, oneLineHeight)
	}
	if label.MinSize().Height <= maxHeight {
		t.Fatalf("test text did not overflow the viewport: label=%v, viewport=%v", label.MinSize().Height, maxHeight)
	}
	if scroll.Offset.Y <= 0 {
		t.Fatalf("realtime viewport did not follow the newest text: offset=%v", scroll.Offset)
	}
}
