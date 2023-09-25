package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TextValueSelect struct {
	widget.Select
	Name      string
	Options   []TextValueOption
	OnChanged func(TextValueOption) `json:"-"`
}

type TextValueOption struct {
	// Text sets the text of the select.
	Text string
	// Value sets the value of the select.
	Value string
}

func (s *TextValueOption) String() string {
	return s.Value
}

// NewTextValueSelect Creates a new TextValue Select widget
// Set defaultIndex to -1 if no default should be selected
func NewTextValueSelect(name string, options []TextValueOption, changed func(TextValueOption), defaultIndex int) *TextValueSelect {
	s := &TextValueSelect{}
	s.Name = name
	s.OnChanged = changed
	s.Select.OnChanged = func(text string) {
		s.updateSelected(text)
	}
	s.Options = options
	s.PlaceHolder = defaultPlaceHolder

	if defaultIndex > -1 && defaultIndex <= len(s.Options) {
		s.Selected = s.Options[defaultIndex].Text
	}

	s.ExtendBaseWidget(s)
	return s
}

// ClearSelected clears the current option of the select widget.  After
// clearing the current option, the Select widget's PlaceHolder will
// be displayed.
func (s *TextValueSelect) ClearSelected() {
	s.updateSelected("")
}

func (s *TextValueSelect) GetName() string {
	return s.Name
}

// SelectedIndex returns the index value of the currently selected item in Options list.
// It will return -1 if there is no selection.
func (s *TextValueSelect) SelectedIndex() int {
	for i, option := range s.Options {
		if s.Selected == option.Text {
			return i
		}
	}
	return -1 // not selected/found
}

func (s *TextValueSelect) GetSelected() *TextValueOption {
	selectedIndex := s.SelectedIndex()
	if selectedIndex == -1 {
		return nil
	}
	return &s.Options[s.SelectedIndex()]
}

// SetSelected sets the current option of the select widget
func (s *TextValueSelect) SetSelected(value string) {
	for _, option := range s.Options {
		if value == option.Value {
			s.updateSelected(option.Text)
		}
	}
}

// SetSelectedIndex will set the Selected option from the value in Options list at index position.
func (s *TextValueSelect) SetSelectedIndex(index int) {
	if index < 0 || index >= len(s.Options) {
		return
	}

	s.updateSelected(s.Options[index].Text)
}

func (s *TextValueSelect) updateSelected(text string) {
	var lastSelected TextValueOption
	for i := 0; i < len(s.Options); i++ {
		if s.Options[i].Text == text {
			s.Selected = s.Options[i].Text
			lastSelected = s.Options[i]
		}
	}

	if s.OnChanged != nil {
		s.OnChanged(lastSelected)
	}

	s.Refresh()
}

func (s *TextValueSelect) ContainsEntry(entry *TextValueOption) bool {
	for i := 0; i < len(s.Options); i++ {
		if entry != nil && s.Options[i].Value == entry.Value {
			return true
		}
	}
	return false
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (s *TextValueSelect) Tapped(tapEvent *fyne.PointEvent) {
	// copy options over to base widget
	var items []string
	for i := range s.Options {
		items = append(items, s.Options[i].Text)
	}
	s.Select.Options = items

	s.Select.Tapped(tapEvent)

}
