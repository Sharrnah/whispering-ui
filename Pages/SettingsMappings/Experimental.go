package SettingsMappings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var ExperimentalSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Start downloads using the UI. (Recommended)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("DisableUiDownloads", !b)
				})
				widgetCheckbox.Checked = !fyne.CurrentApp().Preferences().BoolWithFallback("DisableUiDownloads", false)

				return widgetCheckbox
			},
		},
		{
			SettingsName:         "Run Python backend with UTF-8 encoding. (Recommended)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("RunWithUTF8", b)
				})
				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("RunWithUTF8", true)

				return widgetCheckbox
			},
		},
		{
			SettingsName:         "Focus window on message receive. (Can improve speed in VR)",
			SettingsInternalName: "",
			SettingsDescription:  "",
			DoNotSendToBackend:   true,
			_widget: func() fyne.CanvasObject {
				widgetCheckbox := widget.NewCheck("", func(b bool) {
					fyne.CurrentApp().Preferences().SetBool("AutoRefocusWindow", b)
				})
				widgetCheckbox.Checked = fyne.CurrentApp().Preferences().BoolWithFallback("AutoRefocusWindow", false)

				return widgetCheckbox
			},
		},
	},
}
