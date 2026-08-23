package CustomWidget

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
)

func TestCompletionEntryPopupStaysInsideCanvas(t *testing.T) {
	app := fynetest.NewApp()
	defer app.Quit()

	options := make([]string, 30)
	for i := range options {
		options[i] = fmt.Sprintf("Option %d", i)
	}

	entry := NewCompletionEntry(options)
	entry.Resize(fyne.NewSize(180, 40))
	entry.Move(fyne.NewPos(20, 150))

	content := container.NewWithoutLayout(entry)
	window := fynetest.NewWindow(content)
	window.Resize(fyne.NewSize(240, 200))
	defer window.Close()

	entry.ShowCompletion()

	if entry.popupMenu == nil {
		t.Fatal("ShowCompletion did not create a popup")
	}
	_, areaSize := window.Canvas().InteractiveArea()
	popupPos := entry.popupMenu.Position()
	popupSize := entry.popupMenu.Size()
	if popupPos.Y != 0 {
		t.Fatalf("popup top = %v, want 0 when the complete list is taller than the canvas", popupPos.Y)
	}
	if popupSize.Height != areaSize.Height {
		t.Fatalf("popup height = %v, want canvas-height scroll viewport %v", popupSize.Height, areaSize.Height)
	}
	if popupSize.Height >= entry.itemHeight*float32(len(options)) {
		t.Fatalf("popup height = %v, want a constrained scrollable viewport", popupSize.Height)
	}
}

func TestCompletionEntryPopupMovesUpWhenItDoesNotFitBelow(t *testing.T) {
	app := fynetest.NewApp()
	defer app.Quit()

	entry := NewCompletionEntry([]string{"One", "Two", "Three"})
	entry.Resize(fyne.NewSize(180, 40))
	entry.Move(fyne.NewPos(20, 150))

	content := container.NewWithoutLayout(entry)
	window := fynetest.NewWindow(content)
	window.Resize(fyne.NewSize(240, 200))
	defer window.Close()

	entry.ShowCompletion()

	popupPos := entry.popupMenu.Position()
	popupSize := entry.popupMenu.Size()
	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(entry)
	if popupPos.Y >= entryPos.Y+entry.Size().Height {
		t.Fatalf("popup top = %v, want it shifted above its below-entry position", popupPos.Y)
	}
	_, areaSize := window.Canvas().InteractiveArea()
	if popupPos.Y+popupSize.Height > areaSize.Height {
		t.Fatalf("popup bottom = %v, exceeds canvas height %v", popupPos.Y+popupSize.Height, areaSize.Height)
	}
}
