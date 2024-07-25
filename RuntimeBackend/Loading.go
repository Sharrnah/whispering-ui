package RuntimeBackend

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"strings"
)

type LoadingMessage struct {
	Type string `json:"type"`
	Data struct {
		Name  string `json:"name"`
		Value bool   `json:"value"`
	} `json:"data"`
}

var loadingStates = make(map[string]bool)
var loadingStateContainer *fyne.Container
var loadingStateDialog dialog.Dialog = nil

func InitializeLoadingState() bool {
	// loadingStateDialog already created
	if loadingStateDialog != nil {
		return true
	}
	if fyne.CurrentApp().Driver().AllWindows() == nil || len(fyne.CurrentApp().Driver().AllWindows()) == 0 {
		return false
	}
	statusBar := widget.NewProgressBarInfinite()
	loadingStateContainer = container.NewVBox()
	loadingStateDialog = dialog.NewCustom(
		lang.L("Loading..."),
		lang.L("Hide"),
		container.NewBorder(statusBar, nil, nil, nil, loadingStateContainer),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
	return true
	//LoadingStateContainer.Add(widget.NewLabel(strings.ReplaceAll(loadingMessage.Data.Name, "_", " ")))
}
func ProcessLoadingMessage(line string) bool {
	if !InitializeLoadingState() {
		return false
	}

	var loadingMessage LoadingMessage
	if err := json.Unmarshal([]byte(line), &loadingMessage); err != nil {
		//fmt.Println("Error unmarshalling JSON:", err)
		return false
	}

	// message is not a loading state message
	if loadingMessage.Type != "loading_state" {
		return false
	}

	name := loadingMessage.Data.Name
	value := loadingMessage.Data.Value

	// Update the loading states map
	loadingStates[name] = value

	// Update the loading state container
	updateLoadingStateContainer()

	// Show or hide the dialog based on the current loading states
	if hasActiveLoadingStates() {
		loadingStateDialog.Show()
	} else {
		loadingStateDialog.Hide()
	}
	return true
}

func updateLoadingStateContainer() {
	loadingStateContainer.Objects = nil // Clear current items

	for name, value := range loadingStates {
		if value {
			loadingStateContainer.Add(widget.NewLabel(strings.ReplaceAll(name, "_", " ")))
		}
	}

	loadingStateContainer.Refresh()
}

func hasActiveLoadingStates() bool {
	for _, value := range loadingStates {
		if value {
			return true
		}
	}
	return false
}
