package Settings

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"whispering-tiger-ui/Utilities"
)

//goland:noinspection GoSnakeCaseUsage
type Conf struct {
	// Internal Profile Settings
	SettingsFilename string
	Device_index     interface{} `yaml:"device_index,omitempty"`
	Device_out_index interface{} `yaml:"device_out_index,omitempty"`

	Phrase_time_limit float64 `yaml:"phrase_time_limit,omitempty"`
	Pause             float64 `yaml:"pause,omitempty"`
	Energy            int     `yaml:"energy,omitempty"`

	// Whisper Settings
	Ai_device                  interface{} `yaml:"ai_device"`
	Whisper_task               string      `yaml:"whisper_task"`
	Current_language           string      `yaml:"current_language"`
	Model                      string      `yaml:"model"`
	Condition_on_previous_text bool        `yaml:"condition_on_previous_text"`

	// text translate settings
	Txt_translate         bool   `yaml:"txt_translate"`
	Txt_translator_device string `yaml:"txt_translator_device"`
	Src_lang              string `yaml:"src_lang"`
	Trg_lang              string `yaml:"trg_lang"`
	Txt_ascii             bool   `yaml:"txt_ascii"`
	Txt_translator        string `yaml:"txt_translator"`
	Txt_translator_size   string `yaml:"txt_translator_size"`

	// websocket settings
	Websocket_ip   string `yaml:"websocket_ip"`
	Websocket_port int    `yaml:"websocket_port"`

	// OSC settings
	Osc_ip                      string `yaml:"osc_ip"`
	Osc_port                    int    `yaml:"osc_port"`
	Osc_address                 string `yaml:"osc_address"`
	Osc_typing_indicator        bool   `yaml:"osc_typing_indicator"`
	Osc_convert_ascii           bool   `yaml:"osc_convert_ascii"`
	Osc_auto_processing_enabled bool   `yaml:"osc_auto_processing_enabled"`

	// OCR settings
	Ocr_lang        string `yaml:"ocr_lang"`
	Ocr_window_name string `yaml:"ocr_window_name"`

	// TTS settings
	Tts_enabled   bool     `yaml:"tts_enabled"`
	Tts_ai_device string   `yaml:"tts_ai_device"`
	Tts_answer    bool     `yaml:"tts_answer"`
	Tts_model     []string `yaml:"tts_model"`
	Tts_voice     string   `yaml:"tts_voice"`

	// FLAN-T5 settings
	Flan_enabled                       bool   `yaml:"flan_enabled"`
	Flan_size                          string `yaml:"flan_size"`
	Flan_bits                          int    `yaml:"flan_bits"`
	Flan_device                        string `yaml:"flan_device"`
	Flan_whisper_answer                bool   `yaml:"flan_whisper_answer"`
	Flan_process_only_questions        bool   `yaml:"flan_process_only_questions"`
	Flan_osc_prefix                    string `yaml:"flan_osc_prefix"`
	Flan_translate_to_speaker_language bool   `yaml:"flan_translate_to_speaker_language"`
	Flan_prompt                        string `yaml:"flan_prompt"`
	Flan_memory                        string `yaml:"flan_memory"`
	Flan_conditioning_history          int    `yaml:"flan_conditioning_history"`
}

var ConfigValues map[string]interface{} = nil

// ExcludeConfigFields excludes fields from settings window (all lowercase)
var ExcludeConfigFields = []string{
	"settingsfilename",
	"tts_model",
	"device_index",
	"device_out_index",
	"current_language",
}

var Config Conf

/*
var (
	ErrNoValue      = errors.New("no value for field 'value'")
	ErrInvalidValue = errors.New("invalid value for field 'value'")
)

func (c *conf) UnmarshalYAML(unmarshal func(interface{}) error) error {
	mstr := make(map[string]string)
	if err := unmarshal(&mstr); err == nil {
		if str, ok := mstr["value"]; ok {
			c.TtsModel = []string{str}
			return nil
		}

		return ErrNoValue
	}

	miface := make(map[interface{}]interface{})
	if err := unmarshal(&miface); err == nil {
		sstr := make([]string, 0)
		if val, ok := miface["value"]; ok {
			for _, v := range val.([]interface{}) {
				if str, ok := v.(string); ok {
					sstr = append(sstr, str)
				}
			}

			c.TtsModel = sstr
			return nil
		}

		return ErrNoValue
	}

	return ErrInvalidValue
}*/

// FileExists checks a file's existence
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func confLoader(c interface{}, configFile string) interface{} {
	if FileExists(configFile) == true {
		yamlFile, err := os.ReadFile(configFile)
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(yamlFile, c)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	} else {
		log.Printf("settings yaml not found (Press Enter to exit)")
		fmt.Scanln()
		os.Exit(1)
	}

	return c
}

func (c *Conf) GetConf(configFile string) *Conf {
	return confLoader(c, configFile).(*Conf)
}

func (c *Conf) SetOption(optionName string, value interface{}) {
	switch value.(type) {
	case string:
		// if string value is an integer, convert it
		intValue, err := strconv.Atoi(value.(string))
		if err == nil {
			value = intValue
		}
	}

	values := reflect.ValueOf(c)
	indirectValues := reflect.Indirect(values) // required to indirect the pointer
	types := indirectValues.Type()
	for i := 0; i < indirectValues.NumField(); i++ {
		if strings.ToLower(types.Field(i).Name) == strings.ToLower(optionName) {
			setValue := reflect.ValueOf(value)
			if value == nil {
				setValue = reflect.Zero(types.Field(i).Type)
			}
			switch types.Field(i).Type.Kind() {
			// TODO: fix case where the value is a string and the field is a slice (like in tts_model)
			case reflect.Slice:
				switch value.(type) {
				case string:
					setValue = reflect.ValueOf([]string{value.(string)})
				}
			case reflect.String:
				switch value.(type) {
				case int:
					setValue = reflect.ValueOf(strconv.Itoa(value.(int)))
				}
			case reflect.Int:
				switch value.(type) {
				case string:
					tmpValue, _ := strconv.Atoi(value.(string))
					setValue = reflect.ValueOf(tmpValue)
				}

			}
			indirectValues.Field(i).Set(setValue)
			return
		}
	}
}

func (c *Conf) LoadYamlSettings(fileName string) error {
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		//log.Fatalf("Unmarshal: %v", err)
		return err
	}
	return nil
}

func (c *Conf) WriteYamlSettings(fileName string) {
	// marshal the struct to yaml and save as file
	yamlFile, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("error: %v", err)
	}
	err = os.WriteFile(fileName, yamlFile, 0644)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

var Form *widget.Form

func GetSettingValues(settingField string) ([]string, error) {
	if _, ok := ConfigValues[settingField]; ok {
		var values []string
		switch ConfigValues[settingField].(type) {
		case []interface{}:
			for i := 0; i < len(ConfigValues[settingField].([]interface{})); i++ {
				values = append(values, ConfigValues[settingField].([]interface{})[i].(string))
			}
		}
		return values, nil
	}
	return nil, errors.New("no values for field '" + settingField + "'")
}

func BuildSettingsForm(includeConfigFields []string, settingsFile string) fyne.CanvasObject {
	settingsForm := widget.NewForm()

	settingsForm.Append("Profile", widget.NewLabel(Config.SettingsFilename))

	settingsFields := reflect.ValueOf(Config)

	for i := 0; i < settingsFields.NumField(); i++ {
		if settingsFields.Field(i).CanInterface() {
			settingsName := strings.ToLower(settingsFields.Type().Field(i).Name)

			// check if settingsName is in field List
			if includeConfigFields != nil && !Utilities.Contains(includeConfigFields, settingsName) {
				continue
			}
			if ExcludeConfigFields != nil && Utilities.Contains(ExcludeConfigFields, settingsName) {
				continue
			}

			settingsValue := settingsFields.Field(i).Interface()
			settingsType := settingsFields.Field(i).Type().Name()
			settingsValues, _ := GetSettingValues(settingsName)

			switch settingsType {
			case "string":
				if settingsValues != nil {
					if len(settingsValues) > 0 {
						settingsWidget := widget.NewSelect(settingsValues, func(s string) {
							println(s)
						})

						selectedValue := settingsValue.(string)
						if selectedValue == "" {
							selectedValue = "None"
						}
						settingsWidget.SetSelected(selectedValue)
						settingsForm.Append(settingsName, settingsWidget)
					} else {
						settingsWidget := widget.NewEntry()
						settingsWidget.SetText(settingsValue.(string))
						settingsWidget.Disable()
						settingsForm.Append(settingsName, settingsWidget)
					}
				} else {
					settingsWidget := widget.NewEntry()
					settingsWidget.SetText(settingsValue.(string))
					settingsForm.Append(settingsName, settingsWidget)
				}
			case "":
				if len(settingsValues) > 0 {
					settingsWidget := widget.NewSelect(settingsValues, func(s string) {
						println(s)
					})

					if settingsValue != nil {
						switch settingsValue.(type) {
						case string:
							selectedValue := settingsValue.(string)
							settingsWidget.SetSelected(selectedValue)
						}
					} else {
						settingsWidget.SetSelected("None")
					}

					settingsForm.Append(settingsName, settingsWidget)
				} else {
					settingsWidget := widget.NewEntry()
					settingsForm.Append(settingsName, settingsWidget)
				}

			case "int":
				//settingsWidget := widget.NewSlider(0, 100)
				//settingsWidget.SetValue(float64(settingsValue.(int)))
				//settingsForm.Append(settingsName, settingsWidget)

				settingsWidget := widget.NewEntry()
				settingsWidget.SetText(strconv.Itoa(settingsValue.(int)))
				settingsForm.Append(settingsName, settingsWidget)
			case "bool":
				settingsWidget := widget.NewCheck("", func(checked bool) {})
				settingsWidget.Checked = settingsValue.(bool)
				settingsForm.Append(settingsName, settingsWidget)

			}
		}
	}

	settingsForm.SubmitText = "Save"

	if settingsFile != "" {
		settingsForm.OnSubmit = func() {
			for _, item := range settingsForm.Items {
				var value interface{} = nil
				switch item.Widget.(type) {
				case *widget.Entry:
					value = item.Widget.(*widget.Entry).Text
					if value == "None" {
						value = nil
					}
					Config.SetOption(item.Text, value)
				case *widget.Select:
					value = item.Widget.(*widget.Select).Selected
					if value == "None" {
						value = nil
					}
					Config.SetOption(item.Text, value)
				case *widget.Check:
					value = item.Widget.(*widget.Check).Checked
					Config.SetOption(item.Text, value)
				}

			}
			//Settings.Form.Items[0].Widget.(*widget.Entry).SetText(Settings.Form.Items[0].Widget.(*widget.Entry).Text)

			Config.WriteYamlSettings(settingsFile)

			dialog.ShowInformation("Settings Saved", "Settings have been saved to "+settingsFile+"\n This requires a restart of the application currently.", fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}

	return settingsForm
}
