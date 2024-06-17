package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
			SettingsName:         "OSC minimum time between messages",
			SettingsInternalName: "osc_min_time_between_messages",
			SettingsDescription:  "The minimum allowed time between messages in seconds.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 10)
				sliderState := widget.NewLabel(fmt.Sprintf("%.1f", sliderWidget.Min))
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "OSC typing indicator and notification",
			SettingsInternalName: "osc_typing_indicator",
			SettingsDescription:  "Display a typing indicator if you are currently speaking or the A.I. is processing.\nAnd plays the notification sound on a new message.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "OSC chat prefix",
			SettingsInternalName: "osc_chat_prefix",
			SettingsDescription:  "Adds a prefix to the chat messages.\n\"{src}\" is replaced with the source language,\n\"{trg}\" replaced with the target language.",
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
			SettingsName:         "OSC text splitting send type",
			SettingsInternalName: "osc_send_type",
			SettingsDescription:  "How the OSC messages are send.\nChunks = Send the text in chunks if too long. Messages are separated via '...'.\nFull = Send the full text at once.\nFull or Scroll = Send the full text or scroll it if too long.\nScroll = Scroll the text.",
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
					{Text: "Send both 'Source and Translation'", Value: "both"},
					{Text: "Send both (inverted) 'Translation and Source'", Value: "both_inverted"},
				}, func(s CustomWidget.TextValueOption) {}, 0)
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC transfer split",
			SettingsInternalName: "osc_type_transfer_split",
			SettingsDescription:  "Text that is added between Source and Translation in the OSC message.\n(Only used if OSC Transfer Type is set to 'Both')\n\"{src}\" is replaced with the source language,\n\"{trg}\" replaced with the target language.",
			_widget: func() fyne.CanvasObject {
				settingWidget := widget.NewEntry()
				settingWidget.OnChanged = func(value string) {}
				return settingWidget
			},
		},
		{
			SettingsName:         "OSC delay until audio playback",
			SettingsInternalName: "osc_delay_until_audio_playback",
			SettingsDescription:  "Delays the OSC message until the audio playback started.\n(To better sync the audio with the text)\n(If no audio is played, the message will be delayed until a timeout is reached)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}
