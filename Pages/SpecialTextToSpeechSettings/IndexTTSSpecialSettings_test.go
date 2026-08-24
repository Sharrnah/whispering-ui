package SpecialTextToSpeechSettings

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/lang"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func TestIndexTTSGermanLanguageFollowsCheckpointSelection(t *testing.T) {
	previous := Fields.TtsModelSelectionValues
	defer func() { Fields.TtsModelSelectionValues = previous }()

	germanDisplay := "IndexTTS-2.5-German (German)"
	Fields.TtsModelSelectionValues = map[string][]string{
		germanDisplay: {"German", indexTTSGermanModel},
	}

	if actual := indexTTSCanonicalModel(germanDisplay); actual != indexTTSGermanModel {
		t.Fatalf("canonical German model = %q; want %q", actual, indexTTSGermanModel)
	}
	stockOptions := indexTTSLanguageOptions("IndexTTS-2.5 (Multilingual voice cloning)")
	if containsIndexTTSLanguage(stockOptions, "de") {
		t.Fatalf("stock IndexTTS unexpectedly advertises German: %#v", stockOptions)
	}
	germanOptions := indexTTSLanguageOptions(germanDisplay)
	if !containsIndexTTSLanguage(germanOptions, "de") {
		t.Fatalf("German IndexTTS does not advertise German: %#v", germanOptions)
	}
	if len(germanOptions) != len(stockOptions)+1 {
		t.Fatalf("German option count = %d; want %d", len(germanOptions), len(stockOptions)+1)
	}
}

func TestIndexTTSGermanLabelIsLocalized(t *testing.T) {
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatalf("load translations: %v", err)
	}
	defer lang.SetPreferredLocale("en")

	lang.SetPreferredLocale("de")
	options := indexTTSLanguageOptions(indexTTSGermanModel)
	if actual := options[len(options)-1].Text; actual != "Deutsch" {
		t.Fatalf("German language label = %q; want Deutsch", actual)
	}
}

func TestIndexTTSSettingsRemoveGermanWhenSwitchingToStock(t *testing.T) {
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

	germanDisplay := "IndexTTS-2.5-German (German)"
	stockDisplay := "IndexTTS-2.5 (Multilingual voice cloning)"
	Fields.Field.TtsModelCombo = widget.NewSelect(nil, nil)
	Fields.Field.TtsModelCombo.Selected = germanDisplay
	Fields.TtsModelSelectionValues = map[string][]string{
		germanDisplay: {"German", indexTTSGermanModel},
		stockDisplay:  {"Multilingual voice cloning", "IndexTTS-2.5"},
	}
	Settings.Config.Special_settings = map[string]interface{}{
		"tts_index_tts": map[string]interface{}{"language": "de"},
	}

	_, languageSelect := buildIndexTTSSpecialSettings()
	if selected := languageSelect.GetSelected(); selected == nil || selected.Value != "de" {
		t.Fatalf("initial German language selection = %#v; want de", selected)
	}

	received := make(chan SendMessageChannel.SendMessageStruct, 1)
	go func() { received <- <-SendMessageChannel.SendMessageChannel }()
	Fields.TtsModelSelectionChanged(stockDisplay)
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("switching to stock did not notify the backend about the language fallback")
	}

	if selected := languageSelect.GetSelected(); selected == nil || selected.Value != "auto" {
		t.Fatalf("stock language selection = %#v; want auto", selected)
	}
	if containsIndexTTSLanguage(languageSelect.Options, "de") {
		t.Fatalf("stock language options unexpectedly include German: %#v", languageSelect.Options)
	}
}
