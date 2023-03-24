package Messages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"strings"
)

type LoadingState struct {
	States map[string]bool `json:"data"`
}

var CurrentLoadingState LoadingState
var LoadingStateDialog dialog.Dialog = nil
var LoadingStateContainer *fyne.Container = nil

func (res LoadingState) Update() *LoadingState {
	if LoadingStateDialog == nil {
		res.InitStateWindow()
	}

	if len(res.States) == 0 {
		LoadingStateDialog.Hide()
		return &res
	}

	LoadingStateContainer.RemoveAll()
	showLoading := false
	for name, value := range CurrentLoadingState.States {
		if value {
			LoadingStateContainer.Add(widget.NewLabel(strings.ReplaceAll(name, "_", " ")))
			showLoading = true
		}
	}
	if showLoading {
		LoadingStateDialog.Show()
	} else {
		LoadingStateDialog.Hide()
	}

	return &res
}

func (res LoadingState) GetState(name string) bool {
	return res.States[name]
}

func (res LoadingState) SetState(name string, state bool) {
	res.States[name] = state
}

func (res *LoadingState) InitStateWindow() {
	statusBar := widget.NewProgressBarInfinite()
	LoadingStateContainer = container.NewVBox()

	LoadingStateDialog = dialog.NewCustom(
		"Loading...",
		"Hide",
		container.NewBorder(statusBar, nil, nil, nil, LoadingStateContainer),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
}
