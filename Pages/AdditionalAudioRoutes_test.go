package Pages

import (
	"reflect"
	"testing"

	fynetest "fyne.io/fyne/v2/test"
	"whispering-tiger-ui/CustomWidget"
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
