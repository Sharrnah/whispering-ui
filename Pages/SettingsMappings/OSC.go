package SettingsMappings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
)

var OSCSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "OSC IP address",
			SettingsInternalName: "osc_ip",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC Port",
			SettingsInternalName: "osc_port",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC typing indicator",
			SettingsInternalName: "osc_typing_indicator",
			SettingsDescription:  "Display a typing indicator if you are currently speaking or the A.I. is processing.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "OSC chat prefix",
			SettingsInternalName: "osc_chat_prefix",
			SettingsDescription:  "Adds a prefix to the chat messages.",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC initial time limit",
			SettingsInternalName: "osc_initial_time_limit",
			SettingsDescription:  "Display time of the first OSC message in seconds.",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC time limit",
			SettingsInternalName: "osc_time_limit",
			SettingsDescription:  "Time between OSC messages in seconds.",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC scroll time limit",
			SettingsInternalName: "osc_scroll_time_limit",
			SettingsDescription:  "Time between scrolling OSC messages in seconds.\n(Only used if OSC Send Type is set to 'Scroll')",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC send type",
			SettingsInternalName: "osc_send_type",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				settingWidget := CustomWidget.NewTextValueSelect("osc_send_type", []CustomWidget.TextValueOption{
					{Text: "Chunks", Value: "chunks"},
					{Text: "Full", Value: "full"},
					{Text: "Full or Scroll", Value: "full_or_scroll"},
					{Text: "Scroll", Value: "scroll"},
				}, func(s CustomWidget.TextValueOption) {}, 0)
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC transfer type",
			SettingsInternalName: "osc_type_transfer",
			SettingsDescription:  "Type of OSC message to send.\nOnly Translation, Both or Source.",
			_widget: func() fyne.CanvasObject {
				settingWidget := CustomWidget.NewTextValueSelect("osc_type_transfer", []CustomWidget.TextValueOption{
					{Text: "Send Translation", Value: "translation_result"},
					{Text: "Send Source Text", Value: "source"},
					{Text: "Send both Soruce and Translation", Value: "both"},
				}, func(s CustomWidget.TextValueOption) {}, 0)
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC transfer split",
			SettingsInternalName: "osc_type_transfer_split",
			SettingsDescription:  "Text that is added between Source and Translation in the OSC message.\n(Only used if OSC Transfer Type is set to 'Both')",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
	},
}
