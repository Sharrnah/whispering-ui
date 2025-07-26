package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
		Logging.CaptureException(err)
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
			Logging.CaptureException(err)
			return
		}
		if u != nil {
			err := fyne.CurrentApp().OpenURL(u)
			if err != nil {
				Logging.CaptureException(err)
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
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Advanced->CreateAdvancedWindow")
	})

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

	sendErrorReportButton := widget.NewButtonWithIcon(lang.L("Send error report"), theme.MailSendIcon(), func() {
		infoTitle := widget.NewLabel(lang.L("Please enter a description of the error."))
		attachLogCheckbox := widget.NewCheck(lang.L("Attach log file"), nil)
		attachLogCheckbox.SetChecked(true)
		attachHardwareCheckbox := widget.NewCheck(lang.L("Attach hardware information (GPU Memory, GPU Vendor, GPU Adapter)"), nil)
		attachHardwareCheckbox.SetChecked(true)
		emailEntry := widget.NewEntry()
		emailEntry.PlaceHolder = lang.L("E-Mail Address (Optional)")
		textErrorReportInput := widget.NewMultiLineEntry()
		errorReportWindow, _ := Utilities.GetCurrentMainWindow("Error Report")
		errorReportDialog := dialog.NewCustomConfirm(lang.L("Send error report"), lang.L("Send"), lang.L("Cancel"),
			container.NewBorder(container.NewVBox(emailEntry, infoTitle), container.NewVBox(attachLogCheckbox, attachHardwareCheckbox), nil, nil, textErrorReportInput),
			func(confirm bool) {
				if confirm {
					// Send error to Server
					localHub := Logging.CloneHub()
					logfile := "-"
					if attachLogCheckbox.Checked {
						logfile = strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\n")
					}
					localHub.WithScope(func(scope *sentry.Scope) {
						scope.SetTag("report_type", "User Report")
						if !attachHardwareCheckbox.Checked {
							scope.RemoveTag("GPU Memory")
							scope.RemoveTag("GPU Vendor")
							scope.RemoveTag("GPU Adapter")
							scope.RemoveTag("GPU Compute Capability")
						}
						scope.SetUser(sentry.User{Email: emailEntry.Text})
						scope.SetContext("report", map[string]interface{}{
							"log": logfile,
						})
						localHub.CaptureMessage(textErrorReportInput.Text)
					})
					localHub.Flush(Logging.FlushTimeoutDefault)
					dialog.NewInformation(lang.L("Error Report Sent"), lang.L("Your error report has been sent."), errorReportWindow).Show()
				}
			},
			errorReportWindow,
		)
		windowSize := Utilities.GetInlineDialogSize(errorReportWindow, fyne.NewSize(100, 200), fyne.NewSize(200, 200), errorReportWindow.Canvas().Size())
		errorReportDialog.Resize(windowSize)
		errorReportDialog.Show()
	})

	logTabContent := container.NewBorder(nil, container.NewHBox(RestartBackendButton, writeLogFileCheckbox, copyLogButton, sendErrorReportButton), nil, nil, container.NewScroll(Fields.Field.LogText))

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
			Fields.Field.LogText.SetText(strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\r\n") + "\r\n")
		}
	}

	// Log logText updater thread
	go func(writer io.WriteCloser, reader io.Reader) {
		defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
			scope.SetTag("GoRoutine", "Pages\\Advanced->CreateAdvancedWindow#LogTextUpdaterThread")
		})
		for {
			buf := make([]byte, 1024)
			n, err := reader.Read(buf)
			if n > 0 {
				// Append the text to the log text field
				fyne.Do(func() {
					Fields.Field.LogText.Append(string(buf[:n]))
				})
			}
			if err != nil {
				if err == io.EOF {
					break // Exit the loop on EOF
				}
				Logging.CaptureException(err)
				fyne.LogError("Error reading from log stream", err)
				break
			}
		}
	}(RuntimeBackend.BackendsList[0].WriterBackend, RuntimeBackend.BackendsList[0].ReaderBackend)

	return tabs
}
