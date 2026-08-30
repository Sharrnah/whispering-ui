package Pages

import (
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2/lang"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

func TestPluginRouteCheckDoesNotChangeRoutingWhileRendering(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	changed := 0
	newPluginRouteCheck("Main Microphone", true, func(bool) { changed++ })
	if changed != 0 {
		t.Fatalf("rendering a plugin route invoked change callback %d time(s)", changed)
	}
}

func TestPluginRoutingShowsEnabledAndAssignedDisabledPlugins(t *testing.T) {
	previous := Settings.Config.Plugins
	t.Cleanup(func() { Settings.Config.Plugins = previous })
	Settings.Config.Plugins = map[string]bool{
		"SubtitleDisplayPlugin":  true,
		"SubtitleExportPlugin":   false,
		"SecondaryProfilePlugin": true,
	}

	plugins := enabledAudioRoutePlugins()
	if !reflect.DeepEqual(plugins, []string{"SubtitleDisplayPlugin"}) {
		t.Fatalf("enabled route plugins = %#v", plugins)
	}

	options := audioRoutePluginOptions(
		plugins,
		[]Settings.AdditionalAudioRoute{{Plugins: []string{"SubtitleExportPlugin"}}},
		nil,
	)
	want := []audioRoutePluginOption{
		{Name: "SubtitleDisplayPlugin", Enabled: true},
		{Name: "SubtitleExportPlugin", Enabled: false},
	}
	if !reflect.DeepEqual(options, want) {
		t.Fatalf("plugin routing options = %#v, want %#v", options, want)
	}
}

func TestRouteLanguageCompletionMapsDisplayTextToCanonicalCode(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	options := []CustomWidget.TextValueOption{
		{Text: "Autodetect", Value: ""},
		{Text: "German", Value: "de"},
	}
	option, ok := matchRouteLanguageOption(options, "germ")
	if !ok || option.Value != "de" {
		t.Fatalf("language match = %#v, %v", option, ok)
	}

	selected := "unchanged"
	entry := newRouteLanguageCompletion(options, "de", func(value string) {
		selected = value
	})
	entry.OnSubmitted("Autodetect")
	if selected != "" || entry.Text != "Autodetect" {
		t.Fatalf("autodetect selected value/text = %q / %q", selected, entry.Text)
	}
}

func TestRouteSliderMatchesMainSettingsBehavior(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	changed := 0
	lastValue := 0.0
	slider, _ := newRouteSlider(
		0.4, 0, 1, 0.01, "%.2f", false,
		func(value float64) {
			changed++
			lastValue = value
		},
	)

	if changed != 0 {
		t.Fatalf("creating the slider changed the route %d time(s)", changed)
	}
	if slider.Step != 0.01 || slider.Value != 0.4 {
		t.Fatalf("slider step/value = %v/%v", slider.Step, slider.Value)
	}
	slider.SetValue(0.73)
	if changed != 1 || lastValue != 0.73 {
		t.Fatalf("slider update = %d/%v", changed, lastValue)
	}
}

func TestCloneAudioRoutesCopiesPluginSlices(t *testing.T) {
	original := []Settings.AdditionalAudioRoute{{
		ID:      "game",
		Plugins: []string{"SubtitlePlugin"},
	}}
	cloned := cloneAudioRoutes(original)
	cloned[0].Plugins[0] = "DifferentPlugin"

	if reflect.DeepEqual(original, cloned) || original[0].Plugins[0] != "SubtitlePlugin" {
		t.Fatalf("route plugin slices were not independently cloned: %#v / %#v", original, cloned)
	}
}

func TestNewRouteCopiesProfileAudioProcessingDefaults(t *testing.T) {
	previous := Settings.Config
	t.Cleanup(func() { Settings.Config = previous })
	Settings.Config.Realtime_frequency_time = 0.4
	Settings.Config.Silence_cutting_enabled = true
	Settings.Config.Denoise_audio = "noise_reduce"
	Settings.Config.Denoise_strength = 0.65
	Settings.Config.Vad_smart_turn_enabled = true
	Settings.Config.Vad_smart_turn_min_length = 1.5
	Settings.Config.Vad_smart_turn_probability_threshold = 0.7
	Settings.Config.Vad_smart_turn_pause_length = 0.25
	Settings.Config.Osc_chat_prefix = "[main] "

	route := defaultAdditionalAudioRoute(1)
	if route.Realtime_frequency_time != 0.4 ||
		!route.Silence_cutting_enabled ||
		route.Denoise_audio != "noise_reduce" ||
		route.Denoise_strength != 0.65 ||
		!route.Vad_smart_turn_enabled ||
		route.Vad_smart_turn_min_length != 1.5 ||
		route.Vad_smart_turn_probability_threshold != 0.7 ||
		route.Vad_smart_turn_pause_length != 0.25 {
		t.Fatalf("new route did not copy audio-processing defaults: %#v", route)
	}
	if route.Osc_enabled || route.Osc_typing_indicator || route.Osc_chat_notification {
		t.Fatalf("new route OSC output should start disabled: %#v", route)
	}
	if route.Osc_chat_prefix != "[main] " {
		t.Fatalf("new route OSC prefix = %q, want %q", route.Osc_chat_prefix, "[main] ")
	}
}

func TestAudioRoutePanelTitleCountsOnlyActiveSources(t *testing.T) {
	title := audioRoutePanelTitle([]Settings.AdditionalAudioRoute{
		{Enabled: true},
		{Enabled: false},
	})
	if !strings.Contains(title, "(1/2)") {
		t.Fatalf("route panel title = %q, want active/total count", title)
	}
}

func TestAudioRouteEnabledUpdateRequestTargetsOneRoute(t *testing.T) {
	request := audioRouteEnabledUpdateRequest("game", false)
	if request.Operation != "set_enabled" || request.RouteID != "game" {
		t.Fatalf("unexpected route toggle request: %#v", request)
	}
	if request.Enabled == nil || *request.Enabled {
		t.Fatalf("route toggle enabled value = %#v, want false", request.Enabled)
	}
	if request.Route != nil || len(request.Routes) != 0 {
		t.Fatalf("route toggle should not replace route configuration: %#v", request)
	}
}

func TestAudioRouteNameIsPartOfEnableControl(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)

	check := newAudioRouteEnabledCheck("Game chat", true)
	if check.Text != "Game chat" || !check.Checked {
		t.Fatalf("route enable control = %q/%v", check.Text, check.Checked)
	}
	fynetest.Tap(check)
	if check.Checked {
		t.Fatal("clicking the named route control did not toggle it")
	}
}

func TestAdditionalAudioRoutesManagerHasSourceAndRoutingTabs(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)

	tabs := newAdditionalAudioRoutesManagerTabs(widget.NewLabel("sources"), widget.NewLabel("routing"))
	if len(tabs.Items) != 2 {
		t.Fatalf("manager tabs = %d, want 2", len(tabs.Items))
	}
	if tabs.Items[0].Text != lang.L("Additional Audio Sources") ||
		tabs.Items[1].Text != lang.L("Plugin Routing") {
		t.Fatalf("unexpected manager tabs: %q / %q", tabs.Items[0].Text, tabs.Items[1].Text)
	}
}

func TestCompactAudioRouteButtonShowsConfiguredSources(t *testing.T) {
	application := fynetest.NewApp()
	t.Cleanup(application.Quit)
	previousConfig := Settings.Config
	previousCallback := Fields.AudioRoutesUpdateResult
	previousRefresh := Fields.AudioRoutesRefresh
	t.Cleanup(func() {
		Settings.Config = previousConfig
		Fields.AudioRoutesUpdateResult = previousCallback
		Fields.AudioRoutesRefresh = previousRefresh
	})
	Settings.Config.Additional_audio_routes = []Settings.AdditionalAudioRoute{{
		ID:                 "game",
		Name:               "Game chat",
		Enabled:            true,
		Audio_api:          "WASAPI",
		Audio_input_device: "VRChat.exe",
	}}

	buttonObject := createAdditionalAudioRoutesPanel()
	if buttonObject == nil || buttonObject.MinSize().Width <= 0 || buttonObject.MinSize().Height <= 0 {
		t.Fatalf("compact route button did not render: %#v", buttonObject)
	}
	button, ok := buttonObject.(*widget.Button)
	if !ok {
		t.Fatalf("route control = %T, want one manager button", buttonObject)
	}
	if !strings.Contains(button.Text, "(1/1)") {
		t.Fatalf("route button title = %q, want active/total count", button.Text)
	}
	if button.OnTapped == nil {
		t.Fatal("route manager button is not clickable")
	}
}

func TestRouteAudioPreviewSourceUsesDraftSelection(t *testing.T) {
	source := routeAudioPreviewSourceFromRoute(Settings.AdditionalAudioRoute{
		Audio_api:              "WASAPI",
		Audio_input_device:     "VRChat.exe",
		Audio_input_process:    "VRChat.exe",
		Audio_input_process_id: 42388,
	})
	if source.audioAPI != "WASAPI" || source.device != "VRChat.exe" ||
		source.process != "VRChat.exe" || source.processID != 42388 {
		t.Fatalf("preview source did not preserve draft selection: %#v", source)
	}
}
