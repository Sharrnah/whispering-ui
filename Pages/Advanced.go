package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/UpdateUtility"
	"whispering-tiger-ui/Utilities"
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

	heartImage := canvas.NewImageFromResource(Resources.ResourceHeartPng)
	heartImage.FillMode = canvas.ImageFillContain
	heartImage.ScaleMode = canvas.ImageScaleFastest
	heartImage.SetMinSize(fyne.NewSize(128, 128))
	heartButton := widget.NewButtonWithIcon("Support me on https://ko-fi.com/sharrnah", Resources.ResourceHeartPng, func() {
		u, err := url.Parse("https://ko-fi.com/sharrnah")
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
	supportLabel := widget.NewLabel("If you want to support the development of Whispering Tiger,\nyou can do so by supporting me on Ko-fi.")
	supportLabel.Alignment = fyne.TextAlignCenter

	aboutCard := widget.NewCard("Whispering Tiger UI",
		"Version: "+fyne.CurrentApp().Metadata().Version+" Build: "+strconv.Itoa(fyne.CurrentApp().Metadata().Build),
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

	checkForUpdatesButton := widget.NewButton("Check for updates", func() {
		if !UpdateUtility.VersionCheck(fyne.CurrentApp().Driver().AllWindows()[0], true) {
			dialog.ShowInformation("No update available", "You are running the latest version of Whispering Tiger.", fyne.CurrentApp().Driver().AllWindows()[0])
		}
	})

	updateCheckAtStartupCheckbox := widget.NewCheck("Check for updates at startup", nil)
	updateCheckAtStartupCheckbox.OnChanged = func(b bool) {
		if b {
			fyne.CurrentApp().Preferences().SetBool("CheckForUpdateAtStartup", true)
		} else {
			dialog.ShowConfirm("Disable update check", "Are you sure you want to disable update checks at startup?", func(b bool) {
				if b {
					fyne.CurrentApp().Preferences().SetBool("CheckForUpdateAtStartup", false)
				} else {
					updateCheckAtStartupCheckbox.SetChecked(true)
				}
			}, fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}
	updateCheckAtStartupCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("CheckForUpdateAtStartup", true)

	verticalLayout := container.NewVBox(aboutCard, checkForUpdatesButton, updateCheckAtStartupCheckbox)

	return container.NewCenter(verticalLayout)
}

func CreateAdvancedWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	Settings.Form = Settings.BuildSettingsForm(nil, filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)).(*widget.Form)

	settingsTabContent := container.NewVScroll(Settings.Form)

	RestartBackendButton := widget.NewButton("Restart backend", func() {
		// close running backend process
		if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
			infinityProcessDialog := dialog.NewCustom("Restarting Backend", "OK", container.NewVBox(widget.NewLabel("Restarting backend..."), widget.NewProgressBarInfinite()), fyne.CurrentApp().Driver().AllWindows()[0])
			infinityProcessDialog.Show()
			RuntimeBackend.BackendsList[0].Stop()
			time.Sleep(2 * time.Second)
			RuntimeBackend.BackendsList[0].Start()
			infinityProcessDialog.Hide()
			Fields.Field.SttEnabled.SetChecked(true)
		}
	})

	copyLogButton := widget.NewButtonWithIcon("Copy Log", theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(
			strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\n"),
		)
	})

	writeLogFileCheckbox := widget.NewCheck("Write log file", func(writeLogFile bool) {
		fyne.CurrentApp().Preferences().SetBool("WriteLogfile", writeLogFile)
	})
	writeLogFileCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false)

	logTabContent := container.NewBorder(nil, container.NewHBox(RestartBackendButton, writeLogFileCheckbox, copyLogButton), nil, nil, container.NewScroll(Fields.Field.LogText))

	tabs := container.NewAppTabs(
		container.NewTabItem("About Whispering Tiger", buildAboutInfo()),
		container.NewTabItem("Advanced Settings", settingsTabContent),
		container.NewTabItem("Logs", logTabContent),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Advanced Settings" {
			Settings.Form = Settings.BuildSettingsForm(nil, filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)).(*widget.Form)
			tab.Content.(*container.Scroll).Content = Settings.Form
			tab.Content.(*container.Scroll).Content.Refresh()
			tab.Content.(*container.Scroll).Refresh()
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
