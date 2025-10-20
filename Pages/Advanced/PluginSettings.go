package Advanced

import (
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"gopkg.in/yaml.v3"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

type Type int

const (
	TypeAuto Type = iota // Auto-detection type
	TypeFloat
	TypeInt
	TypeString
	TypeNone
)

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func shortenURL(url string, limit int) string {
	if len(url) > limit {
		// Calculate the positions where we will cut the string
		partLen := limit / 2

		firstPart := url[:partLen]
		secondPart := url[len(url)-partLen:]

		return firstPart + "..." + secondPart
	}

	return url
}

func toFloat64(i interface{}) (float64, error) {
	switch v := i.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, errors.New("Unsupported type")
	}
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

func getPluginStatusString(pluginClassName string) string {
	pluginStatusString := " (❌)"
	if Settings.Config.Plugins[pluginClassName] {
		pluginStatusString = " (✅)"
	}
	return pluginStatusString
}

func _getFilePathDialogInitPath(v map[string]interface{}, entry *widget.Entry) (fyne.ListableURI, string) {
	// get file dialog start folder
	appExec, _ := os.Executable()
	currentFolder := filepath.Dir(appExec)
	currentFilename := ""

	if entry.Text != "" {
		entryTextPath := entry.Text
		if v["type"] != "folder_open" && v["type"] != "dir_open" {
			entryTextPath = filepath.Dir(entryTextPath)
		}
		// check if folder exists
		folderExists := false
		if _, err := os.Stat(entryTextPath); !os.IsNotExist(err) {
			folderExists = true
		}
		if folderExists {
			currentFolder = entryTextPath
			currentFilename = filepath.Base(entry.Text)
		}
	}
	fileURI := storage.NewFileURI(currentFolder)
	fileLister, _ := storage.ListerForURI(fileURI)

	return fileLister, currentFilename
}

var onlyShowEnabledPlugins bool
var openPluginItem = -1

func RebuildSinglePluginSettings(pluginClassName string, pluginAccordionItem *widget.AccordionItem, pluginAccordion *widget.Accordion, reloadButtonRef *widget.Button, window *fyne.Window) {
	if pluginAccordionItem != nil {
		pluginSettingsContainer := BuildSinglePluginSettings(pluginClassName, pluginAccordionItem, pluginAccordion, reloadButtonRef, *window)

		pluginAccordionItem.Detail = pluginSettingsContainer
		pluginAccordionItem.Detail.Refresh()
		pluginAccordion.Refresh()
	} else {
		if reloadButtonRef != nil {
			reloadButtonRef.OnTapped()
		}
	}
}

func BuildSinglePluginSettings(pluginClassName string, pluginAccordionItem *widget.AccordionItem, pluginAccordion *widget.Accordion, reloadButtonRef *widget.Button, window fyne.Window) fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Advanced\\PluginSettings->BuildSinglePluginSettings")
	})

	// load settings file for plugin settings
	SettingsFile := Settings.Conf{}
	err := SettingsFile.LoadYamlSettings(filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename))
	if err != nil {
		SettingsFile = Settings.Config
	}

	// plugin to window button
	pluginToWindowButton := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), nil)
	pluginToWindowButton.OnTapped = func() {
		pluginWindow := fyne.CurrentApp().NewWindow(pluginClassName + " " + lang.L("Settings"))
		reloadButton := widget.NewButtonWithIcon(lang.L("Reload"), theme.ViewRefreshIcon(), nil)

		pluginContentWin := BuildSinglePluginSettings(pluginClassName, nil, nil, reloadButton, pluginWindow)
		pluginWindowContainer := container.NewVScroll(pluginContentWin)

		reloadButton.OnTapped = func() {
			pluginContentWin = BuildSinglePluginSettings(pluginClassName, nil, nil, reloadButton, pluginWindow)
			pluginWindowContainer.Content = pluginContentWin
			pluginWindowContainer.Refresh()
			pluginWindow.Content().Refresh()
		}
		reloadButton.Importance = widget.MediumImportance

		pluginWindow.SetContent(container.NewBorder(container.NewBorder(nil, nil, nil, reloadButton, layout.NewSpacer()), nil, nil, nil, pluginWindowContainer))

		// guess the size
		windowHeight := pluginContentWin.Size().Height + reloadButton.Size().Height + 20
		windowWidth := pluginContentWin.Size().Width
		if windowHeight >= fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height {
			windowHeight = fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height
		}
		if pluginAccordionItem != nil {
			windowHeight = pluginAccordionItem.Detail.Size().Height + reloadButton.Size().Height + 20
			windowWidth = pluginAccordionItem.Detail.Size().Width
			if windowHeight >= fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height {
				windowHeight = fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height
			}
			if windowWidth >= 1400 {
				windowWidth = windowWidth / 2
			}
		}
		pluginWindow.Resize(fyne.NewSize(windowWidth, windowHeight))
		pluginWindow.CenterOnScreen()
		pluginWindow.Show()
	}
	if pluginAccordionItem == nil || pluginAccordion == nil {
		pluginToWindowButton.Hide()
	}

	resetSettingsButton := widget.NewButtonWithIcon(lang.L("Reset"), theme.ContentUndoIcon(), nil)
	resetSettingsButton.OnTapped = func() {
		toResetWindow := window
		if toResetWindow == nil {
			toResetWindow, _ = Utilities.GetCurrentMainWindow("Plugin Settings " + pluginClassName)
		}
		translationVarMap := map[string]interface{}{
			"PluginClassName": pluginClassName,
		}
		dialog.ShowConfirm(lang.L("Reset Plugin Settings", translationVarMap), lang.L("Are you sure you want to reset the settings for this plugin?", translationVarMap), func(reset bool) {
			go func() {
				if reset {
					resetSettings(pluginClassName)
					time.Sleep(2 * time.Second)

					RebuildSinglePluginSettings(pluginClassName, pluginAccordionItem, pluginAccordion, reloadButtonRef, &toResetWindow)
				}
			}()
		}, toResetWindow)
	}
	resetSettingsButton.Importance = widget.LowImportance

	reinitSettingsButton := widget.NewButtonWithIcon(lang.L("ReInitialize"), theme.ViewRefreshIcon(), nil)
	reinitSettingsButton.OnTapped = func() {
		toResetWindow := window
		if toResetWindow == nil {
			toResetWindow, _ = Utilities.GetCurrentMainWindow("Plugin Settings " + pluginClassName)
		}
		translationVarMap := map[string]interface{}{
			"PluginClassName": pluginClassName,
		}
		dialog.ShowConfirm(lang.L("Reinitialize Plugin Settings", translationVarMap), lang.L("Are you sure you want to reinitialize this plugin?", translationVarMap), func(reset bool) {
			go func() {
				if reset {
					reinitSettings(pluginClassName)
					time.Sleep(2 * time.Second)

					RebuildSinglePluginSettings(pluginClassName, pluginAccordionItem, pluginAccordion, reloadButtonRef, &toResetWindow)
				}
			}()
		}, toResetWindow)
	}
	reinitSettingsButton.Importance = widget.LowImportance

	// plugin enabled checkbox
	pluginEnabledCheckbox := widget.NewCheck(lang.L("pluginClass enabled", map[string]interface{}{"PluginClass": pluginClassName}), func(enabled bool) {
		Settings.Config.Plugins[pluginClassName] = enabled
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "plugins",
			Value: Settings.Config.Plugins,
		}
		sendMessage.SendMessage()

		if pluginAccordionItem != nil && pluginAccordion != nil {
			pluginAccordionItem.Title = pluginClassName + getPluginStatusString(pluginClassName)
			pluginAccordion.Refresh()
		}
	})
	pluginEnabledCheckbox.Checked = Settings.Config.Plugins[pluginClassName]

	beginLine := canvas.NewHorizontalGradient(&color.NRGBA{R: 198, G: 123, B: 0, A: 255}, &color.NRGBA{R: 198, G: 123, B: 0, A: 0})
	beginLine.Resize(fyne.NewSize(pluginEnabledCheckbox.Size().Width, 2))

	spacerText := canvas.NewText("", &color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	spacerText.TextSize = theme.CaptionTextSize()

	// get plugin settings
	var pluginSettings map[string]interface{}
	if SettingsFile.Plugin_settings != nil {
		if settings, ok := SettingsFile.Plugin_settings.(map[string]interface{})[pluginClassName]; ok {
			if settingsMap, ok := settings.(map[string]interface{}); ok {
				pluginSettings = settingsMap
			}
		}
	}

	// check if settings_groups exists
	if groupData, exists := pluginSettings["settings_groups"]; exists && groupData != nil {
		// create settings fields grouped by 'settings_groups'
		var settingsGroups map[string]interface{}
		settingsGroupsByte, _ := json.Marshal(groupData)
		json.Unmarshal(settingsGroupsByte, &settingsGroups)

		// Convert the map to a slice and sort
		type kv struct {
			Key   string
			Value interface{}
		}

		var settingsGroupList []kv
		for k, v := range settingsGroups {
			settingsGroupList = append(settingsGroupList, kv{k, v})
		}

		sort.Slice(settingsGroupList, func(i, j int) bool {
			if strings.EqualFold(settingsGroupList[i].Key, "General") {
				return true
			} else if strings.EqualFold(settingsGroupList[j].Key, "General") {
				return false
			}
			return strings.ToLower(settingsGroupList[i].Key) < strings.ToLower(settingsGroupList[j].Key)
		})

		settingsGroupTabs := container.NewAppTabs()
		for _, kv := range settingsGroupList {
			groupName := kv.Key
			groupContainer := container.NewVBox()

			switch group := kv.Value.(type) {
			case []interface{}:
				// Check if it's a slice of strings (single column)
				if len(group) > 0 {
					if _, ok := group[0].(string); ok {
						var settingsGroup []string
						for _, item := range group {
							settingsGroup = append(settingsGroup, item.(string))
						}
						sort.Strings(settingsGroup)
						for _, settingName := range settingsGroup {
							if _, ok := pluginSettings[settingName]; ok && settingName != "settings_groups" {
								settingsFields := createSettingsFields(pluginSettings, settingName, &SettingsFile, pluginClassName, window)
								for _, field := range settingsFields {
									groupContainer.Add(field)
								}
							}
						}
					} else if _, ok := group[0].([]interface{}); ok {
						// Handle multiple columns
						columnContainers := []fyne.CanvasObject{}
						for _, column := range group {
							columnFields := []fyne.CanvasObject{}
							for _, item := range column.([]interface{}) {
								settingName := item.(string)
								if _, ok := pluginSettings[settingName]; ok && settingName != "settings_groups" {
									settingsFields := createSettingsFields(pluginSettings, settingName, &SettingsFile, pluginClassName, window)
									columnFields = append(columnFields, settingsFields...)
								}
							}
							// Create a VBox for each column
							columnContainer := container.NewVBox(columnFields...)
							columnContainers = append(columnContainers, columnContainer)
						}
						// Add all column VBoxes to a Grid columns container
						groupContainer.Add(container.NewGridWithColumns(len(columnContainers), columnContainers...))
					}
				}
			}

			settingsGroupTabs.Append(container.NewTabItem(groupName, groupContainer))
		}

		pluginSettingsContainer := container.NewVBox(
			container.NewBorder(nil, nil, nil, container.NewHBox(reinitSettingsButton, resetSettingsButton, pluginToWindowButton), pluginEnabledCheckbox),
			container.NewGridWithColumns(2, beginLine),
			settingsGroupTabs,
			spacerText,
		)

		return pluginSettingsContainer
	} else {
		// no grouping
		pluginSettingsContainer := container.NewVBox()
		pluginSettingsContainer.Add(container.NewBorder(nil, nil, nil, container.NewHBox(reinitSettingsButton, resetSettingsButton, pluginToWindowButton), pluginEnabledCheckbox))
		pluginSettingsContainer.Add(container.NewGridWithColumns(2, beginLine))

		var sortedSettingNames []string
		for settingName := range pluginSettings {
			sortedSettingNames = append(sortedSettingNames, settingName)
		}
		sort.Strings(sortedSettingNames) // sort the keys in ascending order

		for _, settingName := range sortedSettingNames {
			if settingName != "settings_groups" {
				settingsFields := createSettingsFields(pluginSettings, settingName, &SettingsFile, pluginClassName, window)
				for _, field := range settingsFields {
					pluginSettingsContainer.Add(field)
				}
			}
		}

		pluginSettingsContainer.Add(spacerText)

		return pluginSettingsContainer
	}
}

func getPluginFileAndClassName(pluginClassName string) (string, string) {
	files, err := os.ReadDir(filepath.Join(".", "Plugins"))
	if err != nil {
		println(err)
	}
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "__init__") && (strings.HasSuffix(file.Name(), ".py")) {
			pluginClass := GetClassNameOfPlugin(filepath.Join(".", "Plugins", file.Name()))
			if pluginClass == pluginClassName {
				return file.Name(), pluginClass
			}
		}
	}
	return "", ""
}

func BuildPluginSettingsAccordion(window fyne.Window) (fyne.CanvasObject, int) {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Advanced\\PluginSettings->BuildPluginSettingsAccordion")
	})

	// build plugins list
	var pluginFiles []string
	files, err := os.ReadDir(filepath.Join(".", "Plugins"))
	if err != nil {
		println(err)
	}
	pluginAccordion := widget.NewAccordion()

	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "__init__") && (strings.HasSuffix(file.Name(), ".py")) {
			pluginFiles = append(pluginFiles, file.Name())
			pluginClassName := GetClassNameOfPlugin(filepath.Join(".", "Plugins", file.Name()))

			// only display enabled plugins if onlyShowEnabledPlugins is true
			if onlyShowEnabledPlugins && !Settings.Config.Plugins[pluginClassName] {
				continue
			}

			pluginAccordionItem := widget.NewAccordionItem(
				pluginClassName+getPluginStatusString(pluginClassName),
				nil,
			)

			pluginSettingsContainer := BuildSinglePluginSettings(pluginClassName, pluginAccordionItem, pluginAccordion, nil, window)

			pluginAccordionItem.Detail = pluginSettingsContainer

			pluginAccordion.Append(pluginAccordionItem)
		}
	}

	//if openPluginItem >= 0 {
	//	pluginAccordion.Open(openPluginItem)
	//}
	return pluginAccordion, len(pluginFiles)
}

func CreatePluginSettingsPage() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Advanced\\PluginSettings->CreatePluginSettingsPage")
	})

	pluginAccordion, pluginFilesCount := BuildPluginSettingsAccordion(nil)

	pluginsContent := container.NewVScroll(nil)

	// plugin list to window button
	pluginToWindowButton := widget.NewButtonWithIcon(lang.L("Open List in Window"), theme.ViewFullScreenIcon(), nil)

	pluginToWindowButton.OnTapped = func() {
		pluginWindow := fyne.CurrentApp().NewWindow(lang.L("Plugin Settings"))
		pluginAccordionWin, _ := BuildPluginSettingsAccordion(pluginWindow)
		pluginWindowContainer := container.NewVScroll(pluginAccordionWin)
		reloadButton := widget.NewButtonWithIcon(lang.L("Reload"), theme.ViewRefreshIcon(), nil)
		reloadButton.OnTapped = func() {
			openItem := -1
			for index, item := range pluginAccordionWin.(*widget.Accordion).Items {
				if item.Open {
					openItem = index
					break
				}
			}
			pluginAccordionWin, _ = BuildPluginSettingsAccordion(pluginWindow)
			pluginWindowContainer.Content = pluginAccordionWin
			pluginWindowContainer.Refresh()
			pluginWindow.Content().Refresh()
			if openItem >= 0 && openItem <= len(pluginAccordionWin.(*widget.Accordion).Items) {
				pluginAccordionWin.(*widget.Accordion).Open(openItem)
			}
		}
		reloadButton.Importance = widget.MediumImportance
		pluginWindow.SetContent(container.NewBorder(container.NewBorder(nil, nil, nil, reloadButton), nil, nil, nil, pluginWindowContainer))

		windowHeight := pluginAccordion.Size().Height
		if windowHeight >= fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height {
			windowHeight = fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size().Height
		}
		pluginWindow.Resize(fyne.NewSize(pluginAccordion.Size().Width, windowHeight))
		pluginWindow.CenterOnScreen()

		pluginWindow.Show()
		// loop while pluginWindow is open, do refresh
	}

	downloadButton := widget.NewButton(lang.L("Download / Update Plugins"), nil)
	filterEnabledPluginsCheckbox := widget.NewCheck(lang.L("Only show enabled Plugins"), nil)
	onlyShowEnabledPlugins = fyne.CurrentApp().Preferences().BoolWithFallback("OnlyShowEnabledPlugins", onlyShowEnabledPlugins)
	filterEnabledPluginsCheckbox.Checked = onlyShowEnabledPlugins
	//topContainer := container.NewBorder(nil, nil, nil, filterEnabledPluginsCheckbox, downloadButton)
	topContainer := container.NewBorder(nil, nil, nil, container.NewBorder(nil, nil, filterEnabledPluginsCheckbox, pluginToWindowButton), container.NewBorder(nil, nil, nil, downloadButton))
	downloadButton.OnTapped = func() {
		CreatePluginListWindow(func() {
			pluginAccordion, pluginFilesCount = BuildPluginSettingsAccordion(nil)
			if pluginFilesCount > 0 {
				pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
				pluginsContent.Content = pluginsContainerBorder

				pluginAccordion.Refresh()
				pluginsContent.Refresh()
			}
		}, true)
	}
	downloadButton.Importance = widget.HighImportance
	downloadButton.Refresh()

	filterEnabledPluginsCheckbox.OnChanged = func(enabled bool) {
		onlyShowEnabledPlugins = enabled
		fyne.CurrentApp().Preferences().SetBool("OnlyShowEnabledPlugins", onlyShowEnabledPlugins)

		pluginAccordion, pluginFilesCount = BuildPluginSettingsAccordion(nil)
		if pluginFilesCount > 0 {
			pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
			pluginsContent.Content = pluginsContainerBorder

			pluginAccordion.Refresh()
			pluginsContent.Refresh()
		}
	}

	if pluginFilesCount == 0 {
		openPluginsFolderButton := widget.NewButton(lang.L("Open Plugins folder"), func() {
			appExec, _ := os.Executable()
			appPath := filepath.Dir(appExec)
			uiPluginsFolder, _ := url.Parse(filepath.Join(appPath, "Plugins"))
			err := fyne.CurrentApp().OpenURL(uiPluginsFolder)
			if err != nil {
				println(err)
				Logging.CaptureException(err)
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			}
		})

		pluginsContent.Content = container.NewCenter(
			container.NewVBox(
				widget.NewLabel("\n"+lang.L("No Plugins found. Download Plugins using the button below.")),
				downloadButton,
				widget.NewLabel("\n"+lang.L("Or download Plugins manually from:")),
				widget.NewHyperlink("https://github.com/Sharrnah/whispering-plugins/blob/main/README.md", parseURL("https://github.com/Sharrnah/whispering-plugins/blob/main/README.md")),
				widget.NewLabel(lang.L("and place the *.py file in the Plugins folder.")),
				openPluginsFolderButton,
			),
		)
	} else {
		pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
		pluginsContent.Content = pluginsContainerBorder
	}

	return pluginsContent
}

func createSettingsFields(pluginSettings map[string]interface{}, settingName string, SettingsFile *Settings.Conf, pluginClassName string, window fyne.Window) []fyne.CanvasObject {
	var settingsFields []fyne.CanvasObject

	if window == nil {
		if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
			window = fyne.CurrentApp().Driver().AllWindows()[0]
		} else {
			window = fyne.CurrentApp().NewWindow(lang.L("Plugin Settings") + " " + pluginClassName)
			window.Show()
		}
	}

	// Skip creating a field for "settings_groups"
	if settingName == "settings_groups" {
		return settingsFields
	}

	// widget Entry OnChange function
	_entryOnChange := func(text string) {
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			pluginSettings[settingName] = val
		} else if val, err := strconv.ParseInt(text, 10, 64); err == nil {
			pluginSettings[settingName] = val
		} else if text == "None" {
			pluginSettings[settingName] = nil
		} else {
			pluginSettings[settingName] = text
		}
		updateSettings(*SettingsFile, pluginClassName, pluginSettings)
	}
	_entryOnChangeWithForceType := func(text string, forceType Type) {
		switch forceType {
		case TypeFloat:
			if val, err := strconv.ParseFloat(text, 64); err == nil {
				pluginSettings[settingName] = val
			}
		case TypeInt:
			if val, err := strconv.ParseInt(text, 10, 64); err == nil {
				pluginSettings[settingName] = val
			}
		case TypeString:
			pluginSettings[settingName] = text
		case TypeNone:
			pluginSettings[settingName] = nil
		case TypeAuto:
			// Fall back to the default auto detection
			_entryOnChange(text)
		default:
			// Handle unexpected type by using default behavior
			_entryOnChange(text)
		}
		updateSettings(*SettingsFile, pluginClassName, pluginSettings)
	}
	makeEntryOnChange := func(forceType Type) func(text string) {
		return func(text string) {
			_entryOnChangeWithForceType(text, forceType)
		}
	}

	switch v := pluginSettings[settingName].(type) {
	case bool:
		check := widget.NewCheck(settingName, func(value bool) {
			pluginSettings[settingName] = value
			updateSettings(*SettingsFile, pluginClassName, pluginSettings)
		})
		check.Checked = v
		settingsFields = append(settingsFields, check)
	case int, float64:
		entry := widget.NewEntry()
		entry.SetText(fmt.Sprintf("%v", v))
		entry.OnChanged = makeEntryOnChange(TypeAuto)
		settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
	case string:
		if strings.Contains(v, "\n") {
			entry := widget.NewMultiLineEntry()
			entry.SetText(v)
			entry.OnChanged = func(text string) {
				pluginSettings[settingName] = text
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}
			// count number of lines in pluginSettingsForm and set minRowsVisible
			lines := strings.Count(entry.Text, "\n")
			if lines < 5 {
				lines = 5
			}
			entry.SetMinRowsVisible(lines + 1)
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
		} else {
			entry := widget.NewEntry()
			entry.SetText(v)
			entry.OnChanged = makeEntryOnChange(TypeAuto)
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
		}
	case nil:
		entry := widget.NewEntry()
		entry.SetText("None")
		entry.OnChanged = makeEntryOnChange(TypeAuto)
		settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
	case map[string]interface{}:
		// if 'type' field is set to 'button', create a button
		if v["type"] == "button" {
			label := v["label"].(string)
			button := widget.NewButton(label, func() {
				sendMessage := SendMessageChannel.SendMessageStruct{
					Type:  "plugin_button_press",
					Name:  pluginClassName,
					Value: settingName,
				}
				sendMessage.SendMessage()
			})

			if buttonStyle, ok := v["style"].(string); ok && buttonStyle == "primary" {
				button.Importance = widget.HighImportance
			} else {
				button.Importance = widget.MediumImportance
			}

			settingsFields = append(settingsFields, container.NewHBox(button))
		} else if v["type"] == "slider" {
			min, _ := toFloat64(v["min"])
			max, _ := toFloat64(v["max"])
			step, _ := toFloat64(v["step"])
			value, _ := toFloat64(v["value"])
			slider := widget.NewSlider(min, max)
			slider.Step = step
			slider.SetValue(value)
			// get precision from step size
			precision := 0
			if step < 1 {
				precision = len(strings.Split(strings.TrimRight(fmt.Sprintf("%f", step), "0"), ".")[1])
			}
			// set precision for sprintf
			precisionString := fmt.Sprintf("%%.%df", precision)

			sliderState := widget.NewLabel(fmt.Sprintf(precisionString, value))
			slider.OnChanged = func(value float64) {
				sliderState.SetText(fmt.Sprintf(precisionString, value))
				v["value"] = value
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), sliderState, slider))
		} else if v["type"] == "select" {
			var optionsString []string
			if v["values"] != nil {
				for _, val := range v["values"].([]interface{}) {
					optionsString = append(optionsString, fmt.Sprintf("%v", val))
				}
			}
			selectEntry := widget.NewSelect(optionsString, func(value string) {
				v["value"] = value
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			})
			selectEntry.Selected = v["value"].(string)
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, selectEntry))
		} else if v["type"] == "select_textvalue" {
			var options []CustomWidget.TextValueOption
			if v["values"] != nil {
				for _, val := range v["values"].([]interface{}) {
					if pair, ok := val.([]interface{}); ok && len(pair) == 2 {
						text := fmt.Sprintf("%v", pair[0])
						value := fmt.Sprintf("%v", pair[1])
						options = append(options, CustomWidget.TextValueOption{Text: text, Value: value})
					}
				}
			}
			selectEntry := CustomWidget.NewTextValueSelect(settingName, options, func(value CustomWidget.TextValueOption) {
				v["value"] = value.Text
				v["_value_real"] = value.Value
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}, 0)
			selectEntry.Selected = v["value"].(string)
			selectEntry.Refresh()
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, selectEntry))
		} else if v["type"] == "select_completion" {
			var options []CustomWidget.TextValueOption
			if v["values"] != nil {
				for _, val := range v["values"].([]interface{}) {
					if pair, ok := val.([]interface{}); ok && len(pair) == 2 {
						text := fmt.Sprintf("%v", pair[0])
						value := fmt.Sprintf("%v", pair[1])
						options = append(options, CustomWidget.TextValueOption{Text: text, Value: value})
					}
				}
			}
			selectCompletionEntry := CustomWidget.NewCompletionEntry([]string{})
			selectCompletionEntry.SetValueOptions(options)
			selectCompletionEntry.Text = v["value"].(string)

			selectCompletionEntry.ShowAllEntryText = lang.L("... show all")
			selectCompletionEntry.Entry.PlaceHolder = lang.L("Select a language")
			selectCompletionEntry.OnChanged = func(value string) {
				// filter out the values of Options that do not contain the value
				var filteredValues []string
				for i := 0; i < len(selectCompletionEntry.Options); i++ {
					if len(selectCompletionEntry.Options) > i && strings.Contains(strings.ToLower(selectCompletionEntry.Options[i]), strings.ToLower(value)) {
						filteredValues = append(filteredValues, selectCompletionEntry.Options[i])
					}
				}

				selectCompletionEntry.SetOptionsFilter(filteredValues)
				selectCompletionEntry.ShowCompletion()
			}
			selectCompletionEntry.OnSubmitted = func(value string) {
				// check if value is not in Options
				value = selectCompletionEntry.UpdateCompletionEntryBasedOnValue(value)

				v["value"] = value
				v["_value_real"] = selectCompletionEntry.GetValueOptionEntryByText(value).Value
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}

			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, selectCompletionEntry))
		} else if v["type"] == "textarea" {
			entry := widget.NewMultiLineEntry()
			entry.SetText(v["value"].(string))
			entry.OnChanged = func(text string) {
				v["value"] = text
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}
			if rows, ok := v["rows"].(int); ok {
				entry.SetMinRowsVisible(rows)
			} else if rows, ok := v["rows"].(float64); ok {
				entry.SetMinRowsVisible(int(rows))
			} else {
				entry.SetMinRowsVisible(5)
			}
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
		} else if v["type"] == "textfield" {
			entry := widget.NewEntry()
			if password, ok := v["password"].(bool); ok {
				if password {
					entry = widget.NewPasswordEntry()
				}
			}
			entry.SetText(v["value"].(string))
			entry.OnChanged = func(text string) {
				v["value"] = text
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
		} else if v["type"] == "hyperlink" {
			link := v["value"].(string)
			hyperlink := widget.NewHyperlink(v["label"].(string)+" ("+shortenURL(link, 50)+")", parseURL(link))
			settingsFields = append(settingsFields, container.NewHBox(hyperlink))
		} else if v["type"] == "label" {
			label := widget.NewLabel(v["label"].(string))
			// set style only if key exists
			if _, ok := v["style"]; ok {
				style := v["style"].(string)
				if style == "left" {
					label.Alignment = fyne.TextAlignLeading
				}
				if style == "right" {
					label.Alignment = fyne.TextAlignTrailing
				}
				if style == "center" {
					label.Alignment = fyne.TextAlignCenter
				}
			}
			label.Wrapping = fyne.TextWrapOff
			settingsFields = append(settingsFields, container.NewHScroll(label))
		} else if v["type"] == "file_open" || v["type"] == "file_save" || v["type"] == "folder_open" || v["type"] == "dir_open" {
			selectButtonLabel := lang.L("Select File")
			entry := widget.NewEntry()
			entry.SetText(v["value"].(string))

			// get file dialog start folder
			fileLister, currentFilename := _getFilePathDialogInitPath(v, entry)

			entry.OnChanged = func(text string) {
				v["value"] = text
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)

				// update dialog initpath on change
				fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
			}
			filter := v["accept"].(string)
			filterArray := strings.Split(filter, ",")

			fileSelectFunc := func() {}

			if v["type"] == "file_open" {
				fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err == nil && reader != nil {
						entry.SetText(reader.URI().Path())
					}
				}, window)
				fileDialog.SetFilter(storage.NewExtensionFileFilter(filterArray))
				fileSelectFunc = func() {
					// resize dialog
					dialogSize := window.Canvas().Size()
					dialogSize.Height = dialogSize.Height - 50
					dialogSize.Width = dialogSize.Width - 50
					fileDialog.Resize(dialogSize)

					// update dialog initpath on change
					fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
					fileDialog.SetLocation(fileLister)
					fileDialog.SetFileName(currentFilename)

					fileDialog.Show()
				}
			} else if v["type"] == "file_save" {
				fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err == nil && writer != nil {
						entry.SetText(writer.URI().Path())
					}
				}, window)
				fileDialog.SetFilter(storage.NewExtensionFileFilter(filterArray))

				fileSelectFunc = func() {
					// resize dialog
					dialogSize := window.Canvas().Size()
					dialogSize.Height = dialogSize.Height - 50
					dialogSize.Width = dialogSize.Width - 50
					fileDialog.Resize(dialogSize)

					// update dialog initpath on change
					fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
					fileDialog.SetLocation(fileLister)
					fileDialog.SetFileName(currentFilename)

					fileDialog.Show()
				}
			} else if v["type"] == "folder_open" || v["type"] == "dir_open" {
				selectButtonLabel = lang.L("Select Folder")
				fileDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
					if err == nil && uri != nil {
						entry.SetText(uri.Path())
					}
				}, window)
				fileSelectFunc = func() {
					// resize dialog
					dialogSize := window.Canvas().Size()
					dialogSize.Height = dialogSize.Height - 50
					dialogSize.Width = dialogSize.Width - 50
					fileDialog.Resize(dialogSize)

					// update dialog initpath on change
					fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
					fileDialog.SetLocation(fileLister)

					fileDialog.Show()
				}
			}

			fileSelectButton := widget.NewButton(selectButtonLabel, fileSelectFunc)

			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), fileSelectButton, entry))
		} else if v["type"] == "select_audio" {
			valueTextCleanupFunc := func(value string) string {
				// delete everything after the last "- API:" and trim spaces afterward.
				lastIndex := strings.LastIndex(value, "- API:")
				if lastIndex != -1 {
					// Keep everything before the last occurrence
					value = value[:lastIndex]
				}
				// Trim spaces afterward
				value = strings.TrimSpace(value)
				return value
			}

			var selectEntries []CustomWidget.TextValueOption = nil

			var audioInputDevices = Utilities.AudioDeviceMemory{}
			var audioOutputDevices = Utilities.AudioDeviceMemory{}

			audioInputDevices = Utilities.AudioInputDeviceList[strings.ToLower(Settings.Config.Audio_api)]
			audioOutputDevices = Utilities.AudioOutputDeviceList[strings.ToLower(Settings.Config.Audio_api)]

			selectEntries = audioInputDevices.WidgetOptions
			selectEntries = append(selectEntries, audioOutputDevices.WidgetOptions...)

			selectedDeviceApi := strings.ToLower(Settings.Config.Audio_api)

			if deviceApi, ok := v["device_api"].(string); ok {
				selectedDeviceApi = strings.ToLower(deviceApi)
				selectEntries = Utilities.AudioInputDeviceList[selectedDeviceApi].WidgetOptions
				selectEntries = append(selectEntries, Utilities.AudioOutputDeviceList[selectedDeviceApi].WidgetOptions...)

				if strings.ToLower(deviceApi) == "all" {
					selectEntries = []CustomWidget.TextValueOption{}

					for index := range Utilities.AudioInputDeviceList {
						selectEntries = append(selectEntries,
							CustomWidget.TextValueOption{Text: "==========[" + strings.ToUpper(index) + "]==========", Value: ""},
						)
						selectEntries = append(selectEntries, Utilities.AudioInputDeviceList[index].WidgetOptions...)
						selectEntries = append(selectEntries, Utilities.AudioOutputDeviceList[index].WidgetOptions...)
					}
				}
			}

			if deviceType, ok := v["device_type"].(string); ok {

				if strings.ToLower(deviceType) == "input" {
					selectEntries = audioInputDevices.WidgetOptions
					if selectedDeviceApi != "all" && selectedDeviceApi != "" && len(Utilities.AudioInputDeviceList[selectedDeviceApi].WidgetOptions) > 0 {
						selectEntries = Utilities.AudioInputDeviceList[selectedDeviceApi].WidgetOptions
					} else if selectedDeviceApi == "all" {
						selectEntries = []CustomWidget.TextValueOption{}

						for index := range Utilities.AudioInputDeviceList {
							selectEntries = append(selectEntries,
								CustomWidget.TextValueOption{Text: "==========[" + strings.ToUpper(index) + "]==========", Value: ""},
							)
							selectEntries = append(selectEntries, Utilities.AudioInputDeviceList[index].WidgetOptions...)
						}
					}
				} else if strings.ToLower(deviceType) == "output" {
					selectEntries = audioOutputDevices.WidgetOptions
					if selectedDeviceApi != "all" && selectedDeviceApi != "" && len(Utilities.AudioOutputDeviceList[selectedDeviceApi].WidgetOptions) > 0 {
						selectEntries = Utilities.AudioOutputDeviceList[selectedDeviceApi].WidgetOptions
					} else if selectedDeviceApi == "all" {
						selectEntries = []CustomWidget.TextValueOption{}

						for index := range Utilities.AudioOutputDeviceList {
							selectEntries = append(selectEntries,
								CustomWidget.TextValueOption{Text: "==========[" + strings.ToUpper(index) + "]==========", Value: ""},
							)
							selectEntries = append(selectEntries, Utilities.AudioOutputDeviceList[index].WidgetOptions...)
						}
					}
				}
			}

			audioSelect := CustomWidget.NewTextValueSelect("device_index", selectEntries, nil, 0)

			savedValue := ""
			switch val := v["value"].(type) {
			case int:
				// convert val to string
				savedValue = strconv.Itoa(val)
			case string:
				savedValue = val
			}

			savedValueSplit := strings.Split(savedValue, "#|")
			savedValueAudioApi := ""
			savedValueAudioType := ""
			if len(savedValueSplit) >= 2 {
				savedValueAudioApi = strings.Split(savedValueSplit[1], ",")[0]
				savedValueAudioType = strings.Split(savedValueSplit[1], ",")[1]
			}

			var updateFunc = func(s CustomWidget.TextValueOption) {
				//if s.Value != savedValue {
				v["value"] = s.Value
				v["_value_text"] = valueTextCleanupFunc(s.Text)
				pluginSettings[settingName] = v
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
				//}
			}

			if selectedValueText, ok := v["_value_text"].(string); ok && valueTextCleanupFunc(selectedValueText) != "" {
				containedEntry := audioSelect.GetEntry(&CustomWidget.TextValueOption{
					Text: selectedValueText,
				}, CustomWidget.CompareText)
				if containedEntry != nil {
					if selectedDeviceApi != "all" && selectedDeviceApi != "" {
						audioSelect.SetSelectedByText(selectedValueText)
					} else {
						// loop over all audioSelect.Options and check each value if the AudioApi and AudioType together with the selectedValueText is the same
						for _, option := range audioSelect.Options {
							optionSplit := strings.Split(option.Value, "#|")
							optionAudioApi := ""
							optionAudioType := ""
							if len(optionSplit) >= 2 {
								optionAudioApi = strings.Split(optionSplit[1], ",")[0]
								optionAudioType = strings.Split(optionSplit[1], ",")[1]
							}

							if optionAudioApi == savedValueAudioApi && optionAudioType == savedValueAudioType && valueTextCleanupFunc(option.Text) == valueTextCleanupFunc(selectedValueText) {
								audioSelect.SetSelected(option.Value)
								break
							}
						}
					}
				} else {
					audioSelect.SetSelected(savedValue)
				}
			} else {
				audioSelect.SetSelected(savedValue)
			}

			audioSelect.OnChanged = updateFunc

			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, audioSelect))
		} else {
			yamlBytes, _ := yaml.Marshal(v)
			entry := widget.NewMultiLineEntry()
			entry.SetText(string(yamlBytes))
			entry.OnChanged = func(text string) {
				var newValue map[string]interface{}
				if err := yaml.Unmarshal([]byte(text), &newValue); err == nil {
					pluginSettings[settingName] = newValue
					updateSettings(*SettingsFile, pluginClassName, pluginSettings)
				}
			}
			// count number of lines in pluginSettingsForm and set minRowsVisible
			lines := strings.Count(entry.Text, "\n")
			if lines < 5 {
				lines = 5
			}
			entry.SetMinRowsVisible(lines + 1)
			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
		}
	case []interface{}:
		yamlBytes, _ := yaml.Marshal(v)
		entry := widget.NewMultiLineEntry()
		entry.SetText(string(yamlBytes))
		entry.OnChanged = func(text string) {
			var newValue []interface{}
			if err := yaml.Unmarshal([]byte(text), &newValue); err == nil {
				pluginSettings[settingName] = newValue
				updateSettings(*SettingsFile, pluginClassName, pluginSettings)
			}
		}
		// count number of lines in pluginSettingsForm and set minRowsVisible
		lines := strings.Count(entry.Text, "\n")
		if lines < 5 {
			lines = 5
		}
		entry.SetMinRowsVisible(lines + 1)
		settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), nil, entry))
	}

	return settingsFields
}

// Helper function to update settings
func updateSettings(SettingsFile Settings.Conf, pluginClassName string, pluginSettings map[string]interface{}) {
	SettingsFile.Plugin_settings.(map[string]interface{})[pluginClassName] = pluginSettings
	sendMessage := SendMessageChannel.SendMessageStruct{
		Type:  "setting_change",
		Name:  "plugin_settings",
		Value: SettingsFile.Plugin_settings,
	}
	sendMessage.SendMessage()
}

// Helper function to update settings
func resetSettings(pluginClassName string) {
	sendMessage := SendMessageChannel.SendMessageStruct{
		Type:  "setting_reset_all",
		Name:  "plugin",
		Value: pluginClassName,
	}
	sendMessage.SendMessage()
}

func reinitSettings(pluginClassName string) {
	sendMessage := SendMessageChannel.SendMessageStruct{
		Type:  "setting_reinit",
		Name:  "plugin",
		Value: pluginClassName,
	}
	sendMessage.SendMessage()
}
