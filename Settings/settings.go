package Settings

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type conf struct {
	// Whisper Settings
	AiDevice                string `yaml:"ai_device"`
	WhisperTask             string `yaml:"whisper_task"`
	CurrentLanguage         string `yaml:"current_language"`
	Model                   string `yaml:"model"`
	ConditionOnPreviousText string `yaml:"condition_on_previous_text"`

	// text translate settings
	TxtTranslate      bool   `yaml:"txt_translate"`
	SrcLang           string `yaml:"src_lang"`
	TrgLang           string `yaml:"trg_lang"`
	TxtAscii          string `yaml:"txt_ascii"`
	TxtTranslator     string `yaml:"txt_translator"`
	TxtTranslatorSize string `yaml:"txt_translator_size"`

	// websocket settings
	WebsocketIp   string `yaml:"websocket_ip"`
	WebsocketPort int    `yaml:"websocket_port"`

	// OSC settings
	OscIp              string `yaml:"osc_ip"`
	OscPort            int    `yaml:"osc_port"`
	OscAddress         string `yaml:"osc_address"`
	OscTypingIndicator bool   `yaml:"osc_typing_indicator"`
	OscConvertAscii    bool   `yaml:"osc_convert_ascii"`

	// OCR settings
	OcrLang       string `yaml:"ocr_lang"`
	OcrWindowName string `yaml:"ocr_window_name"`

	// TTS settings
	TtsEnabled     bool     `yaml:"tts_enabled"`
	TtsAiDevice    string   `yaml:"tts_ai_device"`
	TtsAnswer      bool     `yaml:"tts_answer"`
	DeviceOutIndex int      `yaml:"device_out_index"`
	TtsModel       []string `yaml:"tts_model"`
	TtsVoice       string   `yaml:"tts_voice"`

	// FLAN-T5 settings
	FlanEnabled                    bool   `yaml:"flan_enabled"`
	FlanSize                       string `yaml:"flan_size"`
	FlanBits                       int    `yaml:"flan_bits"`
	FlanDevice                     string `yaml:"flan_device"`
	FlanWhisperAnswer              bool   `yaml:"flan_whisper_answer"`
	FlanProcessOnlyQuestions       bool   `yaml:"flan_process_only_questions"`
	FlanOscPrefix                  string `yaml:"flan_osc_prefix"`
	FlanTranslateToSpeakerLanguage bool   `yaml:"flan_translate_to_speaker_language"`
	FlanPrompt                     string `yaml:"flan_prompt"`
	FlanMemory                     string `yaml:"flan_memory"`
	FlanConditioningHistory        int    `yaml:"flan_conditioning_history"`
}

var Config conf

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

func (c *conf) GetConf(configFile string) *conf {
	return confLoader(c, configFile).(*conf)
}
