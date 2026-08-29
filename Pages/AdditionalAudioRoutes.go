package Pages

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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
	RequestID        string                          `json:"request_id"`
	Routes           []Settings.AdditionalAudioRoute `json:"routes"`
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

func createRouteAudioControls(route *Settings.AdditionalAudioRoute) (*widget.Select, fyne.CanvasObject) {
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
			return
		}
		applicationSelection.Hide()
		route.Audio_input_device = option.Text
		route.Audio_input_process = ""
		route.Audio_input_process_id = 0
		route.Device_index = nil
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
	state := widget.NewLabel(fmt.Sprintf(valueFormat, slider.Value))
	slider.OnChanged = func(updated float64) {
		if extendMaximum && updated >= slider.Max {
			slider.Max += 10
		}
		state.SetText(fmt.Sprintf(valueFormat, updated))
		if changed != nil {
			changed(updated)
		}
	}
	return slider, container.NewBorder(nil, nil, nil, state, slider)
}

func createAudioRouteEditor(route *Settings.AdditionalAudioRoute, remove func()) fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetText(route.Name)
	name.OnChanged = func(value string) { route.Name = value }

	enabled := widget.NewCheck("", func(value bool) { route.Enabled = value })
	enabled.SetChecked(route.Enabled)
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

	apiSelection, inputSelection := createRouteAudioControls(route)

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

	_, energy := newRouteSlider(
		float64(route.Energy), 0, EnergySliderMax, 1, "%.0f", true,
		func(value float64) { route.Energy = int(value) },
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

	form := container.New(
		layout.NewFormLayout(),
		widget.NewLabel(lang.L("Audio Source Name")), name,
		widget.NewLabel(lang.L("Enabled")), enabled,
		widget.NewLabel(lang.L("Audio API")), apiSelection,
		widget.NewLabel(lang.L("Audio Input (mic)")), inputSelection,
		widget.NewLabel(lang.L("Speech-to-Text Enabled")), sttEnabled,
		widget.NewLabel(lang.L("Task")), task,
		widget.NewLabel(lang.L("Speech Language")), speechLanguage,
		widget.NewLabel(lang.L("Realtime")), realtime,
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
		widget.NewLabel(lang.L("energy.Name")), energy,
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
	processingAccordion := widget.NewAccordion(
		widget.NewAccordionItem(lang.L("Audio Processing"), processingForm),
	)
	removeButton := widget.NewButtonWithIcon(lang.L("Remove Audio Source"), theme.DeleteIcon(), remove)
	return widget.NewCard(route.Name, "", container.NewVBox(form, processingAccordion, removeButton))
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

func createAdditionalAudioRoutesSettings() fyne.CanvasObject {
	pluginNames := enabledAudioRoutePlugins()
	draftRoutes := cloneAudioRoutes(Settings.Config.Additional_audio_routes)
	committedRoutes := cloneAudioRoutes(draftRoutes)
	draftMainPlugins := clonePluginAllowlist(Settings.Config.Main_audio_plugins)
	committedMainPlugins := clonePluginAllowlist(draftMainPlugins)
	mainSelectionExplicit := draftMainPlugins != nil

	routesBox := container.NewVBox()
	routingBox := container.NewVBox()
	progress := widget.NewProgressBarInfinite()
	progress.Stop()
	progress.Hide()
	addButton := widget.NewButtonWithIcon(lang.L("Add Audio Source"), theme.ContentAddIcon(), nil)
	applyButton := widget.NewButtonWithIcon(lang.L("Apply Audio Sources"), theme.ConfirmIcon(), nil)
	pendingRequestID := ""

	var renderRoutes func()
	var renderPluginRouting func()
	renderRoutes = func() {
		routesBox.RemoveAll()
		if pendingRequestID == "" && len(draftRoutes) < maxAdditionalAudioRoutes {
			addButton.Enable()
		} else {
			addButton.Disable()
		}
		if len(draftRoutes) == 0 {
			message := widget.NewLabel(lang.L("No Additional Audio Sources Configured"))
			message.Alignment = fyne.TextAlignCenter
			routesBox.Add(message)
			return
		}
		for index := range draftRoutes {
			currentIndex := index
			routesBox.Add(createAudioRouteEditor(&draftRoutes[currentIndex], func() {
				draftRoutes = append(draftRoutes[:currentIndex], draftRoutes[currentIndex+1:]...)
				renderRoutes()
				renderPluginRouting()
			}))
		}
		routesBox.Refresh()
	}

	renderPluginRouting = func() {
		routingBox.RemoveAll()
		routingBox.Add(createPluginRouting(
			audioRoutePluginOptions(pluginNames, draftRoutes, draftMainPlugins),
			draftRoutes,
			draftMainPlugins,
			func(pluginName string, selected bool) {
				mainSelectionExplicit = true
				current := append([]string(nil), pluginNames...)
				if draftMainPlugins != nil {
					current = append([]string(nil), (*draftMainPlugins)...)
				}
				current = setPluginSelected(current, pluginName, selected)
				draftMainPlugins = &current
			},
			func(routeIndex int, pluginName string, selected bool) {
				draftRoutes[routeIndex].Plugins = setPluginSelected(
					draftRoutes[routeIndex].Plugins,
					pluginName,
					selected,
				)
			},
		))
		routingBox.Refresh()
	}

	addButton.OnTapped = func() {
		if len(draftRoutes) >= maxAdditionalAudioRoutes {
			return
		}
		draftRoutes = append(draftRoutes, defaultAdditionalAudioRoute(len(draftRoutes)+1))
		renderRoutes()
		renderPluginRouting()
	}
	applyButton.OnTapped = func() {
		if pendingRequestID != "" {
			return
		}
		for _, route := range draftRoutes {
			if strings.TrimSpace(route.Name) == "" {
				dialog.ShowError(errors.New(lang.L("Audio Source Name Is Required")), fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
		}
		requestMainPlugins := draftMainPlugins
		if !mainSelectionExplicit {
			requestMainPlugins = nil
		}
		request := audioRoutesUpdateRequest{
			RequestID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Routes:           cloneAudioRoutes(draftRoutes),
			MainAudioPlugins: clonePluginAllowlist(requestMainPlugins),
		}
		pendingRequestID = request.RequestID
		addButton.Disable()
		applyButton.Disable()
		progress.Show()
		progress.Start()
		SendMessageChannel.SendMessageStruct{Type: "audio_routes_update", Value: request}.SendMessage()
	}

	Fields.AudioRoutesUpdateResult = func(requestID string, success bool, errorMessage string) {
		if requestID == "" || requestID != pendingRequestID {
			return
		}
		pendingRequestID = ""
		progress.Stop()
		progress.Hide()
		if len(draftRoutes) < maxAdditionalAudioRoutes {
			addButton.Enable()
		} else {
			addButton.Disable()
		}
		applyButton.Enable()
		if success {
			committedRoutes = cloneAudioRoutes(draftRoutes)
			committedMainPlugins = clonePluginAllowlist(draftMainPlugins)
			Settings.Config.Additional_audio_routes = cloneAudioRoutes(draftRoutes)
			Settings.Config.Main_audio_plugins = clonePluginAllowlist(draftMainPlugins)
			return
		}

		draftRoutes = cloneAudioRoutes(committedRoutes)
		draftMainPlugins = clonePluginAllowlist(committedMainPlugins)
		mainSelectionExplicit = draftMainPlugins != nil
		renderRoutes()
		renderPluginRouting()
		if strings.TrimSpace(errorMessage) == "" {
			errorMessage = lang.L("Error")
		}
		if app := fyne.CurrentApp(); app != nil && len(app.Driver().AllWindows()) > 0 {
			dialog.ShowError(errors.New(errorMessage), app.Driver().AllWindows()[0])
		}
	}

	renderRoutes()
	renderPluginRouting()
	explanation := widget.NewLabel(lang.L("Additional Audio Sources Share the Loaded Speech Model"))
	explanation.Wrapping = fyne.TextWrapWord
	content := container.NewVBox(
		explanation,
		routesBox,
		container.NewHBox(addButton),
		widget.NewCard(lang.L("Plugin Routing"), "", routingBox),
	)
	scroll := container.NewVScroll(content)
	footer := container.NewVBox(
		widget.NewSeparator(),
		progress,
		container.NewHBox(layout.NewSpacer(), applyButton),
	)
	return container.NewBorder(nil, footer, nil, nil, scroll)
}
