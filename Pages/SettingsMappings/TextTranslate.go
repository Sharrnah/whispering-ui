package SettingsMappings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var TextTranslateSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Translate text in real-time",
			SettingsInternalName: "txt_translate_realtime",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Convert text to romaji",
			SettingsInternalName: "txt_romaji",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}
