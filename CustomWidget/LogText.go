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

func (l *LogText) Resize(size fyne.Size) {
	l.Entry.Resize(size)
	if l.AutoScroll {
		l.ScrollToBottom()
	}
}

func (l *LogText) Bind(data binding.String) {
	l.Data = data
	l.Entry.Bind(data)
	// Initialize internal state from the binding and enforce MaxLines once
	if text, err := data.Get(); err == nil {
		l.setTextTrimmed(text)
	} else {
		l.TextLines = []string{}
	}
	// Subscribe to data changes to enforce MaxLines and optional auto-scroll
	l.dataListener = binding.NewDataListener(func() {
		if l.Data == nil {
			return
		}
		if text, err := l.Data.Get(); err == nil {
			// Enforce MaxLines locally and update the UI immediately; avoid writing back into the binding here
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
	// Initial scroll if needed
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
		MaxLines:  200,
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
	// Prefer the actual Entry text, which reflects trimming and UI state
	return l.Entry.Text
}

func (l *LogText) SetText(text string) {
	l.setTextTrimmed(text)
}

func (l *LogText) Append(text string) {
	l.setTextTrimmed(l.Entry.Text + text)
}

func (l *LogText) ScrollToBottom() {
	lines := strings.Split(l.Text, "\n")
	if len(lines) == 0 {
		return
	}
	l.CursorRow = len(lines) - 1

	l.Refresh()
}

// TypedKey Override to ignore key input when ReadOnly is true.
func (l *LogText) TypedKey(key *fyne.KeyEvent) {
	if l.ReadOnly {
		return
	}
	// Forward to the embedded Entry handler
	l.Entry.TypedKey(key)
}

// TypedRune Override to ignore text input when ReadOnly is true.
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
	// Forward to the embedded Entry handler
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
	// Limit to the last N logical lines. Line breaks are counted as:
	// - '\n'
	// - a standalone '\r' (CRLF counts as a single separator)
	// The content is NOT modified; only the beginning is trimmed.
	if l.MaxLines <= 0 {
		return text
	}
	// Collect indices (byte positions) of line endings
	breaks := make([]int, 0, 128)
	b := []byte(text)
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' {
			breaks = append(breaks, i)
		} else if b[i] == '\r' {
			// If the next character is \n, treat as CRLF and count only once (at the \n)
			if i+1 < len(b) && b[i+1] == '\n' {
				continue
			}
			breaks = append(breaks, i)
		}
	}
	if len(breaks) == 0 {
		// No separators -> a single line; nothing to trim
		return text
	}
	// Check if the last line is complete (ends with \n or \r, including CRLF)
	lastComplete := false
	if len(b) > 0 {
		switch b[len(b)-1] {
		case '\n', '\r':
			lastComplete = true
		}
	}
	// The number of complete lines equals len(breaks).
	// Keep the last MaxLines complete lines (+ optionally the trailing partial line).
	startIdx := 0
	if lastComplete {
		if len(breaks) > l.MaxLines {
			// Cut after the (len(breaks)-MaxLines)-th separator
			cutAt := breaks[len(breaks)-l.MaxLines] + 1
			if cutAt > startIdx {
				startIdx = cutAt
			}
		}
	} else {
		// The last line is incomplete; it counts additionally
		if l.MaxLines-1 > 0 && len(breaks) > (l.MaxLines-1) {
			cutAt := breaks[len(breaks)-(l.MaxLines-1)] + 1
			if cutAt > startIdx {
				startIdx = cutAt
			}
		} else if l.MaxLines-1 <= 0 && len(breaks) > 0 {
			// Keep only the incomplete trailing line
			cutAt := breaks[len(breaks)-1] + 1
			if cutAt > startIdx {
				startIdx = cutAt
			}
		}
	}
	if startIdx <= 0 {
		return text
	}
	return text[startIdx:]
}

// setTextTrimmed applies text with normalization and line limit enforcement, updates internal state, and auto-scrolls if enabled.
func (l *LogText) setTextTrimmed(text string) {
	trimmed := l.normalizeAndTrim(text)
	// Update the UI immediately to avoid perceived missing content while binding events propagate
	if l.Entry.Text != trimmed {
		l.Entry.SetText(trimmed)
	}
	l.TextLines = strings.Split(trimmed, "\n")
	if l.AutoScroll {
		l.ScrollToBottom()
	}
	// Then synchronize the bound data if needed (no-op if already equal)
	if l.Data != nil {
		if current, err := l.Data.Get(); err != nil || current != trimmed {
			_ = l.Data.Set(trimmed)
		}
	}
}

// Note: CR is not handled specially here — this is pure append logic with line limit enforcement
