package Pages

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Utilities/AudioAPI"
	"whispering-tiger-ui/Websocket/Messages"
)

type audioRoutesUpdateRequest struct {
	Operation        string                          `json:"operation,omitempty"`
	RequestID        string                          `json:"request_id"`
	Routes           []Settings.AdditionalAudioRoute `json:"routes,omitempty"`
	Route            *Settings.AdditionalAudioRoute  `json:"route,omitempty"`
	RouteID          string                          `json:"route_id,omitempty"`
	Enabled          *bool                           `json:"enabled,omitempty"`
	RoutePlugins     map[string][]string             `json:"route_plugins,omitempty"`
	MainAudioPlugins *[]string                       `json:"main_audio_plugins"`
}

const maxAdditionalAudioRoutes = 8

func cloneAudioRoutes(routes []Settings.AdditionalAudioRoute) []Settings.AdditionalAudioRoute {
	cloned := append([]Settings.AdditionalAudioRoute(nil), routes...)
	for index := range cloned {
		cloned[index].Plugins = append([]string(nil), routes[index].Plugins...)
	}
	return cloned
}

func clonePluginAllowlist(plugins *[]string) *[]string {
	if plugins == nil {
		return nil
	}
	cloned := append([]string(nil), (*plugins)...)
	return &cloned
}

func enabledAudioRoutePlugins() []string {
	names := make([]string, 0, len(Settings.Config.Plugins))
	for name, enabled := range Settings.Config.Plugins {
		// This compatibility shim is no longer an audio-processing plugin and
		// should not be offered as a native route destination.
		if name == "SecondaryProfilePlugin" || !enabled {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type audioRoutePluginOption struct {
	Name    string
	Enabled bool
}

func audioRoutePluginOptions(
	enabledPluginNames []string,
	routes []Settings.AdditionalAudioRoute,
	mainPlugins *[]string,
) []audioRoutePluginOption {
	names := pluginNameSet(enabledPluginNames)
	for _, route := range routes {
		for _, pluginName := range route.Plugins {
			if pluginName != "SecondaryProfilePlugin" {
				names[pluginName] = true
			}
		}
	}
	if mainPlugins != nil {
		for _, pluginName := range *mainPlugins {
			if pluginName != "SecondaryProfilePlugin" {
				names[pluginName] = true
			}
		}
	}

	pluginNames := selectedPluginNames(names)
	options := make([]audioRoutePluginOption, 0, len(pluginNames))
	for _, pluginName := range pluginNames {
		options = append(options, audioRoutePluginOption{
			Name:    pluginName,
			Enabled: Settings.Config.Plugins[pluginName],
		})
	}
	return options
}

func pluginNameSet(names []string) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = true
	}
	return result
}

func selectedPluginNames(selected map[string]bool) []string {
	names := make([]string, 0, len(selected))
	for name, enabled := range selected {
		if enabled {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func setPluginSelected(names []string, pluginName string, selected bool) []string {
	values := pluginNameSet(names)
	values[pluginName] = selected
	return selectedPluginNames(values)
}

func newPluginRouteCheck(label string, checked bool, changed func(bool)) *widget.Check {
	check := widget.NewCheck(label, nil)
	check.SetChecked(checked)
	check.OnChanged = changed
	return check
}

func routeSpeechLanguageOptions() []CustomWidget.TextValueOption {
	options := []CustomWidget.TextValueOption{{Text: lang.L("Autodetect"), Value: ""}}
	seen := map[string]bool{"": true}
	for _, language := range Messages.TranslateSettings.WhisperLanguages {
		if seen[language.Code] {
			continue
		}
		seen[language.Code] = true
		options = append(options, CustomWidget.TextValueOption{Text: language.Name, Value: language.Code})
	}
	return options
}

func routeTranslationLanguageOptions(includeAuto bool) []CustomWidget.TextValueOption {
	options := make([]CustomWidget.TextValueOption, 0, len(Messages.InstalledLanguages.Languages)+1)
	seen := make(map[string]bool)
	if includeAuto {
		options = append(options, CustomWidget.TextValueOption{Text: lang.L("Autodetect"), Value: "auto"})
		seen["auto"] = true
	}
	for _, language := range Messages.InstalledLanguages.Languages {
		code := strings.TrimSpace(language.Code)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		options = append(options, CustomWidget.TextValueOption{Text: language.Name, Value: code})
	}
	return options
}

func matchRouteLanguageOption(options []CustomWidget.TextValueOption, input string) (CustomWidget.TextValueOption, bool) {
	value := strings.TrimSpace(input)
	for _, option := range options {
		if strings.EqualFold(value, option.Text) || strings.EqualFold(value, option.Value) {
			return option, true
		}
	}
	for _, option := range options {
		if strings.HasPrefix(strings.ToLower(option.Text), strings.ToLower(value)) {
			return option, true
		}
	}
	for _, option := range options {
		if strings.Contains(strings.ToLower(option.Text), strings.ToLower(value)) {
			return option, true
		}
	}
	return CustomWidget.TextValueOption{}, false
}

func newRouteLanguageCompletion(
	options []CustomWidget.TextValueOption,
	selected string,
	changed func(string),
) *CustomWidget.CompletionEntry {
	entry := CustomWidget.NewCompletionEntry(nil)
	entry.SetValueOptions(options)
	entry.ResetOptionsFilter()
	currentValue := selected
	entry.SetSelected(selected)
	if entry.Text == "" && selected != "" {
		entry.Text = selected
	}
	entry.OnSubmitted = func(input string) {
		option, ok := matchRouteLanguageOption(options, input)
		if !ok {
			entry.SetSelected(currentValue)
			return
		}
		currentValue = option.Value
		entry.SetSelected(option.Value)
		if changed != nil {
			changed(option.Value)
		}
	}
	return entry
}

func selectOrAddRouteOption(selection *CustomWidget.TextValueSelect, value string) {
	selection.SetSelected(value)
	if selection.GetSelected() != nil || value == "" {
		return
	}
	option := CustomWidget.TextValueOption{Text: value, Value: value}
	selection.SetValueOptions(append(selection.Options, option))
	selection.SetSelected(value)
}

func selectConfiguredRouteApplication(route *Settings.AdditionalAudioRoute, selection *CustomWidget.TextValueSelect) {
	if selection == nil || strings.TrimSpace(route.Audio_input_process) == "" {
		return
	}
	executable := strings.TrimSpace(route.Audio_input_process)
	exactValue := Utilities.FormatAudioProcessOptionValue(uint32(route.Audio_input_process_id), executable)
	for _, option := range selection.Options {
		pid, optionExecutable, ok := Utilities.ParseAudioProcessOptionValue(option.Value)
		if ok && strings.EqualFold(optionExecutable, executable) && int(pid) == route.Audio_input_process_id {
			selection.SetSelected(option.Value)
			return
		}
	}
	label := strings.TrimSpace(route.Audio_input_device)
	if label == "" {
		label = fmt.Sprintf("%s (PID %d)", executable, route.Audio_input_process_id)
	}
	selection.SetValueOptions(append(selection.Options, CustomWidget.TextValueOption{Text: label, Value: exactValue}))
	selection.SetSelected(exactValue)
}

func createRouteAudioControls(
	route *Settings.AdditionalAudioRoute,
	sourceChanged func(),
) (*widget.Select, fyne.CanvasObject) {
	backendNames := make([]string, 0, len(AudioAPI.AudioBackends))
	for _, backend := range AudioAPI.AudioBackends {
		backendNames = append(backendNames, backend.Name)
	}
	if route.Audio_api == "" {
		route.Audio_api = Settings.Config.Audio_api
	}
	backend := AudioAPI.GetAudioBackendByName(route.Audio_api)

	inputSelection := CustomWidget.NewTextValueSelect(
		"additional_audio_input_"+route.ID,
		liveAudioInputOptions(backend.Backend),
		nil,
		-1,
	)
	applicationSelection := CustomWidget.NewTextValueSelect(
		"additional_audio_application_"+route.ID,
		applicationCaptureOptions(),
		nil,
		-1,
	)
	applicationSelection.Hide()

	if route.Audio_input_process != "" {
		inputSelection.SetSelected(Utilities.AudioApplicationOptionValue)
		applicationSelection.Show()
		selectConfiguredRouteApplication(route, applicationSelection)
	} else {
		selected := false
		for _, option := range inputSelection.Options {
			if strings.EqualFold(strings.TrimSpace(option.Text), strings.TrimSpace(route.Audio_input_device)) {
				inputSelection.SetSelected(option.Value)
				selected = true
				break
			}
		}
		if !selected {
			inputSelection.SetSelected("-1")
		}
	}

	inputSelection.OnChanged = func(option CustomWidget.TextValueOption) {
		if Utilities.IsAudioApplicationOptionValue(option.Value) {
			applicationSelection.Show()
			if selected := applicationSelection.GetSelected(); selected != nil {
				pid, executable, ok := Utilities.ParseAudioProcessOptionValue(selected.Value)
				if ok {
					route.Audio_input_device = selected.Text
					route.Audio_input_process = executable
					route.Audio_input_process_id = int(pid)
					route.Device_index = -1
				}
			}
			if sourceChanged != nil {
				sourceChanged()
			}
			return
		}
		applicationSelection.Hide()
		route.Audio_input_device = option.Text
		route.Audio_input_process = ""
		route.Audio_input_process_id = 0
		route.Device_index = nil
		if sourceChanged != nil {
			sourceChanged()
		}
	}
	applicationSelection.OnChanged = func(option CustomWidget.TextValueOption) {
		pid, executable, ok := Utilities.ParseAudioProcessOptionValue(option.Value)
		if !ok {
			return
		}
		route.Audio_input_device = option.Text
		route.Audio_input_process = executable
		route.Audio_input_process_id = int(pid)
		route.Device_index = -1
		if sourceChanged != nil {
			sourceChanged()
		}
	}
	inputSelection.BeforeTapped = func() {
		currentBackend := AudioAPI.GetAudioBackendByName(route.Audio_api)
		inputSelection.SetValueOptions(liveAudioInputOptions(currentBackend.Backend))
	}
	applicationSelection.BeforeTapped = func() {
		refreshApplicationCaptureOptions(applicationSelection)
	}

	apiSelection := widget.NewSelect(backendNames, nil)
	apiSelection.SetSelected(backend.Name)
	apiSelection.OnChanged = func(name string) {
		selectedBackend := AudioAPI.GetAudioBackendByName(name)
		route.Audio_api = selectedBackend.Name
		route.Audio_input_device = "Default"
		route.Audio_input_process = ""
		route.Audio_input_process_id = 0
		route.Device_index = -1
		applicationSelection.Hide()
		inputSelection.SetValueOptions(liveAudioInputOptions(selectedBackend.Backend))
		inputSelection.SetSelected("-1")
		if sourceChanged != nil {
			sourceChanged()
		}
	}

	return apiSelection, container.NewVBox(inputSelection, applicationSelection)
}

func newRouteSlider(
	value float64,
	minimum float64,
	maximum float64,
	step float64,
	valueFormat string,
	extendMaximum bool,
	changed func(float64),
) (*widget.Slider, fyne.CanvasObject) {
	return newRouteSliderWithFormatter(
		value, minimum, maximum, step,
		func(updated float64) string { return fmt.Sprintf(valueFormat, updated) },
		extendMaximum, changed,
	)
}

func newRouteSliderWithFormatter(
	value float64,
	minimum float64,
	maximum float64,
	step float64,
	formatValue func(float64) string,
	extendMaximum bool,
	changed func(float64),
) (*widget.Slider, fyne.CanvasObject) {
	if extendMaximum && value >= maximum {
		maximum = value + 10
	} else if value > maximum {
		// Preserve valid values from older profiles even when the normal UI
		// range is intentionally more compact.
		maximum = value
	}

	slider := widget.NewSlider(minimum, maximum)
	slider.Step = step
	slider.SetValue(value)
	state := widget.NewLabel(formatValue(slider.Value))
	slider.OnChanged = func(updated float64) {
		if extendMaximum && updated >= slider.Max {
			slider.Max += 10
		}
		state.SetText(formatValue(updated))
		if changed != nil {
			changed(updated)
		}
	}
	return slider, container.NewBorder(nil, nil, nil, state, slider)
}

func createAudioRouteDetails(
	route *Settings.AdditionalAudioRoute,
	preview *routeAudioInputPreview,
) fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetText(route.Name)
	name.OnChanged = func(value string) { route.Name = value }

	sttEnabled := widget.NewCheck("", func(value bool) { route.Stt_enabled = value })
	sttEnabled.SetChecked(route.Stt_enabled)
	realtime := widget.NewCheck("", nil)
	realtime.SetChecked(route.Realtime)
	translate := widget.NewCheck("", func(value bool) { route.Txt_translate = value })
	translate.SetChecked(route.Txt_translate)
	romaji := widget.NewCheck("", func(value bool) { route.Txt_romaji = value })
	romaji.SetChecked(route.Txt_romaji)
	websocketEnabled := widget.NewCheck("", func(value bool) { route.Websocket_enabled = value })
	websocketEnabled.SetChecked(route.Websocket_enabled)
	oscEnabled := widget.NewCheck("", nil)
	oscEnabled.SetChecked(route.Osc_enabled)
	oscTypingIndicator := widget.NewCheck("", func(value bool) {
		route.Osc_typing_indicator = value
	})
	oscTypingIndicator.SetChecked(route.Osc_typing_indicator)
	oscChatNotification := widget.NewCheck("", func(value bool) {
		route.Osc_chat_notification = value
	})
	oscChatNotification.SetChecked(route.Osc_chat_notification)
	oscChatPrefix := widget.NewEntry()
	oscChatPrefix.SetText(route.Osc_chat_prefix)
	oscChatPrefix.OnChanged = func(value string) { route.Osc_chat_prefix = value }

	apiSelection, inputSelection := createRouteAudioControls(route, func() {
		preview.Update(*route)
	})
	inputAndLevel := container.NewVBox(inputSelection, preview.CanvasObject())

	task := CustomWidget.NewTextValueSelect(
		"additional_audio_task_"+route.ID,
		[]CustomWidget.TextValueOption{
			{Text: lang.L("transcribe"), Value: "transcribe"},
			{Text: lang.L("translate (to English)"), Value: "translate"},
		},
		func(option CustomWidget.TextValueOption) { route.Whisper_task = option.Value },
		-1,
	)
	selectOrAddRouteOption(task, route.Whisper_task)

	speechLanguage := newRouteLanguageCompletion(
		routeSpeechLanguageOptions(),
		route.Current_language,
		func(value string) { route.Current_language = value },
	)

	sourceLanguage := newRouteLanguageCompletion(
		routeTranslationLanguageOptions(true),
		route.Src_lang,
		func(value string) { route.Src_lang = value },
	)
	targetLanguage := newRouteLanguageCompletion(
		routeTranslationLanguageOptions(false),
		route.Trg_lang,
		func(value string) { route.Trg_lang = value },
	)

	_, energy := newRouteSliderWithFormatter(
		float64(route.Energy), 0, EnergySliderMax, 1,
		func(value float64) string {
			return Utilities.FormatEnergyThresholdDBFS(value, lang.L("Disabled"))
		}, true,
		func(value float64) {
			route.Energy = int(value)
			preview.SetThreshold(route.Energy)
		},
	)
	_, vadConfidence := newRouteSlider(
		route.Vad_confidence_threshold, 0, 1, 0.01, "%.2f", false,
		func(value float64) { route.Vad_confidence_threshold = value },
	)
	_, phraseLimit := newRouteSlider(
		route.Phrase_time_limit, 0, 30, 0.1, "%.1f", false,
		func(value float64) { route.Phrase_time_limit = value },
	)
	_, pause := newRouteSlider(
		route.Pause, 0, 5, 0.1, "%.1f", false,
		func(value float64) { route.Pause = value },
	)
	realtimeFrequencySlider, realtimeFrequency := newRouteSlider(
		route.Realtime_frequency_time, 0, 20, 0.1, "%.1f", false,
		func(value float64) { route.Realtime_frequency_time = value },
	)
	silenceCutting := widget.NewCheck("", func(value bool) {
		route.Silence_cutting_enabled = value
	})
	silenceCutting.SetChecked(route.Silence_cutting_enabled)
	noiseFilter := CustomWidget.NewTextValueSelect(
		"additional_audio_denoise_"+route.ID,
		[]CustomWidget.TextValueOption{
			{Text: lang.L("Disabled"), Value: ""},
			{Text: lang.L("Noise Reduce"), Value: "noise_reduce"},
			{Text: lang.L("DeepFilterNet"), Value: "deepfilter"},
		},
		nil,
		-1,
	)
	selectOrAddRouteOption(noiseFilter, route.Denoise_audio)
	denoiseStrengthSlider, denoiseStrength := newRouteSlider(
		route.Denoise_strength, 0, 1, 0.01, "%.2f", false,
		func(value float64) { route.Denoise_strength = value },
	)
	smartTurn := widget.NewCheck("", nil)
	smartTurn.SetChecked(route.Vad_smart_turn_enabled)
	smartTurnMinLengthSlider, smartTurnMinLength := newRouteSlider(
		route.Vad_smart_turn_min_length, 0, 8, 0.1, "%.1f", false,
		func(value float64) { route.Vad_smart_turn_min_length = value },
	)
	smartTurnProbabilitySlider, smartTurnProbability := newRouteSlider(
		route.Vad_smart_turn_probability_threshold, 0, 1, 0.01, "%.2f", false,
		func(value float64) { route.Vad_smart_turn_probability_threshold = value },
	)
	smartTurnPauseSlider, smartTurnPause := newRouteSlider(
		route.Vad_smart_turn_pause_length, 0, 5, 0.01, "%.2f", false,
		func(value float64) { route.Vad_smart_turn_pause_length = value },
	)

	updateRealtimeFrequencyState := func() {
		if route.Realtime {
			realtimeFrequencySlider.Enable()
		} else {
			realtimeFrequencySlider.Disable()
		}
	}
	realtime.OnChanged = func(value bool) {
		route.Realtime = value
		updateRealtimeFrequencyState()
	}
	updateRealtimeFrequencyState()

	updateDenoiseState := func() {
		if route.Denoise_audio == "noise_reduce" {
			denoiseStrengthSlider.Enable()
		} else {
			denoiseStrengthSlider.Disable()
		}
	}
	noiseFilter.OnChanged = func(option CustomWidget.TextValueOption) {
		route.Denoise_audio = option.Value
		updateDenoiseState()
	}
	updateDenoiseState()

	updateSmartTurnState := func() {
		if route.Vad_smart_turn_enabled {
			smartTurnMinLengthSlider.Enable()
			smartTurnProbabilitySlider.Enable()
			smartTurnPauseSlider.Enable()
		} else {
			smartTurnMinLengthSlider.Disable()
			smartTurnProbabilitySlider.Disable()
			smartTurnPauseSlider.Disable()
		}
	}
	smartTurn.OnChanged = func(value bool) {
		route.Vad_smart_turn_enabled = value
		updateSmartTurnState()
	}
	updateSmartTurnState()

	updateOSCState := func() {
		if route.Osc_enabled {
			oscTypingIndicator.Enable()
			oscChatNotification.Enable()
			oscChatPrefix.Enable()
		} else {
			oscTypingIndicator.Disable()
			oscChatNotification.Disable()
			oscChatPrefix.Disable()
		}
	}
	oscEnabled.OnChanged = func(value bool) {
		route.Osc_enabled = value
		updateOSCState()
	}
	updateOSCState()

	sourceForm := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(lang.L("Audio Source Name")), name,
		widget.NewLabel(lang.L("Audio API")), apiSelection,
		widget.NewLabel(lang.L("Audio Input (mic)")), inputAndLevel,
		widget.NewLabel(lang.L("energy.Name")), energy,
	)
	recognitionForm := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(lang.L("Speech-to-Text Enabled")), sttEnabled,
		widget.NewLabel(lang.L("Task")), task,
		widget.NewLabel(lang.L("Speech Language")), speechLanguage,
		widget.NewLabel(lang.L("Realtime")), realtime,
	)
	translationOutputForm := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(lang.L("Text-Translate")), translate,
		widget.NewLabel(lang.L("Source Language")), sourceLanguage,
		widget.NewLabel(lang.L("Target Language")), targetLanguage,
		widget.NewLabel(lang.L("txt_romaji.Name")), romaji,
		widget.NewLabel(lang.L("Show Results in Whispering Tiger")), websocketEnabled,
		widget.NewLabel(lang.L("Automatic OSC (VRChat)")), oscEnabled,
		widget.NewLabel(lang.L("osc_chat_prefix.Name")), oscChatPrefix,
		widget.NewLabel(lang.L("VRChat Typing Indicator")), oscTypingIndicator,
		widget.NewLabel(lang.L("VRChat Notification Sound")), oscChatNotification,
	)
	processingForm := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(lang.L("vad_confidence_threshold.Name")), vadConfidence,
		widget.NewLabel(lang.L("phrase_time_limit.Name")), phraseLimit,
		widget.NewLabel(lang.L("pause.Name")), pause,
		widget.NewLabel(lang.L("realtime_frequency_time.Name")), realtimeFrequency,
		widget.NewLabel(lang.L("silence_cutting_enabled.Name")), silenceCutting,
		widget.NewLabel(lang.L("denoise_audio.Name")), noiseFilter,
		widget.NewLabel(lang.L("denoise_strength.Name")), denoiseStrength,
		widget.NewLabel(lang.L("Smart Turn Detection")), smartTurn,
		widget.NewLabel(lang.L("Turn Detection Minimum Length")), smartTurnMinLength,
		widget.NewLabel(lang.L("Turn Probability Threshold")), smartTurnProbability,
		widget.NewLabel(lang.L("Turn Pause Length")), smartTurnPause,
	)
	details := widget.NewAccordion(
		widget.NewAccordionItem(lang.L("Audio Source"), sourceForm),
		widget.NewAccordionItem(lang.L("Speech-to-Text"), recognitionForm),
		widget.NewAccordionItem(lang.L("Text-Translate"), translationOutputForm),
		widget.NewAccordionItem(lang.L("Audio Processing"), processingForm),
	)
	details.Open(0)
	return details
}

func createPluginRouting(
	pluginOptions []audioRoutePluginOption,
	routes []Settings.AdditionalAudioRoute,
	mainPlugins *[]string,
	setMain func(string, bool),
	setRoute func(int, string, bool),
) fyne.CanvasObject {
	explanation := widget.NewLabel(lang.L("Plugin Routing Chooses Audio Sources for Enabled Plugins"))
	explanation.Wrapping = fyne.TextWrapWord
	content := container.NewVBox(explanation)
	if len(pluginOptions) == 0 {
		content.Add(widget.NewLabel(lang.L("No Plugins found. Download Plugins using the button below.")))
		return content
	}

	mainSelected := map[string]bool{}
	if mainPlugins == nil {
		for _, option := range pluginOptions {
			mainSelected[option.Name] = true
		}
	} else {
		mainSelected = pluginNameSet(*mainPlugins)
	}

	for _, configuredPlugin := range pluginOptions {
		plugin := configuredPlugin
		pluginName := plugin.Name
		targets := make([]fyne.CanvasObject, 0, len(routes)+1)
		mainCheck := newPluginRouteCheck(
			lang.L("Main Microphone"),
			mainSelected[pluginName],
			func(selected bool) { setMain(pluginName, selected) },
		)
		targets = append(targets, mainCheck)
		for routeIndex := range routes {
			index := routeIndex
			routeName := strings.TrimSpace(routes[index].Name)
			if routeName == "" {
				routeName = fmt.Sprintf("%s %d", lang.L("Audio Source"), index+1)
			}
			routeCheck := newPluginRouteCheck(
				routeName,
				pluginNameSet(routes[index].Plugins)[pluginName],
				func(selected bool) { setRoute(index, pluginName, selected) },
			)
			targets = append(targets, routeCheck)
		}
		pluginTitle := pluginName
		if !plugin.Enabled {
			pluginTitle += " (" + lang.L("Disabled") + ")"
		}
		targetRow := container.NewHBox(targets...)
		content.Add(widget.NewCard(
			pluginTitle,
			"",
			container.NewHScroll(targetRow),
		))
	}
	return content
}

func defaultAdditionalAudioRoute(index int) Settings.AdditionalAudioRoute {
	return Settings.AdditionalAudioRoute{
		ID:                                   fmt.Sprintf("audio-route-%d", time.Now().UnixNano()),
		Name:                                 fmt.Sprintf("%s %d", lang.L("Audio Source"), index),
		Enabled:                              true,
		Audio_api:                            Settings.Config.Audio_api,
		Audio_input_device:                   "Default",
		Device_index:                         -1,
		Stt_enabled:                          true,
		Current_language:                     Settings.Config.Current_language,
		Whisper_task:                         Settings.Config.Whisper_task,
		Energy:                               Settings.Config.Energy,
		Vad_confidence_threshold:             Settings.Config.Vad_confidence_threshold,
		Phrase_time_limit:                    Settings.Config.Phrase_time_limit,
		Pause:                                Settings.Config.Pause,
		Realtime:                             false,
		Realtime_frequency_time:              Settings.Config.Realtime_frequency_time,
		Silence_cutting_enabled:              Settings.Config.Silence_cutting_enabled,
		Denoise_audio:                        Settings.Config.Denoise_audio,
		Denoise_strength:                     Settings.Config.Denoise_strength,
		Vad_smart_turn_enabled:               Settings.Config.Vad_smart_turn_enabled,
		Vad_smart_turn_min_length:            Settings.Config.Vad_smart_turn_min_length,
		Vad_smart_turn_probability_threshold: Settings.Config.Vad_smart_turn_probability_threshold,
		Vad_smart_turn_pause_length:          Settings.Config.Vad_smart_turn_pause_length,
		Txt_translate:                        false,
		Src_lang:                             Settings.Config.Src_lang,
		Trg_lang:                             Settings.Config.Trg_lang,
		Txt_romaji:                           false,
		Websocket_enabled:                    true,
		Osc_enabled:                          false,
		Osc_typing_indicator:                 false,
		Osc_chat_notification:                false,
		Osc_chat_prefix:                      Settings.Config.Osc_chat_prefix,
		Plugins:                              []string{},
	}
}

type audioRouteResultCallback func(success bool, errorMessage string)

var (
	audioRouteResultMutex     sync.Mutex
	audioRouteResultCallbacks = map[string]audioRouteResultCallback{}
)

func installAudioRouteResultHandler() {
	Fields.AudioRoutesUpdateResult = func(requestID string, success bool, errorMessage string) {
		audioRouteResultMutex.Lock()
		callback := audioRouteResultCallbacks[requestID]
		delete(audioRouteResultCallbacks, requestID)
		audioRouteResultMutex.Unlock()
		if callback != nil {
			callback(success, errorMessage)
		}
	}
}

func sendAudioRouteRequest(request audioRoutesUpdateRequest, callback audioRouteResultCallback) {
	request.RequestID = fmt.Sprintf("%d", time.Now().UnixNano())
	audioRouteResultMutex.Lock()
	audioRouteResultCallbacks[request.RequestID] = callback
	audioRouteResultMutex.Unlock()
	SendMessageChannel.SendMessageStruct{
		Type:  "audio_routes_update",
		Value: request,
	}.SendMessage()
}

func currentAudioRouteWindow() fyne.Window {
	if app := fyne.CurrentApp(); app != nil {
		windows := app.Driver().AllWindows()
		if len(windows) > 0 {
			return windows[0]
		}
	}
	return nil
}

func showAudioRouteError(parent fyne.Window, errorMessage string) {
	if parent == nil {
		return
	}
	if strings.TrimSpace(errorMessage) == "" {
		errorMessage = lang.L("Error")
	}
	dialog.ShowError(errors.New(errorMessage), parent)
}

func resizeAudioRouteDialog(modal *dialog.CustomDialog, parent fyne.Window, maximum fyne.Size) {
	if modal == nil || parent == nil {
		return
	}
	size := parent.Canvas().Size()
	size.Width -= 80
	size.Height -= 80
	if size.Width > maximum.Width {
		size.Width = maximum.Width
	}
	if size.Height > maximum.Height {
		size.Height = maximum.Height
	}
	if size.Width < 480 {
		size.Width = 480
	}
	if size.Height < 360 {
		size.Height = 360
	}
	modal.Resize(size)
}

func showAudioRouteDialog(route Settings.AdditionalAudioRoute, isNew bool, refresh func()) {
	parent := currentAudioRouteWindow()
	if parent == nil {
		return
	}
	draft := route
	draft.Plugins = append([]string(nil), route.Plugins...)
	preview := newRouteAudioInputPreview(draft.Energy)
	details := createAudioRouteDetails(&draft, preview)
	detailsScroll := container.NewVScroll(details)
	detailsScroll.SetMinSize(fyne.NewSize(520, 420))

	progress := widget.NewProgressBarInfinite()
	progress.Stop()
	progress.Hide()
	saveButton := widget.NewButtonWithIcon(lang.L("Save"), theme.ConfirmIcon(), nil)
	saveButton.Importance = widget.HighImportance
	cancelButton := widget.NewButton(lang.L("Cancel"), nil)
	removeButton := widget.NewButtonWithIcon(lang.L("Remove Audio Source"), theme.DeleteIcon(), nil)
	removeButton.Importance = widget.DangerImportance
	if isNew {
		removeButton.Hide()
	}

	var modal *dialog.CustomDialog
	setBusy := func(busy bool) {
		if busy {
			saveButton.Disable()
			cancelButton.Disable()
			removeButton.Disable()
			progress.Show()
			progress.Start()
			return
		}
		saveButton.Enable()
		cancelButton.Enable()
		removeButton.Enable()
		progress.Stop()
		progress.Hide()
	}
	cancelButton.OnTapped = func() { modal.Hide() }
	saveButton.OnTapped = func() {
		if strings.TrimSpace(draft.Name) == "" {
			showAudioRouteError(parent, lang.L("Audio Source Name Is Required"))
			return
		}
		setBusy(true)
		requestDraft := draft
		request := audioRoutesUpdateRequest{
			Operation: "upsert",
			Route:     &requestDraft,
		}
		sendAudioRouteRequest(request, func(success bool, errorMessage string) {
			if success {
				modal.Hide()
				refresh()
				return
			}
			setBusy(false)
			showAudioRouteError(parent, errorMessage)
		})
	}
	removeButton.OnTapped = func() {
		confirm := dialog.NewConfirm(
			lang.L("Remove Audio Source"),
			strings.TrimSpace(draft.Name)+"?",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				setBusy(true)
				sendAudioRouteRequest(audioRoutesUpdateRequest{
					Operation: "delete",
					RouteID:   draft.ID,
				}, func(success bool, errorMessage string) {
					if success {
						modal.Hide()
						refresh()
						return
					}
					setBusy(false)
					showAudioRouteError(parent, errorMessage)
				})
			},
			parent,
		)
		confirm.SetConfirmImportance(widget.DangerImportance)
		confirm.Show()
	}

	footer := container.NewBorder(
		nil,
		nil,
		removeButton,
		container.NewHBox(cancelButton, saveButton),
		progress,
	)
	content := container.NewBorder(
		nil,
		container.NewVBox(widget.NewSeparator(), footer),
		nil,
		nil,
		detailsScroll,
	)
	title := lang.L("Add Audio Source")
	if !isNew {
		title = lang.L("Audio Source") + ": " + route.Name
	}
	modal = dialog.NewCustomWithoutButtons(title, content, parent)
	modal.SetOnClosed(preview.Stop)
	resizeAudioRouteDialog(modal, parent, fyne.NewSize(760, 720))
	modal.Show()
	preview.Update(draft)
}

func createAudioRoutePluginRoutingEditor(
	parent fyne.Window,
	onSaved func(),
) (*fyne.Container, func()) {
	editor := container.NewStack()
	var render func()
	render = func() {
		draftRoutes := cloneAudioRoutes(Settings.Config.Additional_audio_routes)
		draftMainPlugins := clonePluginAllowlist(Settings.Config.Main_audio_plugins)
		mainSelectionExplicit := draftMainPlugins != nil
		pluginNames := enabledAudioRoutePlugins()
		pluginOptions := audioRoutePluginOptions(
			pluginNames, draftRoutes, draftMainPlugins,
		)
		allPluginNames := make([]string, 0, len(pluginOptions))
		for _, option := range pluginOptions {
			allPluginNames = append(allPluginNames, option.Name)
		}

		routing := createPluginRouting(
			pluginOptions,
			draftRoutes,
			draftMainPlugins,
			func(pluginName string, selected bool) {
				mainSelectionExplicit = true
				current := append([]string(nil), allPluginNames...)
				if draftMainPlugins != nil {
					current = append([]string(nil), (*draftMainPlugins)...)
				}
				current = setPluginSelected(current, pluginName, selected)
				draftMainPlugins = &current
			},
			func(routeIndex int, pluginName string, selected bool) {
				draftRoutes[routeIndex].Plugins = setPluginSelected(
					draftRoutes[routeIndex].Plugins, pluginName, selected,
				)
			},
		)
		routingScroll := container.NewVScroll(routing)
		routingScroll.SetMinSize(fyne.NewSize(600, 420))

		progress := widget.NewProgressBarInfinite()
		progress.Stop()
		progress.Hide()
		saveButton := widget.NewButtonWithIcon(lang.L("Save"), theme.ConfirmIcon(), nil)
		saveButton.Importance = widget.HighImportance
		setBusy := func(busy bool) {
			if busy {
				saveButton.Disable()
				progress.Show()
				progress.Start()
				return
			}
			saveButton.Enable()
			progress.Stop()
			progress.Hide()
		}
		saveButton.OnTapped = func() {
			setBusy(true)
			routePlugins := make(map[string][]string, len(draftRoutes))
			for _, route := range draftRoutes {
				routePlugins[route.ID] = append([]string(nil), route.Plugins...)
			}
			requestMainPlugins := draftMainPlugins
			if !mainSelectionExplicit {
				requestMainPlugins = nil
			}
			sendAudioRouteRequest(audioRoutesUpdateRequest{
				Operation:        "plugin_routing",
				RoutePlugins:     routePlugins,
				MainAudioPlugins: clonePluginAllowlist(requestMainPlugins),
			}, func(success bool, errorMessage string) {
				if success {
					render()
					if onSaved != nil {
						onSaved()
					}
					return
				}
				setBusy(false)
				showAudioRouteError(parent, errorMessage)
			})
		}
		footer := container.NewBorder(
			nil, nil, nil, container.NewHBox(saveButton), progress,
		)
		content := container.NewBorder(
			nil,
			container.NewVBox(widget.NewSeparator(), footer),
			nil,
			nil,
			routingScroll,
		)
		editor.RemoveAll()
		editor.Add(content)
		editor.Refresh()
	}
	render()
	return editor, render
}

func audioRouteInputSummary(route Settings.AdditionalAudioRoute) string {
	input := strings.TrimSpace(route.Audio_input_device)
	if input == "" {
		input = "Default"
	}
	api := strings.TrimSpace(route.Audio_api)
	if api == "" {
		return input
	}
	return api + " - " + input
}

func audioRoutePanelTitle(routes []Settings.AdditionalAudioRoute) string {
	activeCount := 0
	for _, route := range routes {
		if route.Enabled {
			activeCount++
		}
	}
	return fmt.Sprintf(
		"%s (%d/%d)", lang.L("Additional Audio Sources"), activeCount, len(routes),
	)
}

func audioRouteEnabledUpdateRequest(routeID string, enabled bool) audioRoutesUpdateRequest {
	requestedEnabled := enabled
	return audioRoutesUpdateRequest{
		Operation: "set_enabled",
		RouteID:   routeID,
		Enabled:   &requestedEnabled,
	}
}

func newAudioRouteEnabledCheck(name string, enabled bool) *widget.Check {
	check := widget.NewCheck(name, nil)
	check.SetChecked(enabled)
	return check
}

func renderAdditionalAudioRouteRows(
	routesBox *fyne.Container,
	routes []Settings.AdditionalAudioRoute,
	refresh func(),
) {
	routesBox.RemoveAll()
	if len(routes) == 0 {
		message := widget.NewLabel(lang.L("No Additional Audio Sources Configured"))
		message.Alignment = fyne.TextAlignCenter
		routesBox.Add(message)
		routesBox.Refresh()
		return
	}

	for routeIndex, configuredRoute := range routes {
		route := configuredRoute
		enabled := newAudioRouteEnabledCheck(route.Name, route.Enabled)
		summary := widget.NewLabel(audioRouteInputSummary(route))
		summary.Truncation = fyne.TextTruncateEllipsis
		status := widget.NewLabel(lang.L("Disabled"))
		if route.Enabled {
			status.SetText(lang.L("Enabled"))
		}
		progress := widget.NewActivity()
		progress.Stop()
		progress.Hide()
		editButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
			showAudioRouteDialog(route, false, refresh)
		})
		editButton.Importance = widget.LowImportance
		enabledCell := container.New(
			layout.NewGridWrapLayout(fyne.NewSize(190, enabled.MinSize().Height)),
			enabled,
		)

		enabled.OnChanged = func(selected bool) {
			enabled.Disable()
			editButton.Disable()
			status.Hide()
			progress.Show()
			progress.Start()
			sendAudioRouteRequest(audioRouteEnabledUpdateRequest(
				route.ID, selected,
			), func(success bool, errorMessage string) {
				if !success {
					showAudioRouteError(currentAudioRouteWindow(), errorMessage)
				}
				refresh()
			})
		}

		row := container.NewBorder(
			nil,
			nil,
			enabledCell,
			container.NewHBox(status, progress, editButton),
			summary,
		)
		routesBox.Add(row)
		if routeIndex < len(routes)-1 {
			routesBox.Add(widget.NewSeparator())
		}
	}
	routesBox.Refresh()
}

func newAdditionalAudioRoutesManagerTabs(
	sources fyne.CanvasObject,
	routing fyne.CanvasObject,
) *container.AppTabs {
	return container.NewAppTabs(
		container.NewTabItem(lang.L("Additional Audio Sources"), sources),
		container.NewTabItem(lang.L("Plugin Routing"), routing),
	)
}

func showAdditionalAudioRoutesManagerDialog(
	refreshSummary func(),
	onClosed func(),
) func() {
	parent := currentAudioRouteWindow()
	if parent == nil {
		return nil
	}
	routesBox := container.NewVBox()
	routesScroll := container.NewVScroll(routesBox)
	addButton := widget.NewButtonWithIcon(
		lang.L("Add Audio Source"), theme.ContentAddIcon(), nil,
	)
	sourcesContent := container.NewBorder(
		nil,
		container.NewHBox(addButton),
		nil,
		nil,
		routesScroll,
	)
	routingEditor, renderPluginRouting := createAudioRoutePluginRoutingEditor(
		parent,
		refreshSummary,
	)
	tabs := newAdditionalAudioRoutesManagerTabs(sourcesContent, routingEditor)
	routingTab := tabs.Items[1]
	tabs.OnSelected = func(selected *container.TabItem) {
		if selected == routingTab {
			renderPluginRouting()
		}
	}

	var render func()
	refreshAll := func() {
		render()
		if refreshSummary != nil {
			refreshSummary()
		}
	}
	render = func() {
		routes := cloneAudioRoutes(Settings.Config.Additional_audio_routes)
		renderAdditionalAudioRouteRows(routesBox, routes, refreshAll)
		if len(routes) >= maxAdditionalAudioRoutes {
			addButton.Disable()
		} else {
			addButton.Enable()
		}
		height := routesBox.MinSize().Height + theme.Padding()*2
		if height < 80 {
			height = 80
		}
		if height > 340 {
			height = 340
		}
		routesScroll.SetMinSize(fyne.NewSize(640, height))
		if tabs.Selected() == routingTab {
			renderPluginRouting()
		}
		sourcesContent.Refresh()
		tabs.Refresh()
	}

	addButton.OnTapped = func() {
		if len(Settings.Config.Additional_audio_routes) >= maxAdditionalAudioRoutes {
			return
		}
		showAudioRouteDialog(
			defaultAdditionalAudioRoute(len(Settings.Config.Additional_audio_routes)+1),
			true,
			refreshAll,
		)
	}
	render()
	manager := dialog.NewCustom(
		lang.L("Additional Audio Sources"), lang.L("Close"), tabs, parent,
	)
	manager.SetOnClosed(onClosed)
	manager.Show()
	resizeAudioRouteDialog(manager, parent, fyne.NewSize(840, 650))
	return render
}

func createAdditionalAudioRoutesPanel() fyne.CanvasObject {
	installAudioRouteResultHandler()
	button := widget.NewButtonWithIcon("", theme.SettingsIcon(), nil)
	button.Alignment = widget.ButtonAlignLeading
	button.Importance = widget.LowImportance
	var managerRefresh func()
	refreshSummary := func() {
		button.SetText(audioRoutePanelTitle(Settings.Config.Additional_audio_routes))
	}
	refresh := func() {
		refreshSummary()
		if managerRefresh != nil {
			managerRefresh()
		}
	}
	button.OnTapped = func() {
		if managerRefresh != nil {
			return
		}
		button.Disable()
		managerRefresh = showAdditionalAudioRoutesManagerDialog(refreshSummary, func() {
			managerRefresh = nil
			button.Enable()
		})
		if managerRefresh == nil {
			button.Enable()
		}
	}
	Fields.AudioRoutesRefresh = refresh
	refresh()
	return button
}
