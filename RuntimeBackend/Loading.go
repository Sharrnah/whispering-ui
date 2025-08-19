package RuntimeBackend

import (
	"encoding/json"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

type LoadingMessage struct {
	Type string `json:"type"`
	Data struct {
		Name  string `json:"name"`
		Value bool   `json:"value"`
	} `json:"data"`
}

var loadingStates = make(map[string]bool)
var loadingStatesMu sync.RWMutex
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

	// Update the loading states map (thread-safe)
	loadingStatesMu.Lock()
	loadingStates[name] = value
	// Build a snapshot of active names to render without holding the lock during UI ops
	activeNames := make([]string, 0, len(loadingStates))
	for n, v := range loadingStates {
		if v {
			activeNames = append(activeNames, n)
		}
	}
	loadingStatesMu.Unlock()

	// Update the loading state container atomically in a single UI pass
	updateLoadingStateContainer(activeNames)

	// Show or hide the dialog based on the current loading states
	if len(activeNames) > 0 {
		fyne.Do(func() { loadingStateDialog.Show() })
	} else {
		fyne.Do(func() { loadingStateDialog.Hide() })
	}
	return true
}

func updateLoadingStateContainer(activeNames []string) {
	fyne.Do(func() {
		// Replace the entire content in one go to avoid duplicates
		loadingStateContainer.Objects = nil
		for _, name := range activeNames {
			loadingStateContainer.Add(widget.NewLabel(strings.ReplaceAll(name, "_", " ")))
		}
		loadingStateContainer.Refresh()
	})
}
