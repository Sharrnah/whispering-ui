package Advanced

import (
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"whispering-tiger-ui/Fields"
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

func BuildPluginSettingsAccordion() (fyne.CanvasObject, int) {
	// load settings file for plugin settings
	SettingsFile := Settings.Conf{}
	err := SettingsFile.LoadYamlSettings(filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename))
	if err != nil {
		SettingsFile = Settings.Config
	}

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

			// plugin enabled checkbox
			pluginEnabledCheckbox := widget.NewCheck(pluginClassName+" enabled", func(enabled bool) {
				Settings.Config.Plugins[pluginClassName] = enabled
				sendMessage := Fields.SendMessageStruct{
					Type:  "setting_change",
					Name:  "plugins",
					Value: Settings.Config.Plugins,
				}
				sendMessage.SendMessage()

				pluginAccordionItem.Title = pluginClassName + getPluginStatusString(pluginClassName)
				pluginAccordion.Refresh()
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
				var settingsGroups map[string][]string
				settingsGroupsByte, _ := json.Marshal(groupData)
				json.Unmarshal(settingsGroupsByte, &settingsGroups)

				// Convert the map to a slice and sort
				type kv struct {
					Key   string
					Value []string
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
					settingsGroup := kv.Value

					groupContainer := container.NewVBox()

					sort.Strings(settingsGroup)

					for _, settingName := range settingsGroup {
						if _, ok := pluginSettings[settingName]; ok && settingName != "settings_groups" {
							settingsFields := createSettingsFields(pluginSettings, settingName, &SettingsFile, pluginClassName)
							for _, field := range settingsFields {
								groupContainer.Add(field)
							}
						}
					}
					settingsGroupTabs.Append(container.NewTabItem(groupName, groupContainer))
				}

				pluginSettingsContainer := container.NewVBox(
					pluginEnabledCheckbox,
					container.NewGridWithColumns(2, beginLine),
					settingsGroupTabs,
					spacerText,
				)
				pluginAccordionItem.Detail = pluginSettingsContainer
			} else {
				// no grouping
				pluginSettingsContainer := container.NewVBox()
				pluginSettingsContainer.Add(pluginEnabledCheckbox)
				pluginSettingsContainer.Add(container.NewGridWithColumns(2, beginLine))

				var sortedSettingNames []string
				for settingName := range pluginSettings {
					sortedSettingNames = append(sortedSettingNames, settingName)
				}
				sort.Strings(sortedSettingNames) // sort the keys in ascending order

				for _, settingName := range sortedSettingNames {
					if settingName != "settings_groups" {
						settingsFields := createSettingsFields(pluginSettings, settingName, &SettingsFile, pluginClassName)
						for _, field := range settingsFields {
							pluginSettingsContainer.Add(field)
						}
					}
				}

				pluginSettingsContainer.Add(spacerText)

				pluginAccordionItem.Detail = pluginSettingsContainer
			}
			pluginAccordion.Append(pluginAccordionItem)
		}
	}

	//if openPluginItem >= 0 {
	//	pluginAccordion.Open(openPluginItem)
	//}

	return pluginAccordion, len(pluginFiles)
}

func CreatePluginSettingsPage() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	pluginAccordion, pluginFilesCount := BuildPluginSettingsAccordion()

	pluginsContent := container.NewVScroll(nil)

	downloadButton := widget.NewButton("Download / Update Plugins", nil)
	filterEnabledPluginsCheckbox := widget.NewCheck("Only show enabled plugins", nil)
	filterEnabledPluginsCheckbox.Checked = onlyShowEnabledPlugins
	topContainer := container.NewBorder(nil, nil, nil, filterEnabledPluginsCheckbox, downloadButton)
	downloadButton.OnTapped = func() {
		CreatePluginListWindow(func() {
			pluginAccordion, pluginFilesCount = BuildPluginSettingsAccordion()
			if pluginFilesCount > 0 {
				pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
				pluginsContent.Content = pluginsContainerBorder

				pluginAccordion.Refresh()
				pluginsContent.Refresh()
			}
		})
	}
	downloadButton.Importance = widget.HighImportance
	downloadButton.Refresh()

	filterEnabledPluginsCheckbox.OnChanged = func(enabled bool) {
		onlyShowEnabledPlugins = enabled

		pluginAccordion, pluginFilesCount = BuildPluginSettingsAccordion()
		if pluginFilesCount > 0 {
			pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
			pluginsContent.Content = pluginsContainerBorder

			pluginAccordion.Refresh()
			pluginsContent.Refresh()
		}
	}

	if pluginFilesCount == 0 {
		openPluginsFolderButton := widget.NewButton("Open Plugins folder", func() {
			appExec, _ := os.Executable()
			appPath := filepath.Dir(appExec)
			uiPluginsFolder, _ := url.Parse(filepath.Join(appPath, "Plugins"))
			err := fyne.CurrentApp().OpenURL(uiPluginsFolder)
			if err != nil {
				println(err)
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			}
		})

		pluginsContent.Content = container.NewCenter(
			container.NewVBox(
				widget.NewLabel("\nNo Plugins found.\n\nDownload Plugins using the button below."),
				downloadButton,
				widget.NewLabel("\nOr download Plugins manually from:"),
				widget.NewHyperlink("https://github.com/Sharrnah/whispering/blob/main/documentation/plugins.md", parseURL("https://github.com/Sharrnah/whispering/blob/main/documentation/plugins.md")),
				widget.NewLabel("and place the *.py file in the Plugins folder."),
				openPluginsFolderButton,
			),
		)
	} else {
		pluginsContainerBorder := container.NewBorder(topContainer, nil, nil, nil, pluginAccordion)
		pluginsContent.Content = pluginsContainerBorder
	}

	return pluginsContent
}

func createSettingsFields(pluginSettings map[string]interface{}, settingName string, SettingsFile *Settings.Conf, pluginClassName string) []fyne.CanvasObject {
	var settingsFields []fyne.CanvasObject

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
				sendMessage := Fields.SendMessageStruct{
					Type:  "plugin_button_press",
					Name:  pluginClassName,
					Value: settingName,
				}
				sendMessage.SendMessage()
			})
			if v["style"] == "primary" {
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
			selectButtonLabel := "Select File"
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
				}, fyne.CurrentApp().Driver().AllWindows()[0])
				fileDialog.SetFilter(storage.NewExtensionFileFilter(filterArray))
				fileSelectFunc = func() {
					// resize dialog
					if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
						dialogSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
						dialogSize.Height = dialogSize.Height - 50
						dialogSize.Width = dialogSize.Width - 50
						fileDialog.Resize(dialogSize)
					}

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
				}, fyne.CurrentApp().Driver().AllWindows()[0])
				fileDialog.SetFilter(storage.NewExtensionFileFilter(filterArray))

				fileSelectFunc = func() {
					// resize dialog
					if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
						dialogSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
						dialogSize.Height = dialogSize.Height - 50
						dialogSize.Width = dialogSize.Width - 50
						fileDialog.Resize(dialogSize)
					}

					// update dialog initpath on change
					fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
					fileDialog.SetLocation(fileLister)
					fileDialog.SetFileName(currentFilename)

					fileDialog.Show()
				}
			} else if v["type"] == "folder_open" || v["type"] == "dir_open" {
				selectButtonLabel = "Select Folder"
				fileDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
					if err == nil && uri != nil {
						entry.SetText(uri.Path())
					}
				}, fyne.CurrentApp().Driver().AllWindows()[0])
				fileSelectFunc = func() {
					// resize dialog
					if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
						dialogSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
						dialogSize.Height = dialogSize.Height - 50
						dialogSize.Width = dialogSize.Width - 50
						fileDialog.Resize(dialogSize)
					}

					// update dialog initpath on change
					fileLister, currentFilename = _getFilePathDialogInitPath(v, entry)
					fileDialog.SetLocation(fileLister)

					fileDialog.Show()
				}
			}

			fileSelectButton := widget.NewButton(selectButtonLabel, fileSelectFunc)

			settingsFields = append(settingsFields, container.NewBorder(nil, nil, widget.NewLabel(settingName), fileSelectButton, entry))
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
	sendMessage := Fields.SendMessageStruct{
		Type:  "setting_change",
		Name:  "plugin_settings",
		Value: SettingsFile.Plugin_settings,
	}
	sendMessage.SendMessage()
}
