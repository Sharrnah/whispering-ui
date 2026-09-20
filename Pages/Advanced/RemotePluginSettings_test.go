package Advanced

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"image/png"
	"os"
	"reflect"
	"testing"
	"whispering-tiger-ui/LocalPlugins"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func walkPluginControls(object fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	visit(object)
	switch item := object.(type) {
	case *fyne.Container:
		for _, child := range item.Objects {
			walkPluginControls(child, visit)
		}
	case *container.Scroll:
		walkPluginControls(item.Content, visit)
	case *widget.Accordion:
		for _, child := range item.Items {
			walkPluginControls(child.Detail, visit)
		}
	case *container.AppTabs:
		for _, child := range item.Items {
			walkPluginControls(child.Content, visit)
		}
	}
}

func TestLocalPluginControlsDoNotChangeAIProfile(t *testing.T) {
	app := test.NewTempApp(t)
	original := Settings.Config
	defer func() { Settings.Config = original }()
	Settings.Config = Settings.Conf{Run_backend: false, Plugins: map[string]bool{"RemoteOnly": true}, Plugin_settings: map[string]interface{}{"RemoteOnly": map[string]interface{}{"value": "remote"}}}
	before, _ := json.Marshal(Settings.Config)
	config := Settings.Conf{Plugins: map[string]bool{}, Plugin_settings: map[string]interface{}{"LocalOnly": map[string]interface{}{"message": map[string]interface{}{"type": "textfield", "value": "old"}, "test": map[string]interface{}{"type": "button", "label": "Test"}}}}
	var sent []interface{}
	scope := localPluginScope(&config, func(message interface{}) { sent = append(sent, message) })
	panel := BuildSinglePluginSettings("LocalOnly", nil, nil, nil, nil, scope)
	walkPluginControls(panel, func(object fyne.CanvasObject) {
		if checkbox, ok := object.(*widget.Check); ok {
			checkbox.SetChecked(true)
		}
		if entry, ok := object.(*widget.Entry); ok {
			entry.SetText("new")
		}
		if button, ok := object.(*widget.Button); ok && button.Text == "Test" {
			button.OnTapped()
		}
	})
	if len(sent) != 3 {
		t.Fatalf("expected local enable, field and button messages, got %#v", sent)
	}
	if !app.Preferences().Bool("local.plugins.enabled") {
		t.Fatal("enabled integrations must auto-start on reconnect")
	}
	for _, message := range sent[:2] {
		if message.(map[string]interface{})["type"] != "configure" {
			t.Fatal(message)
		}
	}
	if sent[2].(SendMessageChannel.SendMessageStruct).Type != "plugin_button_press" {
		t.Fatal(sent[2])
	}
	after, _ := json.Marshal(Settings.Config)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("local plugin controls modified the AI profile")
	}
}

func TestRemotePluginCatalogUsesAIState(t *testing.T) {
	test.NewTempApp(t)
	original := Settings.Config
	defer func() { Settings.Config = original }()
	Settings.Config = Settings.Conf{Run_backend: false, Plugins: map[string]bool{"OnlyOnAI": true}, Plugin_settings: map[string]interface{}{}}
	previous := onlyShowEnabledPlugins
	onlyShowEnabledPlugins = false
	defer func() { onlyShowEnabledPlugins = previous }()
	accordion, count := BuildPluginSettingsAccordion(nil)
	if count != 1 || len(accordion.(*widget.Accordion).Items) != 1 {
		t.Fatalf("AI catalog missing: %d", count)
	}
}

func TestLocalPluginPageKeepsEditorOnSettingsEcho(t *testing.T) {
	test.NewTempApp(t)
	var render func(LocalPlugins.Snapshot)
	panel := createLocalPluginSettingsPage(func(callback func(LocalPlugins.Snapshot), _ func(string)) { render = callback })
	state := LocalPlugins.Snapshot{Available: []string{"Example"}, Enabled: []string{"Example"}, Settings: map[string]map[string]interface{}{"Example": {"message": "hello"}}}
	render(state)
	var before fyne.CanvasObject
	walkPluginControls(panel, func(object fyne.CanvasObject) {
		if _, ok := object.(*widget.Accordion); ok {
			before = object
		}
	})
	render(state)
	var after fyne.CanvasObject
	walkPluginControls(panel, func(object fyne.CanvasObject) {
		if _, ok := object.(*widget.Accordion); ok {
			after = object
		}
	})
	if before != after {
		t.Fatal("acknowledgment recreated the focused editor")
	}
	state.Available = append(state.Available, "NewlyInstalled")
	render(state)
	walkPluginControls(panel, func(object fyne.CanvasObject) {
		if accordion, ok := object.(*widget.Accordion); ok && len(accordion.Items) != 2 {
			t.Fatal("new plugin not shown")
		}
	})
}

func TestRemotePluginPreview(t *testing.T) {
	fixture := os.Getenv("WT_PLUGIN_UI_FIXTURE")
	if fixture == "" {
		t.Skip("optional software-rendered preview")
	}
	lang.SetLanguageOrder([]string{"en"})
	lang.SetPreferredLocale(os.Getenv("PREFERRED_LANGUAGE"))
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatal(err)
	}
	test.NewTempApp(t)
	var render func(LocalPlugins.Snapshot)
	local := createLocalPluginSettingsPage(func(callback func(LocalPlugins.Snapshot), _ func(string)) { render = callback })
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var state LocalPlugins.Snapshot
	if err = json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	render(state)
	walkPluginControls(local, func(object fyne.CanvasObject) {
		if accordion, ok := object.(*widget.Accordion); ok && len(accordion.Items) > 0 {
			accordion.Open(0)
		}
	})
	panel := container.NewAppTabs(container.NewTabItem(lang.L("This PC"), local), container.NewTabItem(lang.L("AI PC"), widget.NewLabel("")))
	window := test.NewWindow(panel)
	defer window.Close()
	window.Resize(fyne.NewSize(1000, 900))
	for i := 0; i < 3; i++ {
		panel.Refresh()
		window.Canvas().Refresh(panel)
	}
	file, err := os.Create(os.Getenv("WT_PLUGIN_UI_PREVIEW"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = png.Encode(file, window.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
