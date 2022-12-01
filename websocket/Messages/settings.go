package Messages

import (
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"strings"
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

	// fill combo-box with whisper languages
	if len(Fields.Field.TranscriptionSpeakerLanguageCombo.Options) < len(TranslateSettings.WhisperLanguages) {
		Fields.Field.TranscriptionSpeakerLanguageCombo.Options = nil
		for _, element := range TranslateSettings.WhisperLanguages {
			Fields.Field.TranscriptionSpeakerLanguageCombo.Options = append(Fields.Field.TranscriptionSpeakerLanguageCombo.Options, cases.Title(language.English, cases.Compact).String(element.Name))
		}
	}

	// Set options to current settings
	if strings.Contains(res.Whisper_task, "translate") && Fields.Field.TranscriptionTaskCombo.Selected != "translate (to en)" {
		Fields.Field.TranscriptionTaskCombo.SetSelected("translate (to en)")
	}
	if strings.Contains(res.Whisper_task, "transcribe") && !strings.Contains(Fields.Field.TranscriptionTaskCombo.Selected, "transcribe") {
		Fields.Field.TranscriptionTaskCombo.SetSelected("transcribe")
	}
	if Fields.Field.TranscriptionSpeakerLanguageCombo.Selected != TranslateSettings.GetWhisperLanguageNameByCode(res.Current_language) {
		Fields.Field.TranscriptionSpeakerLanguageCombo.SetSelected(TranslateSettings.GetWhisperLanguageNameByCode(res.Current_language))
	}
	if strings.ToLower(Fields.Field.TargetLanguageCombo.Selected) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Trg_lang)) {
		if TranslateSettings.Txt_translate {
			Fields.Field.TargetLanguageCombo.SetSelected(cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Trg_lang)))
		} else if Fields.Field.TargetLanguageCombo.Selected != "None" {
			// special case for "None" text translation target language
			Fields.Field.TargetLanguageCombo.SetSelected("None")
		}
	}
	Fields.DataBindings.TextToSpeechEnabledDataBinding.Set(res.Tts_answer)
	Fields.DataBindings.OSCEnabledDataBinding.Set(res.OscAutoProcessingEnabled)

	return &res
}

func (res TranslateSetting) GetWhisperLanguageCodeByName(name string) string {
	for _, entry := range res.WhisperLanguages {
		if strings.ToLower(entry.Name) == strings.ToLower(name) {
			return entry.Code
		}
	}
	return ""
}

func (res TranslateSetting) GetWhisperLanguageNameByCode(code string) string {
	for _, entry := range res.WhisperLanguages {
		if entry.Code == code {
			return entry.Name
		}
	}
	return ""
}
