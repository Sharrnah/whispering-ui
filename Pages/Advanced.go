package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func buildAboutInfo() fyne.CanvasObject {
	aboutImage := canvas.NewImageFromResource(Resources.ResourceAppIconPng)
	aboutImage.FillMode = canvas.ImageFillContain
	aboutImage.ScaleMode = canvas.ImageScaleFastest
	aboutImage.SetMinSize(fyne.NewSize(128, 128))

	heartImage := canvas.NewImageFromResource(Resources.ResourceHeartPng)
	heartImage.FillMode = canvas.ImageFillContain
	heartImage.ScaleMode = canvas.ImageScaleFastest
	heartImage.SetMinSize(fyne.NewSize(128, 128))

	heartButton := widget.NewButtonWithIcon(lang.L("Support me on Ko-Fi", map[string]interface{}{"KofiUrl": lang.L("KofiUrl")}), Resources.ResourceHeartPng, func() {
		u, err := url.Parse(lang.L("KofiUrl"))
		if err != nil {
			return
		}
		if u != nil {
			err := fyne.CurrentApp().OpenURL(u)
			if err != nil {
				fyne.LogError("Failed to open url", err)
			}
		}
	})
	supportLabel := widget.NewLabel(lang.L("If you want to support the development of Whispering Tiger"))
	supportLabel.Alignment = fyne.TextAlignCenter

	aboutCard := widget.NewCard("Whispering Tiger UI",
		lang.L("VersionBuild", map[string]interface{}{"Version": Utilities.AppVersion, "Build": Utilities.AppBuild}),
		container.NewVBox(
			widget.NewHyperlink("https://whispering-tiger.github.io", parseURL("https://whispering-tiger.github.io")),
			widget.NewAccordion(
				widget.NewAccordionItem("Repositories",
					container.NewVBox(
						widget.NewHyperlink("https://github.com/Sharrnah/whispering-ui", parseURL("https://github.com/Sharrnah/whispering-ui")),
						widget.NewHyperlink("https://github.com/Sharrnah/whispering", parseURL("https://github.com/Sharrnah/whispering")),
					),
				),
			),
			widget.NewSeparator(),
			supportLabel,
			heartButton,
			widget.NewSeparator(),
		),
	)
	aboutCard.SetImage(aboutImage)

	verticalLayout := container.NewVBox(aboutCard)

	return container.NewVScroll(container.NewCenter(verticalLayout))
}

func CreateAdvancedWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	Settings.Form = Settings.BuildSettingsForm(nil, filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)).(*widget.Form)

	settingsTabContent := container.NewVScroll(Settings.Form)

	RestartBackendButton := widget.NewButton(lang.L("Restart backend"), func() {
		RuntimeBackend.RestartBackend(false, lang.L("Are you sure you want to restart the backend?"))
	})

	copyLogButton := widget.NewButtonWithIcon(lang.L("Copy Log"), theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(
			strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\n"),
		)
	})

	writeLogFileCheckbox := widget.NewCheck(lang.L("Write log file"), func(writeLogFile bool) {
		fyne.CurrentApp().Preferences().SetBool("WriteLogfile", writeLogFile)
	})
	writeLogFileCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false)

	logTabContent := container.NewBorder(nil, container.NewHBox(RestartBackendButton, writeLogFileCheckbox, copyLogButton), nil, nil, container.NewScroll(Fields.Field.LogText))

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("About Whispering Tiger"), buildAboutInfo()),
		container.NewTabItem(lang.L("Advanced Settings"), settingsTabContent),
		container.NewTabItem(lang.L("Logs"), logTabContent),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == lang.L("Advanced Settings") {
			Settings.Form = Settings.BuildSettingsForm(nil, filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)).(*widget.Form)
			tab.Content.(*container.Scroll).Content = Settings.Form
			tab.Content.(*container.Scroll).Content.Refresh()
			tab.Content.(*container.Scroll).Refresh()
		}
		if tab.Text == lang.L("Logs") {
			Fields.Field.LogText.SetText("")
			Fields.Field.LogText.Write([]byte(strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\r\n") + "\r\n"))
		}
	}

	// Log logText updater thread
	Fields.Field.LogText.Resize(fyne.NewSize(1200, 800))
	go func(writer io.WriteCloser, reader io.Reader) {
		defer Utilities.PanicLogger()
		_ = Fields.Field.LogText.RunWithConnection(writer, reader)
	}(RuntimeBackend.BackendsList[0].WriterBackend, RuntimeBackend.BackendsList[0].ReaderBackend)

	return tabs
}
