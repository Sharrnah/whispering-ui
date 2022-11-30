package Messages

import (
	"fyne.io/fyne/v2/widget"
	"log"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

type WhisperLanguage struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type TranslateSetting struct {
	WhisperLanguages         []WhisperLanguage `json:"whisper_languages"`
	OscAutoProcessingEnabled bool              `json:"osc_auto_processing_enabled"`
	Settings.Conf
}

var TranslateSettings TranslateSetting

func (res TranslateSetting) Update() *TranslateSetting {
	Settings.Config = res.Conf

	Settings.Form = Settings.BuildSettingsForm().(*widget.Form)
	Settings.Form.Refresh()

	log.Println("InstalledLanguages.GetNameByCode()")
	log.Println(InstalledLanguages.GetNameByCode(res.Trg_lang))
	log.Println(res.Trg_lang)

	// Set options to current settings
	Fields.Field.TargetLanguageCombo.SetSelected(InstalledLanguages.GetNameByCode(res.Trg_lang))
	Fields.DataBindings.TextToSpeechEnabledDataBinding.Set(res.Tts_answer)
	Fields.DataBindings.OSCEnabledDataBinding.Set(res.OscAutoProcessingEnabled)

	return &res
}
