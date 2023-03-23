package Settings

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Utilities"
)

//goland:noinspection GoSnakeCaseUsage
type Conf struct {
	// Internal Profile Settings
	SettingsFilename string
	Device_index     interface{} `yaml:"device_index,omitempty" json:"device_index,omitempty"`
	Device_out_index interface{} `yaml:"device_out_index,omitempty" json:"device_out_index,omitempty"`

	Phrase_time_limit float64 `yaml:"phrase_time_limit" json:"phrase_time_limit"`
	Pause             float64 `yaml:"pause" json:"pause"`
	Energy            int     `yaml:"energy" json:"energy"`

	// VAD Settings
	Vad_enabled              bool   `yaml:"vad_enabled" json:"vad_enabled"`
	Vad_on_full_clip         bool   `yaml:"vad_on_full_clip" json:"vad_on_full_clip"`
	Vad_confidence_threshold string `yaml:"vad_confidence_threshold" json:"vad_confidence_threshold"`
	Vad_num_samples          int    `yaml:"vad_num_samples" json:"vad_num_samples"`
	Vad_thread_num           int    `yaml:"vad_thread_num,omitempty" json:"vad_thread_num,omitempty"`

	// Whisper Settings
	Ai_device                  interface{} `yaml:"ai_device" json:"ai_device"`
	Whisper_task               string      `yaml:"whisper_task" json:"whisper_task"`
	Current_language           string      `yaml:"current_language" json:"current_language"`
	Model                      string      `yaml:"model" json:"model"`
	Condition_on_previous_text bool        `yaml:"condition_on_previous_text" json:"condition_on_previous_text"`
	Initial_prompt             string      `yaml:"initial_prompt" json:"initial_prompt"`
	Logprob_threshold          string      `yaml:"logprob_threshold" json:"logprob_threshold"`     // string formatted float or "none" / ""
	No_speech_threshold        string      `yaml:"no_speech_threshold" json:"no_speech_threshold"` // string formatted float or "none" / ""
	Whisper_precision          string      `yaml:"whisper_precision" json:"whisper_precision"`
	Faster_whisper             bool        `yaml:"faster_whisper" json:"faster_whisper"`             // use faster whisper model
	Temperature_fallback       bool        `yaml:"temperature_fallback" json:"temperature_fallback"` // enables/disables temperature fallback (to prevent multiple whisper loops in a row)
	Beam_size                  int         `yaml:"beam_size" json:"beam_size"`
	Whisper_cpu_threads        int         `yaml:"whisper_cpu_threads" json:"whisper_cpu_threads"`
	Whisper_num_workers        int         `yaml:"whisper_num_workers" json:"whisper_num_workers"`
	Realtime                   bool        `yaml:"realtime" json:"realtime"`
	Realtime_frame_multiply    int         `yaml:"realtime_frame_multiply" json:"realtime_frame_multiply"`
	Realtime_whisper_model     string      `yaml:"realtime_whisper_model" json:"realtime_whisper_model"`
	Realtime_whisper_precision string      `yaml:"realtime_whisper_precision" json:"realtime_whisper_precision"`
	Realtime_whisper_beam_size int         `yaml:"realtime_whisper_beam_size" json:"realtime_whisper_beam_size"`

	// text translate settings
	Txt_translate            bool   `yaml:"txt_translate" json:"txt_translate"`
	Txt_translator_device    string `yaml:"txt_translator_device" json:"txt_translator_device"`
	Src_lang                 string `yaml:"src_lang" json:"src_lang"`
	Trg_lang                 string `yaml:"trg_lang" json:"trg_lang"`
	Txt_ascii                bool   `yaml:"txt_ascii" json:"txt_ascii"`
	Txt_translator           string `yaml:"txt_translator" json:"txt_translator"`
	Txt_translator_size      string `yaml:"txt_translator_size" json:"txt_translator_size"`
	Txt_translator_precision string `yaml:"txt_translator_precision" json:"txt_translator_precision"`
	Txt_translate_realtime   bool   `yaml:"txt_translate_realtime" json:"txt_translate_realtime"`

	// websocket settings
	Websocket_ip   string `yaml:"websocket_ip" json:"websocket_ip"`
	Websocket_port int    `yaml:"websocket_port" json:"websocket_port"`
	Run_backend    bool   `yaml:"run_backend" json:"run_backend"`

	// OSC settings
	Osc_ip                      string `yaml:"osc_ip" json:"osc_ip"`
	Osc_port                    int    `yaml:"osc_port" json:"osc_port"`
	Osc_address                 string `yaml:"osc_address" json:"osc_address"`
	Osc_typing_indicator        bool   `yaml:"osc_typing_indicator" json:"osc_typing_indicator"`
	Osc_convert_ascii           bool   `yaml:"osc_convert_ascii" json:"osc_convert_ascii"`
	Osc_auto_processing_enabled bool   `yaml:"osc_auto_processing_enabled" json:"osc_auto_processing_enabled"`
	Osc_chat_prefix             string `yaml:"osc_chat_prefix" json:"osc_chat_prefix"`

	// OCR settings
	Ocr_lang        string `yaml:"ocr_lang" json:"ocr_lang"`
	Ocr_window_name string `yaml:"ocr_window_name" json:"ocr_window_name"`

	// TTS settings
	Tts_enabled       bool     `yaml:"tts_enabled" json:"tts_enabled"`
	Tts_ai_device     string   `yaml:"tts_ai_device" json:"tts_ai_device"`
	Tts_answer        bool     `yaml:"tts_answer" json:"tts_answer"`
	Tts_model         []string `yaml:"tts_model" json:"tts_model"`
	Tts_voice         string   `yaml:"tts_voice" json:"tts_voice"`
	Tts_prosody_rate  string   `yaml:"tts_prosody_rate" json:"tts_prosody_rate"`
	Tts_prosody_pitch string   `yaml:"tts_prosody_pitch" json:"tts_prosody_pitch"`

	// Plugin Settings
	Plugins              map[string]bool `yaml:"plugins,omitempty" json:"plugins,omitempty"`
	Plugin_settings      interface{}     `yaml:"plugin_settings,omitempty" json:"plugin_settings,omitempty"`
	Plugin_timer_timeout interface{}     `yaml:"plugin_timer_timeout,omitempty" json:"plugin_timer_timeout,omitempty"`
	Plugin_timer         interface{}     `yaml:"plugin_timer,omitempty" json:"plugin_timer,omitempty"`
}

var ConfigValues map[string]interface{} = nil

// ExcludeConfigFields excludes fields from settings window (all lowercase)
var ExcludeConfigFields = []string{
	"settingsfilename",
	"tts_model",
	"tts_answer",
	"device_index",
	"device_out_index",
	"current_language",
	"plugins",
	"plugin_settings",
	"plugin_timer_timeout",
	"plugin_timer",
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
		err := fmt.Errorf("Config file %s not found", configFile)
		log.Printf("Error: %v", err)
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
	}

	return c
}

func (c *Conf) GetConf(configFile string) *Conf {
	return confLoader(c, configFile).(*Conf)
}

func (c *Conf) GetOption(option string) (interface{}, error) {
	option = cases.Title(language.English, cases.Compact).String(option)
	fieldByName := reflect.ValueOf(c).Elem().FieldByName(option)
	if fieldByName.IsValid() {
		return fieldByName.Interface(), nil
	}
	return nil, fmt.Errorf("option %s not found", option)
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
			case reflect.Float64:
				switch value.(type) {
				case string:
					tmpValue, _ := strconv.ParseFloat(value.(string), 64)
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

func MergeSettings(firstConf Conf, secondConf Conf) Conf {
	yamlFileFirst, err := yaml.Marshal(firstConf)
	if err != nil {
		log.Printf("error: %v", err)
	}
	yamlFileSecond, err := yaml.Marshal(secondConf)
	if err != nil {
		log.Printf("error: %v", err)
	}

	var mergedConf Conf
	err = yaml.Unmarshal(yamlFileFirst, &mergedConf)
	if err != nil {
		log.Printf("error: %v", err)
	}
	err = yaml.Unmarshal(yamlFileSecond, &mergedConf)
	if err != nil {
		log.Printf("error: %v", err)
	}
	return mergedConf
}

func BuildSettingsForm(includeConfigFields []string, settingsFile string) fyne.CanvasObject {
	settingsForm := widget.NewForm()

	settingsForm.Append("Profile", widget.NewLabel(Config.SettingsFilename))

	// merge local settings with settings file if running local backend
	var settingsFileConf = Conf{}
	MergedConfig := settingsFileConf
	if Config.Run_backend {
		settingsFileConf.GetConf(settingsFile)
		MergedConfig = MergeSettings(Config, settingsFileConf)
	} else {
		MergedConfig = Config
	}

	settingsFields := reflect.ValueOf(MergedConfig)

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

				if settingsValues != nil {
					if len(settingsValues) > 0 {
						settingsWidget := widget.NewSelect(settingsValues, func(s string) {
							println(s)
						})

						selectedValue := strconv.Itoa(settingsValue.(int))
						settingsWidget.SetSelected(selectedValue)
						settingsForm.Append(settingsName, settingsWidget)
					}
				} else {
					settingsWidget := widget.NewEntry()
					settingsWidget.SetText(strconv.Itoa(settingsValue.(int)))
					settingsForm.Append(settingsName, settingsWidget)
				}

			case "float64":
				if settingsValues != nil {
					if len(settingsValues) > 0 {
						settingsWidget := widget.NewSelect(settingsValues, func(s string) {
							println(s)
						})

						selectedValue := strconv.FormatFloat(settingsValue.(float64), 'f', 1, 64)
						settingsWidget.SetSelected(selectedValue)
						settingsForm.Append(settingsName, settingsWidget)
					}
				} else {
					settingsWidget := widget.NewEntry()
					settingsWidget.SetText(strconv.FormatFloat(settingsValue.(float64), 'f', 1, 64))
					settingsForm.Append(settingsName, settingsWidget)
				}

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
			needsSettingUpdate := false
			for _, item := range settingsForm.Items {
				var value interface{} = nil
				switch item.Widget.(type) {
				case *widget.Entry:
					value = item.Widget.(*widget.Entry).Text
					if value == "None" {
						value = nil
					}
				case *widget.Select:
					value = item.Widget.(*widget.Select).Selected
					if value == "None" {
						value = nil
					}
				case *widget.Check:
					value = item.Widget.(*widget.Check).Checked
				}

				preChangeOption, err := MergedConfig.GetOption(item.Text)
				sendValue := value
				if err == nil {
					switch preChangeOption.(type) {
					case int:
						switch value.(type) {
						case string:
							sendValue, _ = strconv.Atoi(value.(string))
						}
					case float64:
						switch value.(type) {
						case string:
							sendValue, _ = strconv.ParseFloat(value.(string), 64)
						}
					}
					if preChangeOption != sendValue {
						needsSettingUpdate = true
						sendMessage := Fields.SendMessageStruct{
							Type:  "setting_change",
							Name:  item.Text,
							Value: sendValue,
						}
						sendMessage.SendMessage()

						Config.SetOption(item.Text, value)
						MergedConfig.SetOption(item.Text, value)
					}
				}
			}
			if needsSettingUpdate {
				sendMessage := Fields.SendMessageStruct{
					Type: "setting_update_req",
				}
				sendMessage.SendMessage()
			}

			if Config.Run_backend {
				MergedConfig.WriteYamlSettings(settingsFile)
			}

			dialog.ShowInformation("Settings Saved", "Settings have been saved to "+settingsFile+"\n This might require a restart of the application for some changes to take effect.", fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}

	return settingsForm
}
