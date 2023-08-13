package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
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
		sendMessage := Fields.SendMessageStruct{
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
		sendMessage := Fields.SendMessageStruct{
			Type:  "setting_change",
			Name:  s.SettingsInternalName,
			Value: value,
		}
		sendMessage.SendMessage()
		Settings.Config.SetOption(s.SettingsInternalName, value)
		fmt.Println("sent message with value" + fmt.Sprintf("%v", value))
	})
}

func (s *SettingMapping) processWidget(settingsValue interface{}, settingWidget interface{}, topMost bool) {
	//fyneWidget := widget.(*fyne.Widget)
	switch settingWidget.(type) {
	case *widget.Check:
		value := settingsValue.(bool)
		settingWidget.(*widget.Check).SetChecked(value)
		println("setting check value to " + fmt.Sprintf("%t", value))
		// Update OnChange callback to trigger settings update
		onChange := settingWidget.(*widget.Check).OnChanged
		settingWidget.(*widget.Check).OnChanged = func(b bool) {
			onChange(b)
			if !s.DoNotSendToBackend {
				s.SendUpdatedValue(b)
			}
		}
	case *widget.Slider:
		var originalType interface{}
		value := 0.0
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
	if topMost && s.SettingsDescription != "" {
		originalWidget := settingWidget.(fyne.CanvasObject)
		infoButton := widget.NewButtonWithIcon("Info", theme.InfoIcon(), func() {
			if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
				dialog.ShowInformation(s.SettingsName, s.SettingsDescription, fyne.CurrentApp().Driver().AllWindows()[0])
			}
		})
		s.Widget = container.NewBorder(nil, nil, nil, infoButton, originalWidget)
	}
}

func CreateSettingsFormByMapping(mappings SettingsMapping) *container.Scroll {

	settingsForm := widget.NewForm()

	// initialize widgets
	for _, mapping := range mappings.Mappings {
		println("mapping.SettingsName")
		println(mapping.SettingsName)
		println("mapping.SettingsInternalName")
		println(mapping.SettingsInternalName)
		mapping.Widget = mapping._widget()
		if mapping.Widget == nil {
			println("No widget created :thinking:")
		}

		singleMapping := mapping

		settingsFields := reflect.ValueOf(Settings.Config)
		for i := 0; i < settingsFields.NumField(); i++ {
			if settingsFields.Field(i).CanInterface() && strings.ToLower(settingsFields.Type().Field(i).Name) == singleMapping.SettingsInternalName {
				settingsName := strings.ToLower(settingsFields.Type().Field(i).Name)
				// check if the settings name is the same as the internal name
				_, err := mappings.FindSettingByInternalName(settingsName)
				if err != nil {
					continue
				} else {
					println("found setting")
				}
				settingsValue := settingsFields.Field(i).Interface()
				println("type of widget")
				println(reflect.TypeOf(singleMapping.Widget).Name())

				singleMapping.processWidget(settingsValue, singleMapping.Widget, true)
				break
			}
		}

		// add widget to form
		if singleMapping.Widget != nil {
			settingsForm.Append(singleMapping.SettingsName, singleMapping.Widget)
		}
	}

	return container.NewVScroll(settingsForm)
}
