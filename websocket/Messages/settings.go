package Messages

import (
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
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

	Settings.Form = Settings.BuildSettingsForm(nil).(*widget.Form)
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
		Fields.Field.TranscriptionSpeakerLanguageCombo.SetSelected(
			cases.Title(language.English, cases.Compact).String(TranslateSettings.GetWhisperLanguageNameByCode(res.Current_language)),
		)
	}

	// Set SourceLanguageCombo
	if strings.ToLower(Fields.Field.SourceLanguageCombo.Selected) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Src_lang)) {
		Fields.Field.SourceLanguageCombo.SetSelected(cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Src_lang)))
	} else if Fields.Field.SourceLanguageCombo.Selected == "" && res.Src_lang == "auto" {
		Fields.Field.SourceLanguageCombo.SetSelected(cases.Title(language.English, cases.Compact).String(res.Src_lang))
	}
	if Fields.Field.SourceLanguageComboTxtTranslateCombo.Selected == "" {
		Fields.Field.SourceLanguageComboTxtTranslateCombo.SetSelected(Fields.Field.SourceLanguageCombo.Selected)
	}

	// Set TargetLanguageCombo
	if strings.ToLower(Fields.Field.TargetLanguageCombo.Selected) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Trg_lang)) {
		if TranslateSettings.Txt_translate {
			Fields.Field.TargetLanguageCombo.SetSelected(cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Trg_lang)))
			// set text translate target language combo-box
			if Fields.Field.TargetLanguageTxtTranslateCombo.Selected == "" {
				Fields.Field.TargetLanguageTxtTranslateCombo.SetSelected(cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Trg_lang)))
			}
		} else if Fields.Field.TargetLanguageCombo.Selected != "None" {
			// special case for "None" text translation target language
			Fields.Field.TargetLanguageCombo.SetSelected("None")
		}
	}
	checkValue, _ := Fields.DataBindings.TextToSpeechEnabledDataBinding.Get()
	if checkValue != res.Tts_enabled {
		Fields.DataBindings.TextToSpeechEnabledDataBinding.Set(res.Tts_answer)
	}
	checkValue, _ = Fields.DataBindings.OSCEnabledDataBinding.Get()
	if checkValue != res.OscAutoProcessingEnabled {
		Fields.DataBindings.OSCEnabledDataBinding.Set(res.OscAutoProcessingEnabled)
	}

	// Set TtsModelCombo
	if Fields.Field.TtsModelCombo.Selected != res.Tts_model[1] {
		Fields.Field.TtsModelCombo.SetSelected(res.Tts_model[1])
	}

	// Set TtsVoiceCombo
	if Fields.Field.TtsVoiceCombo.Selected != res.Tts_voice {
		Fields.Field.TtsVoiceCombo.SetSelected(res.Tts_voice)
	}
	// Set OcrWindowCombo
	//if res.Ocr_window_name != Settings.Config.Ocr_window_name {
	if Fields.Field.OcrWindowCombo.Selected != res.Ocr_window_name {
		if !Utilities.Contains(Fields.Field.OcrWindowCombo.Options, res.Ocr_window_name) {
			Fields.Field.OcrWindowCombo.Options = append(Fields.Field.OcrWindowCombo.Options, res.Ocr_window_name)
		}
		Fields.Field.OcrWindowCombo.SetSelected(res.Ocr_window_name)
	}
	//}
	// Set OcrLanguageCombo
	if Fields.Field.OcrLanguageCombo.Selected != res.Ocr_lang {
		Fields.Field.OcrLanguageCombo.SetSelected(OcrLanguagesList.GetNameByCode(res.Ocr_lang))
	}

	Settings.Config = res.Conf

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
