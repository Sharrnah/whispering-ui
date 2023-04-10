package Pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/UpdateUtilitiy"
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
		if !UpdateUtilitiy.VersionCheck(fyne.CurrentApp().Driver().AllWindows()[0], true) {
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

func GetClassNameOfPlugin(path string) string {
	// Define the regular expression
	re := regexp.MustCompile(`class\s+(\w+)\(Plugins\.Base\)`)

	// Open the file and read its contents
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	// Convert the byte slice to a string
	contents := string(data)

	// Find the first match
	match := re.FindStringSubmatch(contents)

	// Extract the className
	if len(match) > 1 {
		className := match[1]
		return className
	}
	return ""
}

func CreatePluginSettingsPage() fyne.CanvasObject {

	// build plugins list
	var pluginFiles []string
	var pluginFilesAccordionItems []*widget.AccordionItem
	files, err := os.ReadDir("./Plugins")
	if err != nil {
		println(err)
	}
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "__init__") && (strings.HasSuffix(file.Name(), ".py")) {
			pluginFiles = append(pluginFiles, file.Name())
			pluginSettings := container.NewVBox()
			pluginClassName := GetClassNameOfPlugin("./Plugins/" + file.Name())

			// plugin enabled checkbox
			pluginEnabledCheckbox := widget.NewCheck(pluginClassName+" enabled", func(enabled bool) {
				Settings.Config.Plugins[pluginClassName] = enabled
				sendMessage := Fields.SendMessageStruct{
					Type:  "setting_change",
					Name:  "plugins",
					Value: Settings.Config.Plugins,
				}
				sendMessage.SendMessage()
			})
			pluginEnabledCheckbox.Checked = Settings.Config.Plugins[pluginClassName]
			pluginSettings.Add(pluginEnabledCheckbox)

			// plugin settings
			pluginSettingsForm := widget.NewMultiLineEntry()

			if Settings.Config.Plugin_settings != nil {
				if settings, ok := Settings.Config.Plugin_settings.(map[string]interface{})[pluginClassName]; ok {
					if settingsMap, ok := settings.(map[string]interface{}); ok {
						settingsStr, err := yaml.Marshal(settingsMap)
						if err != nil {
							println(err)
						}
						pluginSettingsForm.SetText(string(settingsStr))
					}
				}
			}
			pluginSettingsForm.OnChanged = func(text string) {
				var settingsMap map[string]interface{}
				err := yaml.Unmarshal([]byte(text), &settingsMap)
				if err != nil {
					println(err)
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				} else {
					Settings.Config.Plugin_settings.(map[string]interface{})[pluginClassName] = settingsMap
					sendMessage := Fields.SendMessageStruct{
						Type:  "setting_change",
						Name:  "plugin_settings",
						Value: Settings.Config.Plugin_settings,
					}
					sendMessage.SendMessage()
				}
			}

			pluginSettingsForm.SetMinRowsVisible(6)
			pluginSettings.Add(pluginSettingsForm)

			pluginFilesAccordionItems = append(pluginFilesAccordionItems, widget.NewAccordionItem(pluginClassName, pluginSettings))
		}
	}

	pluginAccordion := widget.NewAccordion(pluginFilesAccordionItems...)

	return container.NewVScroll(pluginAccordion)
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

	writeLogFileCheckbox := widget.NewCheck("Write log file", func(writeLogFile bool) {
		fyne.CurrentApp().Preferences().SetBool("WriteLogfile", writeLogFile)
	})
	writeLogFileCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("WriteLogfile", false)

	logTabContent := container.NewBorder(nil, container.NewHBox(RestartBackendButton, writeLogFileCheckbox), nil, nil, Fields.Field.LogText)

	tabs := container.NewAppTabs(
		container.NewTabItem("Log", logTabContent),
		container.NewTabItem("Settings", settingsTabContent),
		container.NewTabItem("Plugins", CreatePluginSettingsPage()),
		container.NewTabItem("About", buildAboutInfo()),
	)
	tabs.SetTabLocation(container.TabLocationTrailing)

	tabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Settings" {
			Settings.BuildSettingsForm(nil, Settings.Config.SettingsFilename)
			tab.Content.(*container.Scroll).Content = Settings.Form
			tab.Content.(*container.Scroll).Refresh()
		}
		if tab.Text == "Plugins" {
			tab.Content.(*container.Scroll).Content = CreatePluginSettingsPage()
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
