package Messages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type ExceptionMessage struct {
	Type         string `json:"type"`
	ErrorMessage string `json:"data"`
}

func (res ExceptionMessage) Error() string {
	return res.ErrorMessage
}

func (res ExceptionMessage) ShowError(window fyne.Window) {
	errorDialog := dialog.NewError(
		res,
		window,
	)
	errorDialog.Show()
}

func (res ExceptionMessage) ShowInfo(window fyne.Window) {
	infoDialog := dialog.NewInformation(
		"Info",
		res.ErrorMessage,
		window,
	)
	infoDialog.Show()
}
