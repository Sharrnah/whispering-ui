package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"sync"
)

const defaultPlaceHolder string = "(Select one)"

type TappableSelect struct {
	widget.Select
	popUp                *widget.PopUpMenu
	UpdateBeforeOpenFunc func()
	LastTappedPointEvent *fyne.PointEvent

	impl         fyne.Widget
	propertyLock sync.RWMutex
}

var _ fyne.Widget = (*TappableSelect)(nil)
var _ fyne.Tappable = (*TappableSelect)(nil)

// NewSelect creates a new select widget with the set list of options and changes handler
func NewSelect(options []string, changed func(string)) *TappableSelect {
	s := &TappableSelect{}
	s.OnChanged = changed
	s.Options = options
	s.PlaceHolder = defaultPlaceHolder

	s.ExtendBaseWidget(s)
	return s
}
func (s *TappableSelect) super() fyne.Widget {
	return s
}

func (s *TappableSelect) updateSelected(text string) {
	s.Selected = text

	if s.OnChanged != nil {
		s.OnChanged(s.Selected)
	}

	s.Refresh()
}

func (s *TappableSelect) popUpPos() fyne.Position {
	buttonPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(s.super())
	return buttonPos.Add(fyne.NewPos(0, s.Size().Height-theme.InputBorderSize()))
}
func (s *TappableSelect) showPopUp() {
	items := make([]*fyne.MenuItem, len(s.Options))
	for i := range s.Options {
		text := s.Options[i] // capture
		items[i] = fyne.NewMenuItem(text, func() {
			s.updateSelected(text)
			s.popUp.Hide()
			s.popUp = nil
		})
	}

	c := fyne.CurrentApp().Driver().CanvasForObject(s.super())
	s.popUp = widget.NewPopUpMenu(fyne.NewMenu("", items...), c)

	//s.popUp.alignment = s.Alignment
	s.popUp.ShowAtPosition(s.popUpPos())
	s.popUp.Resize(fyne.NewSize(s.Size().Width, s.popUp.MinSize().Height))
}

func (s *TappableSelect) GetPopup() *widget.PopUpMenu {
	return s.popUp
}

func (s *TappableSelect) ShopPopup() {
	if s.Disabled() {
		return
	}

	s.Refresh()

	s.showPopUp()
}

func (s *TappableSelect) Tapped(ev *fyne.PointEvent) {
	s.LastTappedPointEvent = ev
	s.UpdateBeforeOpenFunc()
}
