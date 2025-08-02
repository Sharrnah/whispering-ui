package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"strings"
)

type LogText struct {
	widget.Entry
	TextLines    []string
	MaxLines     int
	AutoScroll   bool                 // If true, automatically scroll to the bottom when new text is added
	ReadOnly     bool                 // if true, LogText is read-only
	Data         binding.String       // bound data source
	dataListener binding.DataListener // listener for data change events
}

//
//func (l *LogText) MinSize() fyne.Size {
//	return l.Entry.MinSize()
//}
//
//func (l *LogText) Move(position fyne.Position) {
//	l.Entry.Move(position)
//}
//
//func (l *LogText) Position() fyne.Position {
//	return l.Entry.Position()
//}

func (l *LogText) Resize(size fyne.Size) {
	l.Entry.Resize(size)
	if l.AutoScroll {
		l.ScrollToBottom()
	}
}

//
//func (l *LogText) Size() fyne.Size {
//	return l.Entry.Size()
//}
//
//func (l *LogText) Hide() {
//	l.Entry.Hide()
//}
//
//func (l *LogText) Visible() bool {
//	return l.Entry.Visible()
//}
//
//func (l *LogText) Show() {
//	l.Entry.Show()
//}
//
//func (l *LogText) Refresh() {
//	l.Entry.Refresh()
//}

func (l *LogText) Bind(data binding.String) {
	l.Data = data
	l.Entry.Bind(data)
	// initialize internal lines from binding
	if text, err := data.Get(); err == nil {
		l.TextLines = strings.Split(text, "\n")
	}
	// subscribe to data changes and trigger scroll when AutoScroll is enabled
	l.dataListener = binding.NewDataListener(func() {
		if l.AutoScroll {
			fyne.Do(func() {
				l.ScrollToBottom()
			})
		}
	})
	data.AddListener(l.dataListener)
	// initial scroll if needed
	if l.AutoScroll {
		l.ScrollToBottom()
	}
}
func (l *LogText) Unbind() {
	if l.Data != nil && l.dataListener != nil {
		l.Data.RemoveListener(l.dataListener)
		l.dataListener = nil
	}
	l.Entry.Unbind()
	l.Data = nil
}

func NewLogText() *LogText {
	c := &LogText{
		TextLines: []string{},
		MaxLines:  100,
	}
	c.MultiLine = true
	c.Wrapping = fyne.TextWrapOff
	c.TextStyle = fyne.TextStyle{
		Monospace: true,
	}
	c.ExtendBaseWidget(c)
	return c
}

func NewLogTextWithData(data binding.String) *LogText {
	logText := NewLogText()
	logText.Bind(data)
	logText.Validator = nil
	return logText
}

func (l *LogText) GetText() string {
	if l.Data != nil {
		if text, err := l.Data.Get(); err == nil {
			return text
		}
	}
	return strings.Join(l.TextLines, "")
}

func (l *LogText) SetText(text string) {
	l.TextLines = strings.Split(text, "\n")
	if len(l.TextLines) > l.MaxLines {
		l.TextLines = l.TextLines[len(l.TextLines)-l.MaxLines:]
	}
	if l.Data != nil {
		_ = l.Data.Set(text)
	}
	l.Entry.SetText(text)
	if l.AutoScroll {
		l.ScrollToBottom()
	}
}

func (l *LogText) Append(text string) {
	l.TextLines = append(l.TextLines, text)
	if len(l.TextLines) > l.MaxLines {
		l.TextLines = l.TextLines[len(l.TextLines)-l.MaxLines:]
	}
	l.Entry.Append(text)
	if l.Data != nil {
		// update bound data to current full content
		_ = l.Data.Set(l.Entry.Text)
	}
	if l.AutoScroll {
		l.ScrollToBottom()
	}
}

func (l *LogText) ScrollToBottom() {
	lines := strings.Split(l.Text, "\n")
	if len(lines) == 0 {
		return
	}
	//lastLine := lines[len(lines)-1]
	l.CursorRow = len(lines) - 1
	//entry.CursorColumn = len([]rune(lastLine))
	l.Refresh()
}

// Override TypedKey to ignore key input when ReadOnly is true.
func (l *LogText) TypedKey(key *fyne.KeyEvent) {
	if l.ReadOnly {
		return
	}
	// forward to the embedded Entry handler
	l.Entry.TypedKey(key)
}

// Override TypedRune to ignore text input when ReadOnly is true.
func (l *LogText) TypedRune(r rune) {
	if l.ReadOnly {
		return
	}
	l.Entry.TypedRune(r)
}

// TappedSecondary Override to remove Cut/Paste options when ReadOnly is true
func (l *LogText) TappedSecondary(pe *fyne.PointEvent) {
	if l.ReadOnly {
		clipboard := fyne.CurrentApp().Clipboard()
		copyItem := fyne.NewMenuItem(lang.L("Copy"), func() {
			l.TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
		})
		selectAllItem := fyne.NewMenuItem(lang.L("Select all"), func() {
			l.TypedShortcut(&fyne.ShortcutSelectAll{})
		})
		entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(l)
		popUpPos := entryPos.Add(pe.Position)
		c := fyne.CurrentApp().Driver().CanvasForObject(l)
		menu := fyne.NewMenu("", copyItem, selectAllItem)
		popup := widget.NewPopUpMenu(menu, c)
		popup.ShowAtPosition(popUpPos)
		return
	}
	// forward to the embedded Entry handler
	l.Entry.TappedSecondary(pe)
}

// TypedShortcut Override to ignore Cut (CTRL+X) and Paste (CTRL+V) when ReadOnly is true.
func (l *LogText) TypedShortcut(s fyne.Shortcut) {
	if l.ReadOnly {
		switch s.(type) {
		case *fyne.ShortcutCut, *fyne.ShortcutPaste:
			return
		}
	}
	l.Entry.TypedShortcut(s)
}
