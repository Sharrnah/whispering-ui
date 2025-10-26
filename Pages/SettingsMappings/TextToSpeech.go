package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
)

var TextToSpeechSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		//{
		//	SettingsName:         "Integrated Text-to-Speech",
		//	SettingsInternalName: "tts_type",
		//	SettingsDescription:  "",
		//	_widget: func() fyne.CanvasObject {
		//		settingWidget := CustomWidget.NewTextValueSelect("tts_type", []CustomWidget.TextValueOption{
		//			{Text: "Silero", Value: "silero"},
		//			{Text: "F5/E2", Value: "f5_e2"},
		//			{Text: "Zonos", Value: "zonos"},
		//			{Text: "Kokoro", Value: "kokoro"},
		//			{Text: lang.L("Disabled"), Value: ""},
		//		}, func(s CustomWidget.TextValueOption) {
		//		}, 0)
		//		confirmationFunction := func(s CustomWidget.TextValueOption) {
		//			if Settings.Config.Tts_type != s.Value {
		//				RuntimeBackend.RestartBackend(true, lang.L("Restarting backend required to apply changes. Are you sure you want to restart the backend?"))
		//			}
		//		}
		//		settingWidget.OnChanged = confirmationFunction
		//		return settingWidget
		//	},
		//},
		{
			SettingsName:         "Speed/Rate",
			SettingsInternalName: "tts_prosody_rate",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				settingWidget := CustomWidget.NewTextValueSelect("tts_prosody_rate", []CustomWidget.TextValueOption{
					{Text: lang.L("Default"), Value: ""},
					{Text: "x-slow", Value: "x-slow"},
					{Text: "slow", Value: "slow"},
					{Text: "medium", Value: "medium"},
					{Text: "fast", Value: "fast"},
					{Text: "x-fast", Value: "x-fast"},
				}, func(s CustomWidget.TextValueOption) {}, 0)
				return settingWidget
			},
		},
		{
			SettingsName:         "Normalize Text-to-Speech Volume",
			SettingsInternalName: "tts_normalize",
			SettingsDescription:  "Automatically normalize the volume of the Text-to-Speech output.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Text-to-Speech Volume Adjustment",
			SettingsInternalName: "tts_volume",
			SettingsDescription:  "Adjust the volume of the text-to-speech output.\n1.0 = normal volume",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0.0, 2.0)
				sliderState := widget.NewLabel(fmt.Sprintf("%.2f", sliderWidget.Min))
				sliderWidget.Step = 0.01
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.2f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Streamed playback if supported",
			SettingsInternalName: "tts_streamed_playback",
			SettingsDescription:  "Play text-to-speech by streaming audio if it is supported by the TTS service.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Streamed playback chunk size",
			SettingsInternalName: "tts_streamed_chunk_size",
			SettingsDescription:  "Chunk size for streaming audio (If supported by the TTS service)",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(20, 2000)
				sliderState := widget.NewLabel(fmt.Sprintf("%.0f", sliderWidget.Min))
				sliderWidget.Step = 1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.0f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Streamed playback minimum play time",
			SettingsInternalName: "tts_streamed_min_play_time",
			SettingsDescription:  "Minimum play time before playback starts (in seconds)\nCan reduce gaps at the start of playback.",
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
			SettingsName:         "Route text-to-speech to secondary audio device",
			SettingsInternalName: "tts_use_secondary_playback",
			SettingsDescription:  "Play text-to-speech on a secondary audio device at the same time as the selected output device.\n(By default uses windows default audio device)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}
