package Profiles

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

//goland:noinspection GoSnakeCaseUsage
type Profile struct {
	SettingsFilename string
	Device_index     interface{} `yaml:"device_index"`
	Device_out_index interface{} `yaml:"device_out_index"`

	Audio_input_device  string `yaml:"audio_input_device"`
	Audio_output_device string `yaml:"audio_output_device"`

	Phrase_time_limit float64 `yaml:"phrase_time_limit"`
	Pause             float64 `yaml:"pause"`
	Energy            int     `yaml:"energy"`

	Vad_enabled              bool   `yaml:"vad_enabled"`
	Vad_on_full_clip         bool   `yaml:"vad_on_full_clip"`
	Vad_confidence_threshold string `yaml:"vad_confidence_threshold"`

	// Whisper Settings
	Ai_device                interface{} `yaml:"ai_device"`
	Model                    string      `yaml:"model"`
	Txt_translator_size      string      `yaml:"txt_translator_size"`
	Txt_translator_device    string      `yaml:"txt_translator_device"`
	Txt_translator_precision string      `yaml:"txt_translator_precision"`
	Websocket_ip             string      `yaml:"websocket_ip"`
	Websocket_port           int         `yaml:"websocket_port"`
	Run_Backend              bool        `yaml:"run_backend"`
	Osc_ip                   string      `yaml:"osc_ip"`
	Osc_port                 int         `yaml:"osc_port"`
	Tts_enabled              bool        `yaml:"tts_enabled"`
	Tts_ai_device            string      `yaml:"tts_ai_device"`
	Whisper_precision        string      `yaml:"whisper_precision"`
	Faster_whisper           bool        `yaml:"faster_whisper"`
	Realtime                 bool        `yaml:"realtime"`
}

func (p *Profile) Load(fileName string) {
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &p)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func (p *Profile) Save(fileName string) {
	// marshal the struct to yaml and save as file
	yamlFile, err := yaml.Marshal(p)
	if err != nil {
		log.Printf("error: %v", err)
	}
	err = os.WriteFile(fileName, yamlFile, 0644)
	if err != nil {
		log.Printf("error: %v", err)
	}
}
