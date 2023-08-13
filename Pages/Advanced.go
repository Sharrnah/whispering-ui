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
	"strconv"
	"strings"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/UpdateUtility"
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

	settingsLabel := widget.NewLabel("\nExperimental Flags:")
	// UI downloading Flag
	disableUiDownloadsCheckbox := widget.NewCheck("Disable experimental UI downloading of AI Models.", nil)
	disableUiDownloadsCheckbox.OnChanged = func(b bool) {
		fyne.CurrentApp().Preferences().SetBool("DisableUiDownloads", b)
	}
	disableUiDownloadsCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("DisableUiDownloads", false)

	// refocus flag
	autoRefocusCheckbox := widget.NewCheck("focus window on message receive (can improve speed in VR)", nil)
	autoRefocusCheckbox.OnChanged = func(b bool) {
		fyne.CurrentApp().Preferences().SetBool("AutoRefocusWindow", b)
	}
	autoRefocusCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("AutoRefocusWindow", false)

	verticalLayout := container.NewVBox(aboutCard, checkForUpdatesButton, updateCheckAtStartupCheckbox, settingsLabel, disableUiDownloadsCheckbox, autoRefocusCheckbox)

	return container.NewCenter(verticalLayout)
}

func CreateAdvancedWindow() fyne.CanvasObject {
	Settings.Form = Settings.BuildSettingsForm(nil, Settings.Config.SettingsFilename).(*widget.Form)

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
			Settings.Form = Settings.BuildSettingsForm(nil, Settings.Config.SettingsFilename).(*widget.Form)
			tab.Content.(*container.Scroll).Content = Settings.Form
			tab.Content.(*container.Scroll).Content.Refresh()
			tab.Content.(*container.Scroll).Refresh()
		}
	}

	// Log logText updater thread
	Fields.Field.LogText.Resize(fyne.NewSize(1200, 800))
	go func(writer io.WriteCloser, reader io.Reader) {
		_ = Fields.Field.LogText.RunWithConnection(writer, reader)
	}(RuntimeBackend.BackendsList[0].WriterBackend, RuntimeBackend.BackendsList[0].ReaderBackend)

	return tabs
}
