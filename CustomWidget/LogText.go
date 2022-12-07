package CustomWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"strings"
)

type LogText struct {
	Widget    fyne.CanvasObject
	TextLines []string
	MaxLines  int
}

func NewLogText() *LogText {
	return &LogText{
		Widget:    widget.NewLabel(""),
		TextLines: []string{},
		MaxLines:  300,
	}
}

func (l *LogText) GetText() string {
	return strings.Join(l.TextLines, "")
}

func (l *LogText) AppendText(text string) {
	l.TextLines = append(l.TextLines, text)
	if len(l.TextLines) > l.MaxLines {
		l.TextLines = l.TextLines[len(l.TextLines)-l.MaxLines:]
	}
	l.Widget.(*widget.Label).SetText(l.GetText())
}
