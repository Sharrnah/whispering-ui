package ProfileForm

import (
	"reflect"
	"testing"

	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/test"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
)

func TestAudioCppNewProfilesRoundTripInEnglishAndGerman(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatal(err)
	}
	defer lang.SetPreferredLocale("")
	for _, locale := range []string{"en", "de"} {
		lang.SetPreferredLocale(locale)
		for _, model := range audioCppAdditionalTTSModels {
			controls := &AllProfileControls{}
			engine := NewFormEngine(controls, &Coordinator{Controls: controls})
			controls.TTSType = CustomWidget.NewTextValueSelect("tts_type", TTSTypeOptions(), nil, -1)
			controls.TTSType.SetSelected("audio_cpp")
			models, _, _ := TTSModelOptions("audio_cpp")
			controls.TTSModel = CustomWidget.NewTextValueSelect("tts_model", models, nil, 0)
			controls.TTSPrecision = CustomWidget.NewTextValueSelect("tts_precision", GenericTTSPrecisionOptions(), nil, 0)
			engine.Register("tts_model", controls.TTSModel)
			engine.Register("tts_precision", controls.TTSPrecision)
			for _, precision := range model.precisions {
				conf := Settings.Conf{Tts_type: "audio_cpp", Tts_model: []string{model.group, model.name}, Tts_precision: precision}
				engine.LoadFromSettings(&conf)
				engine.SaveToSettings(&conf)
				if !reflect.DeepEqual(conf.Tts_model, []string{model.group, model.name}) || conf.Tts_precision != precision {
					t.Fatalf("%s: lost model/precision %s/%s: %#v / %s", locale, model.name, precision, conf.Tts_model, conf.Tts_precision)
				}
			}
		}
	}
}
