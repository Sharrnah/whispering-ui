package CustomWidget

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
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
	// initialize internal state from binding and enforce MaxLines once
	if text, err := data.Get(); err == nil {
		l.setTextTrimmed(text)
	} else {
		l.TextLines = []string{}
	}
	// subscribe to data changes to enforce MaxLines and optional autoscroll
	l.dataListener = binding.NewDataListener(func() {
		if l.Data == nil {
			return
		}
		if text, err := l.Data.Get(); err == nil {
			// Enforce MaxLines locally and update UI immediately; avoid writing back into the binding here
			// to prevent recursive updates and potential delays.
			trimmed := l.normalizeAndTrim(text)
			if l.Entry.Text != trimmed {
				fyne.Do(func() {
					l.Entry.SetText(trimmed)
					l.TextLines = strings.Split(trimmed, "\n")
					if l.AutoScroll {
						l.ScrollToBottom()
					}
				})
			} else if l.AutoScroll {
				fyne.Do(func() { l.ScrollToBottom() })
			}
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
		MaxLines:  10,
	}
	c.MultiLine = true
	c.Wrapping = fyne.TextWrapOff
	c.TextStyle = fyne.TextStyle{Monospace: true}
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
	// Prefer the actual Entry text, which reflects trimming and UI state.
	return l.Entry.Text
}

func (l *LogText) SetText(text string) {
	l.setTextTrimmed(text)
}

func (l *LogText) Append(text string) {
	// Einfaches Anhängen und danach MaxLines durchsetzen.
	l.setTextTrimmed(l.Entry.Text + text)
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

// normalizeAndTrim converts CRLF to LF and limits the text to the last MaxLines lines (if MaxLines > 0).
func (l *LogText) normalizeAndTrim(text string) string {
	// Keine Zeilenenden-Manipulation; nur auf die letzten N Zeilen begrenzen.
	// Besonderheit: Die aktuell „laufende“ (nicht mit \n terminierte) Zeile wird nicht als volle Zeile gezählt,
	// damit sie beim Trimmen nicht verloren geht.
	if l.MaxLines <= 0 {
		return text
	}
	tokens := strings.SplitAfter(text, "\n") // Tokens enthalten das abschließende \n, falls die Zeile abgeschlossen ist.
	if len(tokens) <= l.MaxLines {
		return text
	}
	last := tokens[len(tokens)-1]
	if strings.HasSuffix(last, "\n") {
		// Letzte Zeile ist abgeschlossen -> behalte die letzten MaxLines Zeilentokens
		return strings.Join(tokens[len(tokens)-l.MaxLines:], "")
	}
	// Letzte Zeile ist unvollständig -> behalte (MaxLines-1) vollständige Zeilen + die unvollständige
	completed := tokens[:len(tokens)-1]
	if l.MaxLines-1 <= 0 {
		// nur die unvollständige behalten
		return last
	}
	if len(completed) > l.MaxLines-1 {
		completed = completed[len(completed)-(l.MaxLines-1):]
	}
	return strings.Join(append(completed, last), "")
}

// setTextTrimmed applies text with normalization and line limit enforcement, updates internal state, and auto-scrolls if enabled.
func (l *LogText) setTextTrimmed(text string) {
	trimmed := l.normalizeAndTrim(text)
	// Update UI immediately to avoid perceived missing content while binding events propagate.
	if l.Entry.Text != trimmed {
		l.Entry.SetText(trimmed)
	}
	l.TextLines = strings.Split(trimmed, "\n")
	if l.AutoScroll {
		l.ScrollToBottom()
	}
	// Then synchronize the bound data if needed (no-op if already equal).
	if l.Data != nil {
		if current, err := l.Data.Get(); err != nil || current != trimmed {
			_ = l.Data.Set(trimmed)
		}
	}
}

// (CR wird nicht speziell behandelt – reine Append-Logik und Zeilenlimit.)
