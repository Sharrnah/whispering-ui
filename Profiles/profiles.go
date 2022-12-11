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
	// Whisper Settings
	Ai_device           interface{} `yaml:"ai_device"`
	Model               string      `yaml:"model"`
	Txt_translator_size string      `yaml:"txt_translator_size"`
	Websocket_ip        string      `yaml:"websocket_ip"`
	Websocket_port      int         `yaml:"websocket_port"`
	Osc_ip              string      `yaml:"osc_ip"`
	Osc_port            int         `yaml:"osc_port"`
	Tts_enabled         bool        `yaml:"tts_enabled"`
	Tts_ai_device       string      `yaml:"tts_ai_device"`
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
