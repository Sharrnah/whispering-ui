package SettingsMappings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var TextToSpeechSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Integrated Text-to-Speech",
			SettingsInternalName: "tts_enabled",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Route text-to-speech to secondary audio device",
			SettingsInternalName: "tts_use_secondary_playback",
			SettingsDescription:  "Play text-to-speech on a secondary audio device at the same time as the selected output device.\n(By default uses windows default audio device)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}
