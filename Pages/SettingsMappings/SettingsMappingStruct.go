package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

var timerLock sync.Mutex
var debounceTimers = make(map[string]*time.Timer)

const debounceDuration = 500 * time.Millisecond

const EnergySliderMax = 2000

type SettingsMapping struct {
	Mappings []SettingMapping
}

type SettingMapping struct {
	SettingsName         string
	SettingsInternalName string
	SettingsDescription  string
	_widget              func() fyne.CanvasObject
	Widget               fyne.CanvasObject
	DoNotSendToBackend   bool
}

func (s *SettingsMapping) FindSettingByInternalName(internalName string) (*SettingMapping, error) {
	for _, mapping := range s.Mappings {
		if mapping.SettingsInternalName == internalName {
			return &mapping, nil
		}
	}
	return nil, fmt.Errorf("could not find setting with internal name %s", internalName)
}

func (s *SettingMapping) SendUpdatedValue(value interface{}) {
	timerLock.Lock()
	defer timerLock.Unlock()

	// If the value is a boolean, send the update immediately
	if v, ok := value.(bool); ok {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  s.SettingsInternalName,
			Value: v,
		}
		sendMessage.SendMessage()
		Settings.Config.SetOption(s.SettingsInternalName, v)
		fmt.Println("sent message with value" + fmt.Sprintf("%v", v))
		return
	}

	// Check if there's an existing timer for this setting and stop it
	if timer, exists := debounceTimers[s.SettingsInternalName]; exists {
		timer.Stop()
	}

	// Set up a new timer that calls the actual message sending after debounceDuration
	debounceTimers[s.SettingsInternalName] = time.AfterFunc(debounceDuration, func() {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  s.SettingsInternalName,
			Value: value,
		}
		sendMessage.SendMessage()
		Settings.Config.SetOption(s.SettingsInternalName, value)
		fmt.Println("sent message with value" + fmt.Sprintf("%v", value))
	})
}

// processWidget processes the widgets of each setting
// settingsValue is the current value of the setting
// settingWidget is the widget of the setting
// _topMost is true if the widget is the topmost widget of the setting (used internally for recursive calls)
func (s *SettingMapping) processWidget(settingsValue interface{}, settingWidget interface{}, _topMost bool) {
	switch settingWidget.(type) {
	case *widget.Check:
		if settingsValue == nil {
			settingsValue = settingWidget.(*widget.Check).Checked
		}
		value := settingsValue.(bool)
		settingWidget.(*widget.Check).SetChecked(value)
		println("setting check value to " + fmt.Sprintf("%t", value))
		// Update OnChange callback to trigger settings update
		onChange := settingWidget.(*widget.Check).OnChanged
		settingWidget.(*widget.Check).OnChanged = func(b bool) {
			onChange(b)
			if b {
				settingWidget.(*widget.Check).SetText(lang.L("Enabled"))
			} else {
				settingWidget.(*widget.Check).SetText(lang.L("Disabled"))
			}
			if !s.DoNotSendToBackend {
				s.SendUpdatedValue(b)
			}
		}
		if value {
			settingWidget.(*widget.Check).SetText(lang.L("Enabled"))
		} else {
			settingWidget.(*widget.Check).SetText(lang.L("Disabled"))
		}
	case *widget.Slider:
		var originalType interface{}
		value := 0.0
		if settingsValue == nil {
			settingsValue = settingWidget.(*widget.Slider).Value
		}
		switch settingsValue.(type) {
		case float64:
			originalType = float64(0)
			value = settingsValue.(float64)
		case int:
			originalType = int(0)
			value = float64(settingsValue.(int))
		}
		println("setting slider value to " + fmt.Sprintf("%f", value))
		settingWidget.(*widget.Slider).SetValue(value)
		onChange := settingWidget.(*widget.Slider).OnChanged
		settingWidget.(*widget.Slider).OnChanged = func(value float64) {
			onChange(value)
			if !s.DoNotSendToBackend {
				// Convert value back to its original type
				switch originalType.(type) {
				case float64:
					convertedValue := float64(value)
					s.SendUpdatedValue(convertedValue)
					return
				case int:
					convertedValue := int(value)
					s.SendUpdatedValue(convertedValue)
					return
				}
				s.SendUpdatedValue(value)
			}
		}
	case *widget.Entry:
		var originalType interface{}
		value := ""
		if settingsValue == nil {
			settingsValue = settingWidget.(*widget.Entry).Text
		}
		switch v := settingsValue.(type) {
		case float64:
			originalType = float64(0)
			value = fmt.Sprintf("%f", v)
		case int:
			originalType = int(0)
			value = fmt.Sprintf("%d", v)
		case string:
			originalType = ""
			value = v
		}
		println("setting entry value to " + value)
		settingWidget.(*widget.Entry).SetText(value)
		onChange := settingWidget.(*widget.Entry).OnChanged
		settingWidget.(*widget.Entry).OnChanged = func(value string) {
			onChange(value)
			if !s.DoNotSendToBackend {
				// Convert value back to its original type
				switch originalType.(type) {
				case float64:
					convertedValue, _ := strconv.ParseFloat(value, 64)
					s.SendUpdatedValue(convertedValue)
				case int:
					convertedValue, _ := strconv.Atoi(value)
					s.SendUpdatedValue(convertedValue)
				case string:
					s.SendUpdatedValue(value)
				}
			}
		}
	case *CustomWidget.TextValueSelect:
		if settingsValue == nil {
			settingsValue = settingWidget.(*CustomWidget.TextValueSelect).Selected
		}
		value := settingsValue.(string)
		println("setting entry value to " + value)
		settingWidget.(*CustomWidget.TextValueSelect).SetSelected(value)
		onChange := settingWidget.(*CustomWidget.TextValueSelect).OnChanged
		settingWidget.(*CustomWidget.TextValueSelect).OnChanged = func(value CustomWidget.TextValueOption) {
			onChange(value)
			if !s.DoNotSendToBackend {
				s.SendUpdatedValue(value.Value)
			}
		}
	case *fyne.Container:
		for _, settingChild := range settingWidget.(*fyne.Container).Objects {
			s.processWidget(settingsValue, settingChild, false)
		}
	}

	// find translations for settings name and description
	settingsName := s.SettingsName
	SettingsDescription := s.SettingsDescription
	if s.SettingsInternalName != "" {
		settingsTranslateName := lang.L(s.SettingsInternalName + ".Name")
		if settingsTranslateName != "" && settingsTranslateName != s.SettingsInternalName+".Name" {
			settingsName = settingsTranslateName
		}
		settingsTranslateDescription := lang.L(s.SettingsInternalName + ".Description")
		if settingsTranslateDescription != "" && settingsTranslateDescription != s.SettingsInternalName+".Description" {
			SettingsDescription = settingsTranslateDescription
		}
	} else {
		settingsTranslateName := lang.L(settingsName)
		if settingsTranslateName != "" && settingsTranslateName != settingsName {
			settingsName = settingsTranslateName
		}
		settingsTranslateDescription := lang.L(SettingsDescription)
		if settingsTranslateDescription != "" && settingsTranslateDescription != SettingsDescription {
			SettingsDescription = settingsTranslateDescription
		}
	}

	if _topMost && SettingsDescription != "" {
		originalWidget := settingWidget.(fyne.CanvasObject)
		infoButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
				dialog.ShowInformation(settingsName, SettingsDescription, fyne.CurrentApp().Driver().AllWindows()[0])
			}
		})
		s.Widget = container.NewBorder(nil, nil, infoButton, nil, originalWidget)
	} else {
		originalWidget := settingWidget.(fyne.CanvasObject)
		spacer := canvas.NewText("       ", color.Transparent)
		s.Widget = container.NewBorder(nil, nil, container.New(layout.NewPaddedLayout(), spacer), nil, originalWidget)
	}
}

func CreateSettingsFormByMapping(mappings SettingsMapping) *container.Scroll {
	defer Utilities.PanicLogger()

	settingsForm := widget.NewForm()

	// initialize widgets
	for _, mapping := range mappings.Mappings {
		mapping.Widget = mapping._widget()
		if mapping.Widget == nil {
			println("No widget created :thinking:")
		}

		singleMapping := mapping

		if !mapping.DoNotSendToBackend {
			settingsFields := reflect.ValueOf(Settings.Config)
			for i := 0; i < settingsFields.NumField(); i++ {
				if settingsFields.Field(i).CanInterface() && strings.ToLower(settingsFields.Type().Field(i).Name) == singleMapping.SettingsInternalName {
					settingsName := strings.ToLower(settingsFields.Type().Field(i).Name)
					// check if the settings name is the same as the internal name
					_, err := mappings.FindSettingByInternalName(settingsName)
					if err != nil {
						continue
					}
					settingsValue := settingsFields.Field(i).Interface()
					singleMapping.processWidget(settingsValue, singleMapping.Widget, true)
					break
				}
			}
		} else {
			singleMapping.processWidget(nil, singleMapping.Widget, true)
		}

		// add widget to form
		if singleMapping.Widget != nil {
			// find translations for settings name
			settingsName := singleMapping.SettingsName
			if singleMapping.SettingsInternalName != "" {
				settingsTranslateName := lang.L(singleMapping.SettingsInternalName + ".Name")
				if settingsTranslateName != "" && settingsTranslateName != singleMapping.SettingsInternalName+".Name" {
					settingsName = settingsTranslateName
				}
			} else {
				settingsTranslateName := lang.L(settingsName)
				if settingsTranslateName != "" && settingsTranslateName != settingsName {
					settingsName = settingsTranslateName
				}
			}
			settingsForm.Append(settingsName, singleMapping.Widget)
		}
	}

	return container.NewVScroll(settingsForm)
}
