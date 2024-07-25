package Messages

import (
	"fmt"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path/filepath"
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

	Settings.Form = Settings.BuildSettingsForm(nil, filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)).(*widget.Form)
	Settings.Form.Refresh()

	// fill combo-box with whisper languages
	if len(Fields.Field.TranscriptionSpeakerLanguageCombo.Options) < len(TranslateSettings.WhisperLanguages) {
		Fields.Field.TranscriptionSpeakerLanguageCombo.Options = nil
		for _, element := range TranslateSettings.WhisperLanguages {
			Fields.Field.TranscriptionSpeakerLanguageCombo.Options = append(Fields.Field.TranscriptionSpeakerLanguageCombo.Options, cases.Title(language.English, cases.Compact).String(element.Name))
		}
	}
	// fill combo-box with target languages
	if len(Fields.Field.TranscriptionTargetLanguageCombo.Options) < len(TranslateSettings.WhisperLanguages) {
		Fields.Field.TranscriptionTargetLanguageCombo.Options = nil
		for _, element := range TranslateSettings.WhisperLanguages {
			if element.Code != "" && strings.ToLower(element.Code) != "auto" {
				Fields.Field.TranscriptionTargetLanguageCombo.Options = append(Fields.Field.TranscriptionTargetLanguageCombo.Options, cases.Title(language.English, cases.Compact).String(element.Name))
			}
		}
	}

	// Set options to current settings
	if strings.Contains(res.Whisper_task, "translate") && Fields.Field.TranscriptionTaskCombo.Selected != "translate (to English)" {
		Fields.Field.TranscriptionTaskCombo.SetSelected("translate (to English)")
	}
	if strings.Contains(res.Whisper_task, "transcribe") && !strings.Contains(Fields.Field.TranscriptionTaskCombo.Selected, "transcribe") {
		Fields.Field.TranscriptionTaskCombo.SetSelected("transcribe")
	}
	if Fields.Field.TranscriptionSpeakerLanguageCombo.Text != TranslateSettings.GetWhisperLanguageNameByCode(res.Current_language) {
		Fields.Field.TranscriptionSpeakerLanguageCombo.Text = cases.Title(language.English, cases.Compact).String(TranslateSettings.GetWhisperLanguageNameByCode(res.Current_language))
		Fields.Field.TranscriptionSpeakerLanguageCombo.ResetOptionsFilter()
	}
	if Fields.Field.TranscriptionTargetLanguageCombo.Text != TranslateSettings.GetWhisperLanguageNameByCode(res.Target_language) {
		Fields.Field.TranscriptionTargetLanguageCombo.Text = cases.Title(language.English, cases.Compact).String(TranslateSettings.GetWhisperLanguageNameByCode(res.Target_language))
		Fields.Field.TranscriptionTargetLanguageCombo.ResetOptionsFilter()
	}

	// Set SourceLanguageCombo
	if strings.ToLower(Fields.Field.SourceLanguageCombo.Text) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Src_lang)) {
		if Fields.Field.SourceLanguageCombo.Text == "" {
			Fields.Field.SourceLanguageCombo.Text = cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Src_lang))
		}
	} else if Fields.Field.SourceLanguageCombo.Text == "" && strings.ToLower(res.Src_lang) == "auto" {
		if Fields.Field.SourceLanguageCombo.Text == "" {
			Fields.Field.SourceLanguageCombo.Text = cases.Title(language.English, cases.Compact).String(res.Src_lang)
		}
	}
	Fields.Field.SourceLanguageCombo.ResetOptionsFilter()

	// ocr text translation combo fields
	if strings.ToLower(Fields.Field.SourceLanguageTxtTranslateCombo.Text) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Ocr_txt_src_lang)) {
		if Fields.Field.SourceLanguageTxtTranslateCombo.Text == "" {
			Fields.Field.SourceLanguageTxtTranslateCombo.Text = cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Ocr_txt_src_lang))
		}
	} else if Fields.Field.SourceLanguageTxtTranslateCombo.Text == "" && strings.ToLower(res.Ocr_txt_src_lang) == "auto" {
		if Fields.Field.SourceLanguageTxtTranslateCombo.Text == "" {
			Fields.Field.SourceLanguageTxtTranslateCombo.Text = cases.Title(language.English, cases.Compact).String(res.Ocr_txt_src_lang)
		}
	}
	Fields.Field.SourceLanguageTxtTranslateCombo.ResetOptionsFilter()

	// Set TargetLanguageCombo
	if strings.ToLower(Fields.Field.TargetLanguageCombo.Text) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Trg_lang)) {
		Fields.Field.TargetLanguageCombo.Text = cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Trg_lang))
		Fields.Field.TargetLanguageCombo.ResetOptionsFilter()
	}

	// ocr text translation combo fields
	if strings.ToLower(Fields.Field.TargetLanguageTxtTranslateCombo.Text) != strings.ToLower(InstalledLanguages.GetNameByCode(res.Ocr_txt_trg_lang)) {
		//if Fields.Field.TargetLanguageTxtTranslateCombo.Text == "" {
		Fields.Field.TargetLanguageTxtTranslateCombo.Text = cases.Title(language.English, cases.Compact).String(InstalledLanguages.GetNameByCode(res.Ocr_txt_trg_lang))
		//}
		Fields.Field.TargetLanguageTxtTranslateCombo.ResetOptionsFilter()
	}

	Fields.Field.TextTranslateEnabled.Text = lang.L("SttTextTranslateLabel", map[string]interface{}{
		"FromLang": Fields.Field.SourceLanguageCombo.GetValueOptionEntryByText(Fields.Field.SourceLanguageCombo.Text).Value,
		"ToLang":   Fields.Field.TargetLanguageCombo.Text,
	})
	Fields.Field.TextTranslateEnabled.Refresh()

	checkValue, _ := Fields.DataBindings.SpeechToTextEnabledDataBinding.Get()
	if checkValue != res.Stt_enabled {
		Fields.DataBindings.SpeechToTextEnabledDataBinding.Set(res.Stt_enabled)
	}

	checkValue, _ = Fields.DataBindings.TextTranslateEnabledDataBinding.Get()
	if checkValue != res.Txt_translate {
		Fields.DataBindings.TextTranslateEnabledDataBinding.Set(res.Txt_translate)
	}

	checkValue, _ = Fields.DataBindings.TextToSpeechEnabledDataBinding.Get()
	if checkValue != res.Tts_answer {
		Fields.DataBindings.TextToSpeechEnabledDataBinding.Set(res.Tts_answer)
	}

	checkValue, _ = Fields.DataBindings.OSCEnabledDataBinding.Get()
	if checkValue != res.OscAutoProcessingEnabled {
		Fields.DataBindings.OSCEnabledDataBinding.Set(res.OscAutoProcessingEnabled)
	}

	// Set TtsModelCombo
	if len(res.Tts_model) > 0 && len(Fields.Field.TtsModelCombo.Options) > 0 && Fields.Field.TtsModelCombo.Selected != res.Tts_model[1] {
		Fields.Field.TtsModelCombo.SetSelected(res.Tts_model[1])
	}

	// Set TtsVoiceCombo
	// only set new tts voice if select is not received tts_voice and
	// if select is not empty and does not contain only one empty element
	if Fields.Field.TtsVoiceCombo.Selected != res.Tts_voice && (len(Fields.Field.TtsVoiceCombo.Options) > 0 &&
		(len(Fields.Field.TtsVoiceCombo.Options) == 1 && Fields.Field.TtsVoiceCombo.Options[0] != "")) {
		Fields.Field.TtsVoiceCombo.SetSelected(res.Tts_voice)
	}
	// Set OcrWindowCombo
	if Fields.Field.OcrWindowCombo.Selected != res.Ocr_window_name {
		if !Utilities.Contains(Fields.Field.OcrWindowCombo.Options, res.Ocr_window_name) {
			Fields.Field.OcrWindowCombo.Options = append(Fields.Field.OcrWindowCombo.Options, res.Ocr_window_name)
		}
		Fields.Field.OcrWindowCombo.SetSelected(res.Ocr_window_name)
	}
	//}
	// Set OcrLanguageCombo
	if Fields.Field.OcrLanguageCombo.Text != res.Ocr_lang {
		Fields.Field.OcrLanguageCombo.Text = OcrLanguagesList.GetNameByCode(res.Ocr_lang)
	}
	Fields.Field.OcrLanguageCombo.ResetOptionsFilter()

	// set oscEnabledLabel Update function
	if res.OscAutoProcessingEnabled {
		Fields.Field.OscLimitHint.Show()
	} else {
		Fields.Field.OscLimitHint.Hide()
	}
	Fields.OscLimitHintUpdateFunc = func() {
		transcriptionInputCount := Utilities.CountUTF16CodeUnits(Fields.Field.TranscriptionInput.Text)
		transcriptionTranslationInputCount := Utilities.CountUTF16CodeUnits(Fields.Field.TranscriptionTranslationInput.Text)
		oscSplitCount := Utilities.CountUTF16CodeUnits(Settings.Config.Osc_type_transfer_split)
		maxCount := res.Conf.Osc_chat_limit

		Fields.Field.OscLimitHint.Text = fmt.Sprintf(Fields.OscLimitLabelConst, 0, maxCount)
		switch res.Conf.Osc_type_transfer {
		case "source":
			Fields.Field.OscLimitHint.Text = fmt.Sprintf(Fields.OscLimitLabelConst, transcriptionInputCount, maxCount)
		case "translation_result":
			Fields.Field.OscLimitHint.Text = fmt.Sprintf(Fields.OscLimitLabelConst, transcriptionTranslationInputCount, maxCount)
		case "both":
			Fields.Field.OscLimitHint.Text = fmt.Sprintf(Fields.OscLimitLabelConst, transcriptionInputCount+oscSplitCount+transcriptionTranslationInputCount, maxCount)
		}
		Fields.Field.OscLimitHint.Refresh()
	}
	Fields.OscLimitHintUpdateFunc()

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
