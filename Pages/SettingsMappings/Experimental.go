package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
		{
			SettingsName:         "Noise Filter before recording trigger",
			SettingsInternalName: "denoise_audio_before_trigger",
			SettingsDescription:  "Noise Filter will be applied on audio before Volume + VAD trigger conditions are detected.\nThis can heavily influence audio quality since it increases processing time per chunk!",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Recognize speaker changes",
			SettingsInternalName: "speaker_diarization",
			SettingsDescription:  "Process speaker changes in conversation.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Recognize min. speaker length",
			SettingsInternalName: "min_speaker_length",
			SettingsDescription:  "Minimum length of a speaker in a conversation. (in seconds)",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0.0, 9.9)
				sliderState := widget.NewLabel(fmt.Sprintf("%.1f", sliderWidget.Min))
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Maximum speakers",
			SettingsInternalName: "max_speakers",
			SettingsDescription:  "The maximum number of speakers in a conversation.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(1, 5)
				sliderState := widget.NewLabel(fmt.Sprintf("%.0f", sliderWidget.Min))
				sliderWidget.Step = 1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.0f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
	},
}
