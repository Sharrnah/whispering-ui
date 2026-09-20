package Advanced

import (
	"reflect"
	"testing"

	"whispering-tiger-ui/Settings"
)

func TestNewlyEnabledPluginJoinsExplicitMainMicrophoneRoute(t *testing.T) {
	previous := Settings.Config
	t.Cleanup(func() { Settings.Config = previous })
	selected := []string{"ExistingPlugin"}
	Settings.Config.Main_audio_plugins = &selected

	if !routeNewlyEnabledPluginToMainMicrophone("SubtitlePlugin") {
		t.Fatal("new plugin was not added")
	}
	want := []string{"ExistingPlugin", "SubtitlePlugin"}
	if !reflect.DeepEqual(*Settings.Config.Main_audio_plugins, want) {
		t.Fatalf("main microphone plugins = %#v, want %#v", *Settings.Config.Main_audio_plugins, want)
	}
	if routeNewlyEnabledPluginToMainMicrophone("SubtitlePlugin") {
		t.Fatal("existing plugin was added twice")
	}
}

func TestNewlyEnabledPluginKeepsLegacyAllPluginMode(t *testing.T) {
	previous := Settings.Config
	t.Cleanup(func() { Settings.Config = previous })
	Settings.Config.Main_audio_plugins = nil

	if routeNewlyEnabledPluginToMainMicrophone("SubtitlePlugin") {
		t.Fatal("legacy all-plugin mode was converted to an explicit allowlist")
	}
	if Settings.Config.Main_audio_plugins != nil {
		t.Fatalf("main microphone plugins = %#v, want nil", Settings.Config.Main_audio_plugins)
	}
}

func TestSecondaryProfileCompatibilityPluginIsNotRouted(t *testing.T) {
	previous := Settings.Config
	t.Cleanup(func() { Settings.Config = previous })
	selected := []string{}
	Settings.Config.Main_audio_plugins = &selected

	if routeNewlyEnabledPluginToMainMicrophone("SecondaryProfilePlugin") {
		t.Fatal("compatibility plugin was routed to the main microphone")
	}
	if len(*Settings.Config.Main_audio_plugins) != 0 {
		t.Fatalf("main microphone plugins = %#v, want empty", *Settings.Config.Main_audio_plugins)
	}
}
