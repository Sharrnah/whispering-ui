package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/url"
	"strconv"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func buildAboutInfo() *fyne.Container {
	aboutImage := canvas.NewImageFromResource(Resources.ResourceAppIconPng)
	aboutImage.FillMode = canvas.ImageFillContain
	aboutImage.ScaleMode = canvas.ImageScaleFastest
	aboutImage.SetMinSize(fyne.NewSize(128, 128))

	aboutCard := widget.NewCard("Whispering Tiger UI",
		"Version: "+fyne.CurrentApp().Metadata().Version+" Build: "+strconv.Itoa(fyne.CurrentApp().Metadata().Build),
		container.NewVBox(
			widget.NewHyperlink("https://github.com/Sharrnah/whispering-ui", parseURL("https://github.com/Sharrnah/whispering-ui")),
			widget.NewHyperlink("https://github.com/Sharrnah/whispering", parseURL("https://github.com/Sharrnah/whispering")),
		),
	)
	aboutCard.SetImage(aboutImage)

	return container.NewCenter(aboutCard)
}

func CreateAdvancedWindow() fyne.CanvasObject {
	Settings.Form = Settings.BuildSettingsForm(nil, Settings.Config.SettingsFilename).(*widget.Form)

	settingsTabContent := container.NewVScroll(Settings.Form)

	logText := CustomWidget.NewLogText()

	logText.Widget.(*widget.Label).Wrapping = fyne.TextWrapWord
	logText.Widget.(*widget.Label).TextStyle = fyne.TextStyle{Monospace: true}

	logTabContent := container.NewVScroll(logText.Widget)

	// Log logText updater thread
	go func(writer io.Writer, reader io.Reader) {
		if reader != nil {
			buffer := make([]byte, 1024)
			for {
				n, err := reader.Read(buffer) // Read from the pipe
				if err != nil {
					//panic(err)
					logText.AppendText(err.Error())
				}
				logText.AppendText(string(buffer[0:n]))
				logTabContent.ScrollToBottom()
			}
		}
	}(RuntimeBackend.BackendsList[0].WriterBackend, RuntimeBackend.BackendsList[0].ReaderBackend)

	tabs := container.NewAppTabs(
		container.NewTabItem("Log", logTabContent),
		container.NewTabItem("Settings", settingsTabContent),
		container.NewTabItem("About", buildAboutInfo()),
	)
	tabs.SetTabLocation(container.TabLocationTrailing)

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Settings" {
			Settings.BuildSettingsForm(nil, Settings.Config.SettingsFilename)
			tab.Content.(*container.Scroll).Content = Settings.Form
		}
	}

	return tabs
}
