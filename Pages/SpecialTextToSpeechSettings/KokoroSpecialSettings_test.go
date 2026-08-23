package SpecialTextToSpeechSettings

import (
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2/lang"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func TestKokoroCanonicalModel(t *testing.T) {
	previous := Fields.TtsModelSelectionValues
	defer func() { Fields.TtsModelSelectionValues = previous }()

	Fields.TtsModelSelectionValues = map[string][]string{
		"Kokoro-German-Thorsten (German)": {"German", kokoroThorstenModel},
		"Kokoro-Russian-v2 (Russian)":     {"Russian", kokoroRussianModel},
	}

	cases := map[string]string{
		"Kokoro-German-Thorsten (German)": kokoroThorstenModel,
		"Kokoro-Russian-v2 (Russian)":     kokoroRussianModel,
		kokoroThorstenModel:               kokoroThorstenModel,
		kokoroRussianModel:                kokoroRussianModel,
		"kokoro-v1_0 (Default)":           "kokoro-v1_0 (Default)",
	}
	for displayValue, expected := range cases {
		if actual := kokoroCanonicalModel(displayValue); actual != expected {
			t.Fatalf("kokoroCanonicalModel(%q) = %q; want %q", displayValue, actual, expected)
		}
	}

	want := []string{"German", kokoroThorstenModel}
	if !reflect.DeepEqual(Fields.TtsModelSelectionValues["Kokoro-German-Thorsten (German)"], want) {
		t.Fatal("canonical Kokoro model mapping changed unexpectedly")
	}
	if option, forced := kokoroForcedLanguage("Kokoro-German-Thorsten (German)"); !forced || option.Value != "d" {
		t.Fatalf("Thorsten forced language = %#v, %v; want German", option, forced)
	}
	if option, forced := kokoroForcedLanguage("Kokoro-Russian-v2 (Russian)"); !forced || option.Value != "r" {
		t.Fatalf("Russian v2 forced language = %#v, %v; want Russian", option, forced)
	}
	if _, forced := kokoroForcedLanguage("kokoro-v1_0 (Default)"); forced {
		t.Fatal("stock Kokoro should not have a forced language")
	}
}

func TestKokoroGermanLabelIsLocalized(t *testing.T) {
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatalf("load translations: %v", err)
	}
	defer lang.SetPreferredLocale("en")

	lang.SetPreferredLocale("en")
	if actual := lang.L("German"); actual != "German" {
		t.Fatalf("English German label = %q", actual)
	}
	lang.SetPreferredLocale("de")
	if actual := lang.L("German"); actual != "Deutsch" {
		t.Fatalf("German German label = %q; want Deutsch", actual)
	}
}

func TestKokoroSettingsBuildDoesNotWaitForWebsocketConsumer(t *testing.T) {
	app := fynetest.NewApp()
	defer app.Quit()

	previousCombo := Fields.Field.TtsModelCombo
	previousCallback := Fields.TtsModelSelectionChanged
	previousSelections := Fields.TtsModelSelectionValues
	previousSpecialSettings := Settings.Config.Special_settings
	defer func() {
		Fields.Field.TtsModelCombo = previousCombo
		Fields.TtsModelSelectionChanged = previousCallback
		Fields.TtsModelSelectionValues = previousSelections
		Settings.Config.Special_settings = previousSpecialSettings
	}()

	displayValue := "Kokoro-German-Thorsten (German)"
	Fields.Field.TtsModelCombo = widget.NewSelect(nil, nil)
	Fields.Field.TtsModelCombo.Selected = displayValue
	Fields.TtsModelSelectionValues = map[string][]string{
		displayValue:                  {"German", kokoroThorstenModel},
		"Kokoro-Russian-v2 (Russian)": {"Russian", kokoroRussianModel},
	}
	Settings.Config.Special_settings = map[string]interface{}{
		"tts_kokoro": map[string]interface{}{"language": "a"},
	}

	languageResult := make(chan *CustomWidget.CompletionEntry, 1)
	go func() {
		_, languageSelect := buildKokoroSpecialSettings()
		languageResult <- languageSelect
	}()

	var languageSelect *CustomWidget.CompletionEntry
	select {
	case languageSelect = <-languageResult:
	case <-time.After(time.Second):
		t.Fatal("building Kokoro settings blocked without a WebSocket message consumer")
	}

	kokoroSettings := Settings.Config.Special_settings["tts_kokoro"].(map[string]interface{})
	if actual := kokoroSettings["language"]; actual != "d" {
		t.Fatalf("local Kokoro language = %v; want d", actual)
	}
	if !languageSelect.Disabled() {
		t.Fatal("Thorsten language selector should be disabled")
	}
	if values := completionEntryValues(languageSelect); !reflect.DeepEqual(values, []string{"d"}) {
		t.Fatalf("Thorsten language options = %v; want [d]", values)
	}

	received := make(chan SendMessageChannel.SendMessageStruct, 1)
	go func() {
		received <- <-SendMessageChannel.SendMessageChannel
	}()
	Fields.TtsModelSelectionChanged("Kokoro-Russian-v2 (Russian)")
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("switching to Russian did not notify the backend")
	}
	if !languageSelect.Disabled() {
		t.Fatal("Russian language selector should be disabled")
	}
	if values := completionEntryValues(languageSelect); !reflect.DeepEqual(values, []string{"r"}) {
		t.Fatalf("Russian language options = %v; want [r]", values)
	}

	received = make(chan SendMessageChannel.SendMessageStruct, 1)
	go func() {
		received <- <-SendMessageChannel.SendMessageChannel
	}()
	Fields.TtsModelSelectionChanged("kokoro-v1_0 (Default)")
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("switching from Thorsten did not notify the backend")
	}
	if languageSelect.Disabled() {
		t.Fatal("stock Kokoro language selector should be enabled")
	}
	values := completionEntryValues(languageSelect)
	if slicesContain(values, "d") {
		t.Fatalf("stock Kokoro unexpectedly advertises German: %v", values)
	}
	if slicesContain(values, "r") {
		t.Fatalf("stock Kokoro unexpectedly advertises Russian: %v", values)
	}
	if len(values) != 9 {
		t.Fatalf("stock Kokoro language option count = %d; want 9", len(values))
	}
}

func completionEntryValues(entry *CustomWidget.CompletionEntry) []string {
	values := make([]string, 0, len(entry.OptionsTextValue))
	for _, option := range entry.OptionsTextValue {
		values = append(values, option.Value)
	}
	return values
}

func slicesContain(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
