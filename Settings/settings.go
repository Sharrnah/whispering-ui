package Settings

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"reflect"
	"strconv"
)

//goland:noinspection GoSnakeCaseUsage
type Conf struct {
	// Whisper Settings
	Ai_device                  string
	Whisper_task               string
	Current_language           string
	Model                      string
	Condition_on_previous_text bool

	// text translate settings
	Txt_translate       bool
	Src_lang            string
	Trg_lang            string
	Txt_ascii           bool
	Txt_translator      string
	Txt_translator_size string

	// websocket settings
	Websocket_ip   string
	Websocket_port int

	// OSC settings
	Osc_ip               string
	Osc_port             int
	Osc_address          string
	Osc_typing_indicator bool
	Osc_convert_ascii    bool

	// OCR settings
	Ocr_lang        string
	Ocr_window_name string

	// TTS settings
	Tts_enabled      bool
	Tts_ai_device    string
	Tts_answer       bool
	Device_out_index int
	Tts_model        []string
	Tts_voice        string

	// FLAN-T5 settings
	Flan_enabled                       bool
	Flan_size                          string
	Flan_bits                          int
	Flan_device                        string
	Flan_whisper_answer                bool
	Flan_process_only_questions        bool
	Flan_osc_prefix                    string
	Flan_translate_to_speaker_language bool
	Flan_prompt                        string
	Flan_memory                        string
	//Flan_conditioning_history          int
}

var Config Conf

var (
	ErrNoValue      = errors.New("no value for field 'value'")
	ErrInvalidValue = errors.New("invalid value for field 'value'")
)

/*func (c *conf) UnmarshalYAML(unmarshal func(interface{}) error) error {
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
		log.Printf("settings.yaml not found (Press Enter to exit)")
		fmt.Scanln()
		os.Exit(1)
	}

	return c
}

func (c *Conf) GetConf(configFile string) *Conf {
	return confLoader(c, configFile).(*Conf)
}

var Form *widget.Form

func BuildSettingsForm() fyne.CanvasObject {
	settingsForm := widget.NewForm()

	settingsFields := reflect.ValueOf(Config)

	for i := 0; i < settingsFields.NumField(); i++ {
		if settingsFields.Field(i).CanInterface() {
			settingsName := settingsFields.Type().Field(i).Name
			settingsValue := settingsFields.Field(i).Interface()
			settingsType := settingsFields.Field(i).Type().Name()

			switch settingsType {
			case "string":
				settingsWidget := widget.NewEntry()
				settingsWidget.SetText(settingsValue.(string))
				settingsForm.Append(settingsName, settingsWidget)
			case "int":
				//settingsWidget := widget.NewSlider(0, 100)
				//settingsWidget.SetValue(float64(settingsValue.(int)))
				//settingsForm.Append(settingsName, settingsWidget)

				settingsWidget := widget.NewEntry()
				settingsWidget.SetText(strconv.Itoa(settingsValue.(int)))
				settingsForm.Append(settingsName, settingsWidget)
			case "bool":
				settingsWidget := widget.NewCheck("", func(checked bool) {})
				settingsForm.Append(settingsName, settingsWidget)
			}
		}
	}

	return settingsForm
}
