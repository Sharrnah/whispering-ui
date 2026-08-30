package Pages

import (
	"sync/atomic"
	"whispering-tiger-ui/Utilities"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type audioThresholdMeter struct {
	container *fyne.Container
	marker    *canvas.Rectangle
	threshold atomic.Int64
}

type audioThresholdLayout struct {
	threshold *atomic.Int64
}

func energyMeterPosition(energy int) float32 {
	if energy <= 0 {
		return 0
	}
	return float32(normalizedAudioMeterLevel(Utilities.EnergyThresholdAmplitude(float64(energy))) / audioMeterMax)
}

func (l *audioThresholdLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
	markerWidth := float32(3)
	position := energyMeterPosition(int(l.threshold.Load()))
	markerX := position * (size.Width - markerWidth)
	objects[1].Move(fyne.NewPos(markerX, 0))
	objects[1].Resize(fyne.NewSize(markerWidth, size.Height))
}

func (l *audioThresholdLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	return objects[0].MinSize()
}

func newAudioThresholdMeter(level *widget.ProgressBar, energy int) *audioThresholdMeter {
	marker := canvas.NewRectangle(theme.Color(theme.ColorNameWarning))
	meter := &audioThresholdMeter{marker: marker}
	meter.threshold.Store(int64(energy))
	meter.container = container.New(
		&audioThresholdLayout{threshold: &meter.threshold}, level, marker,
	)
	if energy <= 0 {
		marker.Hide()
	}
	return meter
}

func (m *audioThresholdMeter) CanvasObject() fyne.CanvasObject {
	return m.container
}

func (m *audioThresholdMeter) SetThreshold(energy int) {
	m.threshold.Store(int64(energy))
	if energy <= 0 {
		m.marker.Hide()
	} else {
		m.marker.Show()
	}
	m.container.Refresh()
}
