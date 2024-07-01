package SettingsMappings

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
)

var SpeechToTextSettingsMapping = SettingsMapping{
	Mappings: []SettingMapping{
		{
			SettingsName:         "Speech volume level",
			SettingsInternalName: "energy",
			SettingsDescription:  "Volume level at which the speech detection will trigger.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, EnergySliderMax)
				sliderState := widget.NewLabel(fmt.Sprintf("%.0f", sliderWidget.Min))
				sliderWidget.Step = 1
				sliderWidget.OnChanged = func(value float64) {
					if value >= sliderWidget.Max {
						sliderWidget.Max += 10
					}
					sliderState.SetText(fmt.Sprintf("%.0f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Voice Activity Confidence",
			SettingsInternalName: "vad_confidence_threshold",
			SettingsDescription:  "Voice Activity Detection (VAD) confidence threshold. Can be 0-1",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 1)
				sliderState := widget.NewLabel(fmt.Sprintf("%.2f", sliderWidget.Min))
				sliderWidget.Step = 0.01
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.2f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Speech pause detection",
			SettingsInternalName: "pause",
			SettingsDescription:  "Pause time in seconds after which the speech detection will stop and A.I. processing starts.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 5)
				sliderState := widget.NewLabel(fmt.Sprintf("%.1f", sliderWidget.Min))
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Phrase time limit",
			SettingsInternalName: "phrase_time_limit",
			SettingsDescription:  "Maximum time limit in seconds after which the audio processing starts.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 30)
				sliderState := widget.NewLabel(fmt.Sprintf("%.1f", sliderWidget.Min))
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Run each transcription in a separate thread",
			SettingsInternalName: "thread_per_transcription",
			SettingsDescription:  "If enabled, each transcription will run in a separate thread.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Noise Filter",
			SettingsInternalName: "denoise_audio",
			SettingsDescription:  "Requires a restart when switching the Noise Filter type.\nDisabling will stop applying it even without restart.",
			_widget: func() fyne.CanvasObject {
				settingWidget := CustomWidget.NewTextValueSelect("denoise_audio", []CustomWidget.TextValueOption{
					{Text: "", Value: ""},
					{Text: "Noise Reduce", Value: "noise_reduce"},
					{Text: "DeepFilterNet", Value: "deepfilter"},
				}, func(s CustomWidget.TextValueOption) {}, 0)
				return settingWidget
			},
		},
		{
			SettingsName:         "Cut silent audio parts",
			SettingsInternalName: "silence_cutting_enabled",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Normalize audio",
			SettingsInternalName: "normalize_enabled",
			SettingsDescription:  "",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Search beams",
			SettingsInternalName: "beam_size",
			SettingsDescription:  "Number of beams to search for the best result.\nCan be 1-5. (lower = faster)",
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
		{
			SettingsName:         "Temperature fallback",
			SettingsInternalName: "temperature_fallback",
			SettingsDescription:  "If enabled, the temperature will fallback the temperature on low confidence.\n(disable for faster processing)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Condition on previous text",
			SettingsInternalName: "condition_on_previous_text",
			SettingsDescription:  "Provides the previous text to the A.I. to improve the results.\n(disable for faster processing)\n(disabling can also reduce hallucinations)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Repetition penalty",
			SettingsInternalName: "repetition_penalty",
			SettingsDescription:  "penalize the score of previously generated tokens (set > 1 to penalize)",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 2)
				sliderState := widget.NewLabel(fmt.Sprintf("%.2f", sliderWidget.Min))
				sliderWidget.Step = 0.01
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.2f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Apply voice markers to audio",
			SettingsInternalName: "whisper_apply_voice_markers",
			SettingsDescription:  "Can reduce A.I. hallucinations.\nMight not work correctly with Speech Language set to \"Auto\".",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Real-time mode",
			SettingsInternalName: "realtime",
			SettingsDescription:  "Transcribes and processes speech while it is recorded.",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
		{
			SettingsName:         "Real-time frequency",
			SettingsInternalName: "realtime_frequency_time",
			SettingsDescription:  "How often the audio is processed in seconds.",
			_widget: func() fyne.CanvasObject {
				sliderWidget := widget.NewSlider(0, 20)
				sliderState := widget.NewLabel(fmt.Sprintf("%.1f", sliderWidget.Min))
				sliderWidget.Step = 0.1
				sliderWidget.OnChanged = func(value float64) {
					sliderState.SetText(fmt.Sprintf("%.1f", value))
				}
				return container.NewBorder(nil, nil, nil, sliderState, sliderWidget)
			},
		},
		{
			SettingsName:         "Search beams for real-time mode",
			SettingsInternalName: "realtime_whisper_beam_size",
			SettingsDescription:  "Number of beams to search for the best result.\nCan be 1-5. (lower = faster)",
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
		{
			SettingsName:         "Temperature fallback for real-time mode",
			SettingsInternalName: "realtime_temperature_fallback",
			SettingsDescription:  "If enabled, the temperature will fallback the temperature on low confidence.\n(disable for faster processing)",
			_widget: func() fyne.CanvasObject {
				return widget.NewCheck("", func(b bool) {})
			},
		},
	},
}
