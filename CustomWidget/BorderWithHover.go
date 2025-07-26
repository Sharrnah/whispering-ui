package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// HoverableBorder wraps a fyne.CanvasObject and detects mouse hover events.
type HoverableBorder struct {
	widget.BaseWidget
	content      *fyne.Container
	onMouseEnter func()
	onMouseLeave func()
}

func NewHoverableBorder(content *fyne.Container, onEnter, onLeave func()) *HoverableBorder {
	hb := &HoverableBorder{content: content, onMouseEnter: onEnter, onMouseLeave: onLeave}
	hb.ExtendBaseWidget(hb)
	return hb
}

func (hb *HoverableBorder) GetContainer() *fyne.Container {
	return hb.content
}

func (hb *HoverableBorder) SetOnMouseEnter(fn func()) {
	hb.onMouseEnter = fn
}
func (hb *HoverableBorder) SetOnMouseLeave(fn func()) {
	hb.onMouseLeave = fn
}

func (hb *HoverableBorder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(hb.content)
}

// Implement desktop.Hoverable.
func (hb *HoverableBorder) MouseIn(e *desktop.MouseEvent) {
	if hb.onMouseEnter != nil {
		hb.onMouseEnter()
	}
}

func (hb *HoverableBorder) MouseOut() {
	if hb.onMouseLeave != nil {
		hb.onMouseLeave()
	}
}

func (hb *HoverableBorder) MouseMoved(e *desktop.MouseEvent) {
	// Mouse moved over widget.
}
