package Pages

import (
	"testing"

	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
)

func TestVibeVoiceStreamingPanelInEnglishAndGerman(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	previous := Settings.Config
	defer func() { Settings.Config = previous; lang.SetPreferredLocale("en") }()
	Settings.Config.Osc_chat_limit = 144
	Settings.Config.Special_settings = nil
	Settings.Config.Streaming_display_mode = ""
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatal(err)
	}
	for _, locale := range []struct{ code, label string }{{"en", "Replace old lines"}, {"de", "Alte Zeilen ersetzen"}} {
		lang.SetPreferredLocale(locale.code)
		panel := buildVibeVoiceStreamingSettings()
		form := panel.(*widget.Form)
		mode := form.Items[0].Widget.(*CustomWidget.TextValueSelect)
		if mode.Selected != locale.label || mode.GetSelected().Value != "blocks" {
			t.Fatalf("unexpected %s display control: %q", locale.code, mode.Selected)
		}
		entry := form.Items[1].Widget.(*widget.Entry)
		if entry.Text != "144" || entry.Validator("0") == nil || entry.Validator("145") != nil {
			t.Fatal("chatbox length must restore the profile and validate whole numbers")
		}
		window := test.NewWindow(panel)
		window.Resize(fyne.NewSize(640, 120))
		if panel.MinSize().Height > 120 {
			t.Fatalf("streaming controls must stay compact: %v", panel.MinSize())
		}
		if window.Canvas().Capture().Bounds().Empty() {
			t.Fatalf("empty %s panel render", locale.code)
		}
		window.Close()
	}
}

func TestAdditionalSourceLiveControlsAreUnderOutputsAndStayIndependent(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	previous := Settings.Config
	defer func() { Settings.Config = previous; lang.SetPreferredLocale("en") }()
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatal(err)
	}
	Settings.Config.Stt_type, Settings.Config.Model = "vibevoice_asr", "VibeVoice-ASR-Streaming-1.5B"
	Settings.Config.Streaming_display_mode, Settings.Config.Osc_chat_limit = "blocks", 144
	Settings.Config.Audio_api = "WASAPI"
	for _, locale := range []string{"en", "de"} {
		lang.SetPreferredLocale(locale)
		route := defaultAdditionalAudioRoute(1)
		route.Osc_enabled = true
		preview := newRouteAudioInputPreview(300)
		panel := createAudioRouteDetails(&route, preview).(*widget.Accordion)
		if panel.Items[3].Title != lang.L("Outputs") {
			t.Fatal("output controls must have their own section")
		}
		output := panel.Items[3].Detail.(*fyne.Container)
		var mode *CustomWidget.TextValueSelect
		var limit *widget.Entry
		for _, object := range output.Objects {
			if control, ok := object.(*CustomWidget.TextValueSelect); ok && control.Name == "streaming_display_"+route.ID {
				mode = control
			}
			if control, ok := object.(*widget.Entry); ok && control.PlaceHolder == lang.L("Use main setting")+": 144" {
				limit = control
			}
		}
		if mode == nil || mode.GetSelected().Value != "" || limit == nil || limit.Text != "" {
			t.Fatal("a new route must explicitly show inheritance for display mode and chat limit")
		}
		mode.SetSelected("rolling")
		limit.SetText("72")
		if route.Streaming_display_mode != "rolling" || route.Osc_chat_limit == nil || *route.Osc_chat_limit != 72 || Settings.Config.Streaming_display_mode != "blocks" || Settings.Config.Osc_chat_limit != 144 {
			t.Fatal("route controls changed main settings")
		}
		encoded, err := json.Marshal(route)
		if err != nil {
			t.Fatal(err)
		}
		var restored Settings.AdditionalAudioRoute
		if err := json.Unmarshal(encoded, &restored); err != nil {
			t.Fatal(err)
		}
		if restored.Streaming_display_mode != "rolling" || restored.Osc_chat_limit == nil || *restored.Osc_chat_limit != 72 {
			t.Fatal("route display settings were not persisted")
		}
		mode.SetSelected("")
		limit.SetText("")
		if route.Streaming_display_mode != "" || route.Osc_chat_limit != nil {
			t.Fatal("clearing overrides must restore inheritance")
		}
		panel.Close(0)
		panel.Open(3)
		window := test.NewWindow(panel)
		window.Resize(fyne.NewSize(850, 650))
		if window.Canvas().Capture().Bounds().Empty() {
			t.Fatal("empty outputs render")
		}
		window.Close()
		preview.Stop()
	}
}
