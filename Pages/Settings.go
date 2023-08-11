package Pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"reflect"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

var writeTimer *time.Timer
var timerLock sync.Mutex

const debounceDuration = 500 * time.Millisecond

type SettingsMapping struct {
	Mappings []SettingMapping
}

func (s *SettingsMapping) FindSettingByInternalName(internalName string) (*SettingMapping, error) {
	for _, mapping := range s.Mappings {
		if mapping.SettingsInternalName == internalName {
			return &mapping, nil
		}
	}
	return nil, fmt.Errorf("could not find setting with internal name %s", internalName)
}

type SettingMapping struct {
	SettingsName         string
	SettingsInternalName string
	SettingsDescription  string
	_widget              func() fyne.CanvasObject
	Widget               fyne.CanvasObject
}

func (s *SettingMapping) SendUpdatedValue(value interface{}) {
	sendMessage := Fields.SendMessageStruct{
		Type:  "setting_change",
		Name:  s.SettingsInternalName,
		Value: value,
	}
	sendMessage.SendMessage()
	Settings.Config.SetOption(s.SettingsInternalName, value)
	//reflect.ValueOf(&Settings.Config).Elem().FieldByName(Utilities.Capitalize(s.SettingsInternalName)).Set(reflect.ValueOf(value))
	//if Settings.Config.Run_backend {
	//	Settings.Config.WriteYamlSettings(Settings.Config.SettingsFilename)
	//}
	println("sent message with value" + fmt.Sprintf("%v", value))
}

var DisplaySettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Speech volume Level",
			SettingsInternalName: "energy",
			SettingsDescription:  "The volume level at which the speech detection will trigger.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, EnergySliderMax)
				sliderState := widget.NewLabel("0")
				sliderWidget.Step = 1
				sliderWidget.OnChanged = func(value float64) {
					if value >= sliderWidget.Max {
						sliderWidget.Max += 10
					}
					sliderState.SetText(fmt.Sprintf("%.0f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Speech pause detection",
			SettingsInternalName: "pause",
			SettingsDescription:  "The pause time in seconds after which the speech detection will stop and A.I. processing starts.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 5)
				sliderState := widget.NewLabel("0.0")
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Phrase time limit",
			SettingsInternalName: "phrase_time_limit",
			SettingsDescription:  "The max. time limit in seconds after which the audio processing starts.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 30)
				sliderState := widget.NewLabel("0.0")
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Realtime Mode",
			SettingsInternalName: "realtime",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Realtime Frequency",
			SettingsInternalName: "realtime_frequency_time",
			SettingsDescription:  "how often the audio is processed in seconds.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 20)
				sliderState := widget.NewLabel("0.0")
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "A.I. Denoise Audio",
			SettingsInternalName: "denoise_audio",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Apply Voice Markers to Audio",
			SettingsInternalName: "whisper_apply_voice_markers",
			SettingsDescription:  "Can reduce A.I. hallucinations, but does not work correctly with Speech Language set to \"Auto\".",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Cut silent audio parts",
			SettingsInternalName: "silence_cutting_enabled",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Normalize audio",
			SettingsInternalName: "normalize_enabled",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}

func CreateSettingsWindow() fyne.CanvasObject {
	settingsForm := widget.NewForm()

	// initialize widgets
	for _, mapping := range DisplaySettingsMapping.Mappings {
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
			println("init " + singleMapping.SettingsInternalName + " inter setting with field " + strings.ToLower(settingsFields.Type().Field(i).Name))
			if settingsFields.Field(i).CanInterface() && strings.ToLower(settingsFields.Type().Field(i).Name) == singleMapping.SettingsInternalName {
				settingsName := strings.ToLower(settingsFields.Type().Field(i).Name)
				// check if the settings name is the same as the internal name
				_, err := DisplaySettingsMapping.FindSettingByInternalName(settingsName)
				if err != nil {
					continue
				} else {
					println("found setting")
				}
				settingsValue := settingsFields.Field(i).Interface()
				println("type of widget")
				println(reflect.TypeOf(singleMapping.Widget).Name())

				switch singleMapping.Widget.(type) {
				case *widget.Check:
					singleMapping.Widget.(*widget.Check).SetChecked(settingsValue.(bool))
					println("setting check value to " + fmt.Sprintf("%t", settingsValue.(bool)))
					// Update OnChange callback to trigger settings update
					onChange := singleMapping.Widget.(*widget.Check).OnChanged
					singleMapping.Widget.(*widget.Check).OnChanged = func(b bool) {
						onChange(b)
						singleMapping.SendUpdatedValue(b)
					}
				case *widget.Slider:
					value := settingsValue.(float64)
					println("setting slider value to " + fmt.Sprintf("%f", value))
					singleMapping.Widget.(*widget.Slider).SetValue(value)
					onChange := singleMapping.Widget.(*widget.Slider).OnChanged
					singleMapping.Widget.(*widget.Slider).OnChanged = func(value float64) {
						onChange(value)
						singleMapping.SendUpdatedValue(value)
					}
				case *widget.Entry:
					value := settingsValue.(string)
					println("setting entry value to " + value)
					singleMapping.Widget.(*widget.Entry).SetText(value)
					onChange := singleMapping.Widget.(*widget.Entry).OnChanged
					singleMapping.Widget.(*widget.Entry).OnChanged = func(value string) {
						onChange(value)
						singleMapping.SendUpdatedValue(value)
					}
				case *fyne.Container:
					for _, settingChild := range singleMapping.Widget.(*fyne.Container).Objects {
						switch settingChild.(type) {
						case *widget.Slider:
							value := 0.0
							switch settingsValue.(type) {
							case float64:
								value = settingsValue.(float64)
							case int:
								value = float64(settingsValue.(int))
							}
							println("setting slider value to " + fmt.Sprintf("%f", value))
							settingChild.(*widget.Slider).SetValue(value)
							onChange := settingChild.(*widget.Slider).OnChanged
							settingChild.(*widget.Slider).OnChanged = func(value float64) {
								onChange(value)
								singleMapping.SendUpdatedValue(value)
							}
						}
					}
				}
				break
			}
		}

		// add widget to form
		if singleMapping.Widget != nil {
			settingsForm.Append(singleMapping.SettingsName, singleMapping.Widget)
		}
	}

	return settingsForm
}
