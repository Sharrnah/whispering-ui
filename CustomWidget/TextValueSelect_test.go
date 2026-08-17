package CustomWidget

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestTextValueSelectKeepsCanonicalSelection(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	changeCount := 0
	selection := NewTextValueSelect("translator", []TextValueOption{
		{Text: "NLLB", Value: "nllb"},
		{Text: "MiLMMT", Value: "milmmt"},
	}, func(TextValueOption) {
		changeCount++
	}, 0)

	if len(selection.Select.Options) != len(selection.Options) {
		t.Fatalf("embedded options = %d, value options = %d", len(selection.Select.Options), len(selection.Options))
	}

	selection.SetSelected("milmmt")
	validChangeCount := changeCount
	selection.Select.Selected = ""
	selection.Select.OnChanged("")

	selected := selection.GetSelected()
	if selected == nil || selected.Value != "milmmt" {
		t.Fatalf("selection after stale event = %#v, want milmmt", selected)
	}
	if changeCount != validChangeCount {
		t.Fatalf("stale event invoked value callback: got %d calls, want %d", changeCount, validChangeCount)
	}

	selection.ClearSelected()
	if selection.GetSelected() != nil {
		t.Fatalf("ClearSelected left selection %#v", selection.GetSelected())
	}
}
