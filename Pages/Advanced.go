package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"io"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
)

func CreateAdvancedWindow() fyne.CanvasObject {
	Settings.Form = Settings.BuildSettingsForm(nil).(*widget.Form)

	settingsTabContent := container.NewVScroll(Settings.Form)

	logText := CustomWidget.NewLogText()

	logText.Widget.(*widget.Label).Wrapping = fyne.TextWrapWord
	logText.Widget.(*widget.Label).TextStyle = fyne.TextStyle{Monospace: true}

	logTabContent := container.NewVScroll(logText.Widget)

	// Log logText updater thread
	go func(writer io.Writer, reader io.Reader) {
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer) // Read from the pipe
			if err != nil {
				panic(err)
			}
			logText.AppendText(string(buffer[0:n]))
			logTabContent.ScrollToBottom()
		}
	}(RuntimeBackend.WriterBackend, RuntimeBackend.ReaderBackend)

	tabs := container.NewAppTabs(
		container.NewTabItem("Log", logTabContent),
		container.NewTabItem("Settings", settingsTabContent),
	)
	tabs.SetTabLocation(container.TabLocationTrailing)

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Settings" {
			Settings.BuildSettingsForm(nil)
			tab.Content.(*container.Scroll).Content = Settings.Form
		}
	}

	return tabs
}
