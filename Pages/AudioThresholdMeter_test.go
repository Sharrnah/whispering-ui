package Pages

import (
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestAudioThresholdMeterDoesNotGateVisibleActivity(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	level := widget.NewProgressBar()
	level.Max = audioMeterMax
	level.SetValue(20)
	meter := newAudioThresholdMeter(level, 300)
	meter.container.Resize(fyne.NewSize(100, 20))

	if level.Value != 20 {
		t.Fatalf("threshold changed visible input level to %.1f, want 20", level.Value)
	}
	initialMarker := meter.marker.Position().X
	if initialMarker <= 0 || !meter.marker.Visible() {
		t.Fatalf("enabled threshold marker is not visible at a useful position: %.1f", initialMarker)
	}

	meter.SetThreshold(1000)
	if meter.marker.Position().X <= initialMarker {
		t.Fatalf("higher threshold did not move marker right: %.1f <= %.1f", meter.marker.Position().X, initialMarker)
	}
	meter.SetThreshold(0)
	if meter.marker.Visible() {
		t.Fatal("disabled volume trigger should hide its threshold marker")
	}
}
