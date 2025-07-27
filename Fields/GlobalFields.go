package Fields

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"image/color"
	"log"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/SendMessageChannel"
)

const OscLimitLabelConst = "[%d / %d]"

var OscLimitHintUpdateFunc = func() {}

var fieldCreationFunctions = struct {
	TranscriptionTaskCombo        func() *CustomWidget.TextValueSelect
	TranscriptionInput            func(dataBinding binding.String) *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationInput func(dataBinding binding.String) *CustomWidget.EntryWithPopupMenu
	TextTranslateEnabled          func() *widget.Check
	SttEnabled                    func() *widget.Check
	TtsEnabled                    func(dataBinding binding.Bool) *widget.Check
	OscEnabled                    func() *widget.Check
}{
	TranscriptionTaskCombo: func() *CustomWidget.TextValueSelect {
		return CustomWidget.NewTextValueSelect("whisper_task", []CustomWidget.TextValueOption{{
			Text:  lang.L("transcribe"),
			Value: "transcribe",
		}, {
			Text:  lang.L("translate (to English)"),
			Value: "translate",
		}}, func(valueOption CustomWidget.TextValueOption) {
			value := valueOption.Value
			switch valueOption.Value {
			case "transcribe":
				value = "transcribe"
			case "translate":
				value = "translate"
			case "translate (to English)": // deprecated
				value = "translate"
			}
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "setting_change",
				Name:  "whisper_task",
				Value: value,
			}
			sendMessage.SendMessage()
		}, 0)
	},
	TranscriptionInput: func(dataBinding binding.String) *CustomWidget.EntryWithPopupMenu {
		entry := CustomWidget.NewMultiLineEntryWithData(dataBinding)
		entry.Wrapping = fyne.TextWrapWord
		entry.Validator = nil

		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to Text-to-Speech"), func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to OSC (VRChat)"), func() {
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "send_osc",
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to Both (TTS + OSC)"), func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessageTts := SendMessageChannel.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessageTts.SendMessage()
			sendMessageOsc := SendMessageChannel.SendMessageStruct{
				Type: "send_osc",
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessageOsc.SendMessage()
		}))
		return entry
	},
	TranscriptionTranslationInput: func(dataBinding binding.String) *CustomWidget.EntryWithPopupMenu {
		entry := CustomWidget.NewMultiLineEntryWithData(dataBinding)
		entry.Wrapping = fyne.TextWrapWord
		entry.Validator = nil

		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to Text-to-Speech"), func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to OSC (VRChat)"), func() {

			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "send_osc",
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to Both (TTS + OSC)"), func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessageTts := SendMessageChannel.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessageTts.SendMessage()
			sendMessageOsc := SendMessageChannel.SendMessageStruct{
				Type: "send_osc",
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessageOsc.SendMessage()
		}))
		return entry
	},

	TextTranslateEnabled: func() *widget.Check {
		return widget.NewCheckWithData(lang.L("SttTextTranslateLabel", map[string]interface{}{"FromLang": "?", "ToLang": "?"})+AdditionalLanguagesCountString(" ", "[]"), DataBindings.TextTranslateEnabledDataBinding)
	},
	SttEnabled: func() *widget.Check {
		return widget.NewCheckWithData(lang.L("Speech-to-Text Enabled"), DataBindings.SpeechToTextEnabledDataBinding)
	},
	TtsEnabled: func(dataBinding binding.Bool) *widget.Check {
		return widget.NewCheckWithData(lang.L("Automatic Text-to-Speech"), dataBinding)
	},
	OscEnabled: func() *widget.Check {
		return widget.NewCheckWithData(lang.L("Automatic OSC (VRChat)"), DataBindings.OSCEnabledDataBinding)
	},
}

var Field = struct {
	RealtimeResultLabel                             *widget.Label // only displayed if realtime is enabled
	ProcessingStatus                                *widget.Activity
	WhisperResultList                               *widget.List
	TranscriptionTaskCombo                          *CustomWidget.TextValueSelect
	TranscriptionSpeakerLanguageCombo               *CustomWidget.CompletionEntry
	TranscriptionTargetLanguageCombo                *CustomWidget.CompletionEntry
	TranscriptionSpeechToTextInput                  *CustomWidget.EntryWithPopupMenu // Transcription (spoken source) textarea
	TranscriptionTextTranslationInput               *CustomWidget.EntryWithPopupMenu
	TranscriptionTextToSpeechInput                  *CustomWidget.EntryWithPopupMenu
	TranscriptionOcrInput                           *CustomWidget.EntryWithPopupMenu
	TranscriptionInputHint                          *canvas.Text
	TranscriptionInputHintOnTxtTranslate            *canvas.Text
	TranscriptionTranslationSpeechToTextInput       *CustomWidget.EntryWithPopupMenu // Translation Result textarea
	TranscriptionTranslationTextTranslationInput    *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationTextToSpeechInput       *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationOcrInput                *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationInputHint               *canvas.Text
	TranscriptionTranslationInputHintOnTxtTranslate *canvas.Text
	SourceLanguageCombo                             *CustomWidget.CompletionEntry
	TargetLanguageCombo                             *CustomWidget.CompletionEntry
	SourceLanguageTxtTranslateCombo                 *CustomWidget.CompletionEntry // used in OCR tab
	TargetLanguageTxtTranslateCombo                 *CustomWidget.CompletionEntry // used in OCR tab
	TtsModelCombo                                   *widget.Select
	TtsVoiceCombo                                   *widget.Select
	TextTranslateEnabled                            *widget.Check
	SttEnabled                                      *widget.Check
	TtsEnabledOnStt                                 *widget.Check
	TtsEnabledOnTxtTranslate                        *widget.Check
	OscEnabled                                      *widget.Check
	OscLimitHint                                    *canvas.Text
	OcrLanguageCombo                                *CustomWidget.CompletionEntry
	OcrWindowCombo                                  *CustomWidget.TappableSelect
	OcrImageContainer                               *fyne.Container
	LogText                                         *CustomWidget.LogText
	StatusBar                                       *widget.ProgressBar
	StatusText                                      *widget.Label
	StatusRow                                       *fyne.Container
}{}

func createFields() {
	Field.TranscriptionTaskCombo = fieldCreationFunctions.TranscriptionTaskCombo()
	Field.TranscriptionSpeechToTextInput = fieldCreationFunctions.TranscriptionInput(DataBindings.TranscriptionInputBinding)
	Field.TranscriptionTextTranslationInput = fieldCreationFunctions.TranscriptionInput(DataBindings.TranscriptionInputBinding)
	Field.TranscriptionTextToSpeechInput = fieldCreationFunctions.TranscriptionInput(DataBindings.TranscriptionInputBinding)
	Field.TranscriptionOcrInput = fieldCreationFunctions.TranscriptionInput(DataBindings.TranscriptionInputBinding)
	Field.TranscriptionTranslationSpeechToTextInput = fieldCreationFunctions.TranscriptionTranslationInput(DataBindings.TranscriptionTranslationInputBinding)
	Field.TranscriptionTranslationTextTranslationInput = fieldCreationFunctions.TranscriptionTranslationInput(DataBindings.TranscriptionTranslationInputBinding)
	Field.TranscriptionTranslationTextToSpeechInput = fieldCreationFunctions.TranscriptionTranslationInput(DataBindings.TranscriptionTranslationInputBinding)
	Field.TranscriptionTranslationOcrInput = fieldCreationFunctions.TranscriptionTranslationInput(DataBindings.TranscriptionTranslationInputBinding)
	Field.TextTranslateEnabled = fieldCreationFunctions.TextTranslateEnabled()
	Field.SttEnabled = fieldCreationFunctions.SttEnabled()
	Field.TtsEnabledOnStt = fieldCreationFunctions.TtsEnabled(DataBindings.TextToSpeechEnabledDataBinding)
	Field.TtsEnabledOnTxtTranslate = fieldCreationFunctions.TtsEnabled(DataBindings.TextToSpeechEnabledDataBinding)
	Field.OscEnabled = fieldCreationFunctions.OscEnabled()
}

func InitializeGlobalFields() {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Fields\\GlobalFields->InitializeGlobalFields")
	})

	// Initialize basic fields
	Field.RealtimeResultLabel = widget.NewLabelWithData(DataBindings.WhisperResultIntermediateResult)
	Field.ProcessingStatus = widget.NewActivity()
	Field.TranscriptionSpeakerLanguageCombo = CustomWidget.NewCompletionEntry([]string{"Auto"})
	Field.TranscriptionTargetLanguageCombo = CustomWidget.NewCompletionEntry([]string{})
	Field.TranscriptionInputHint = canvas.NewText("0", color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff})
	Field.TranscriptionInputHintOnTxtTranslate = canvas.NewText("0", color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff})
	Field.TranscriptionTranslationInputHint = canvas.NewText("0", color.NRGBA{R: 0xb2, G: 0xb2, B: 0xff})
	Field.TranscriptionTranslationInputHintOnTxtTranslate = canvas.NewText("0", color.NRGBA{R: 0xb2, G: 0xb2, B: 0xff})
	Field.SourceLanguageCombo = CustomWidget.NewCompletionEntry([]string{"Auto"})
	Field.TargetLanguageCombo = CustomWidget.NewCompletionEntry([]string{"None"})
	Field.SourceLanguageTxtTranslateCombo = CustomWidget.NewCompletionEntry([]string{"Auto"})
	Field.TargetLanguageTxtTranslateCombo = CustomWidget.NewCompletionEntry([]string{"None"})

	Field.TtsModelCombo = widget.NewSelect([]string{}, func(value string) {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_model",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	})

	Field.TtsVoiceCombo = widget.NewSelect([]string{}, func(value string) {

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_voice",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	})

	Field.OscLimitHint = canvas.NewText(fmt.Sprintf(OscLimitLabelConst, 0, 0), color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff})

	Field.OcrLanguageCombo = CustomWidget.NewCompletionEntry([]string{})

	Field.OcrWindowCombo = CustomWidget.NewSelect([]string{}, func(value string) {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_window_name",
			Value: value,
		}
		sendMessage.SendMessage()
	})

	Field.OcrImageContainer = container.NewStack()
	Field.LogText = CustomWidget.NewLogTextWithData(DataBindings.LogBinding)
	Field.LogText.AutoScroll = true
	Field.LogText.ReadOnly = true
	Field.StatusText = widget.NewLabelWithData(DataBindings.StatusTextBinding)

	createFields()

	Field.RealtimeResultLabel.Wrapping = fyne.TextWrapWord
	Field.RealtimeResultLabel.TextStyle = fyne.TextStyle{Italic: true}

	Field.SourceLanguageCombo.OptionsTextValue = []CustomWidget.TextValueOption{{
		Text:  "Auto",
		Value: "auto",
	}}
	Field.SourceLanguageCombo.ShowAllEntryText = lang.L("... show all")
	Field.SourceLanguageCombo.Entry.PlaceHolder = lang.L("Select source language")
	Field.SourceLanguageCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.SourceLanguageCombo.Options); i++ {
			if len(Field.SourceLanguageCombo.Options) > i && strings.Contains(strings.ToLower(Field.SourceLanguageCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.SourceLanguageCombo.Options[i])
			}
		}

		Field.SourceLanguageCombo.SetOptionsFilter(filteredValues)
		Field.SourceLanguageCombo.ShowCompletion()
	}
	Field.SourceLanguageCombo.OnSubmitted = func(value string) {
		value = Field.SourceLanguageCombo.UpdateCompletionEntryBasedOnValue(value)

		valueObj := Field.SourceLanguageCombo.GetValueOptionEntryByText(value)

		if strings.ToLower(value) == "auto" {
			value = ""
		}

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "src_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		Field.TextTranslateEnabled.Text = lang.L("SttTextTranslateLabel", map[string]interface{}{"FromLang": valueObj.Text, "ToLang": Field.TargetLanguageCombo.Text}) + AdditionalLanguagesCountString(" ", "[]")
		Field.TextTranslateEnabled.Refresh()

		log.Println("Select set to", value)
	}

	Field.TargetLanguageCombo.ShowAllEntryText = lang.L("... show all")
	Field.TargetLanguageCombo.Entry.PlaceHolder = lang.L("Select target language")
	Field.TargetLanguageCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.TargetLanguageCombo.Options); i++ {
			if len(Field.TargetLanguageCombo.Options) > i && strings.Contains(strings.ToLower(Field.TargetLanguageCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.TargetLanguageCombo.Options[i])
			}
		}

		Field.TargetLanguageCombo.SetOptionsFilter(filteredValues)
		Field.TargetLanguageCombo.ShowCompletion()
	}
	Field.TargetLanguageCombo.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Field.TargetLanguageCombo.UpdateCompletionEntryBasedOnValue(value)
		valueObj := Field.TargetLanguageCombo.GetValueOptionEntryByText(value)

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "trg_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		Field.TextTranslateEnabled.Text = lang.L("SttTextTranslateLabel", map[string]interface{}{"FromLang": Field.SourceLanguageCombo.Text, "ToLang": valueObj.Text}) + AdditionalLanguagesCountString(" ", "[]")
		Field.TextTranslateEnabled.Refresh()

		log.Println("Select set to", value)
	}

	Field.SourceLanguageTxtTranslateCombo.ShowAllEntryText = lang.L("... show all")
	Field.SourceLanguageTxtTranslateCombo.Entry.PlaceHolder = lang.L("Select source language")
	Field.SourceLanguageTxtTranslateCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.SourceLanguageTxtTranslateCombo.Options); i++ {
			if len(Field.SourceLanguageTxtTranslateCombo.Options) > i && strings.Contains(strings.ToLower(Field.SourceLanguageTxtTranslateCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.SourceLanguageTxtTranslateCombo.Options[i])
			}
		}

		Field.SourceLanguageTxtTranslateCombo.SetOptionsFilter(filteredValues)
		Field.SourceLanguageTxtTranslateCombo.ShowCompletion()
	}
	Field.SourceLanguageTxtTranslateCombo.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Field.SourceLanguageTxtTranslateCombo.UpdateCompletionEntryBasedOnValue(value)

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_txt_src_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}

	Field.TargetLanguageTxtTranslateCombo.ShowAllEntryText = lang.L("... show all")
	Field.TargetLanguageTxtTranslateCombo.Entry.PlaceHolder = lang.L("Select target language")
	Field.TargetLanguageTxtTranslateCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.TargetLanguageTxtTranslateCombo.Options); i++ {
			if len(Field.TargetLanguageTxtTranslateCombo.Options) > i && strings.Contains(strings.ToLower(Field.TargetLanguageTxtTranslateCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.TargetLanguageTxtTranslateCombo.Options[i])
			}
		}
		Field.TargetLanguageTxtTranslateCombo.SetOptionsFilter(filteredValues)
		Field.TargetLanguageTxtTranslateCombo.ShowCompletion()
	}
	Field.TargetLanguageTxtTranslateCombo.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Field.TargetLanguageTxtTranslateCombo.UpdateCompletionEntryBasedOnValue(value)

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_txt_trg_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}

	Field.TranscriptionSpeakerLanguageCombo.ShowAllEntryText = lang.L("... show all")
	Field.TranscriptionSpeakerLanguageCombo.Entry.PlaceHolder = lang.L("Select a language")
	Field.TranscriptionSpeakerLanguageCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.TranscriptionSpeakerLanguageCombo.Options); i++ {
			if len(Field.TranscriptionSpeakerLanguageCombo.Options) > i && strings.Contains(strings.ToLower(Field.TranscriptionSpeakerLanguageCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.TranscriptionSpeakerLanguageCombo.Options[i])
			}
		}

		Field.TranscriptionSpeakerLanguageCombo.SetOptionsFilter(filteredValues)
		Field.TranscriptionSpeakerLanguageCombo.ShowCompletion()
	}
	Field.TranscriptionSpeakerLanguageCombo.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Field.TranscriptionSpeakerLanguageCombo.UpdateCompletionEntryBasedOnValue(value)

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "current_language",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.TranscriptionTargetLanguageCombo.ShowAllEntryText = lang.L("... show all")
	Field.TranscriptionTargetLanguageCombo.Entry.PlaceHolder = lang.L("Select a language")
	Field.TranscriptionTargetLanguageCombo.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.TranscriptionTargetLanguageCombo.Options); i++ {
			if len(Field.TranscriptionTargetLanguageCombo.Options) > i && strings.Contains(strings.ToLower(Field.TranscriptionTargetLanguageCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.TranscriptionTargetLanguageCombo.Options[i])
			}
		}

		Field.TranscriptionTargetLanguageCombo.SetOptionsFilter(filteredValues)
		Field.TranscriptionTargetLanguageCombo.ShowCompletion()
	}
	Field.TranscriptionTargetLanguageCombo.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Field.TranscriptionTargetLanguageCombo.UpdateCompletionEntryBasedOnValue(value)

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "target_language",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.TextTranslateEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "txt_translate",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.SttEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "stt_enabled",
			Value: value,
		}
		sendMessage.SendMessage()
	}
	Field.TtsEnabledOnStt.OnChanged = func(value bool) {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_answer",
			Value: value,
		}
		sendMessage.SendMessage()
	}
	Field.OscEnabled.OnChanged = func(value bool) {
		if value {
			Field.OscLimitHint.Show()
		} else {
			Field.OscLimitHint.Hide()
		}
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "osc_auto_processing_enabled",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.OcrLanguageCombo.ShowAllEntryText = lang.L("... show all")
	Field.OcrLanguageCombo.Entry.PlaceHolder = lang.L("Select language in image")
	Field.OcrLanguageCombo.OnChanged = func(value string) {
		// filter out the values of Field.TargetLanguageTxtTranslateCombo.Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(Field.OcrLanguageCombo.Options); i++ {
			if len(Field.OcrLanguageCombo.Options) > i && strings.Contains(strings.ToLower(Field.OcrLanguageCombo.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, Field.OcrLanguageCombo.Options[i])
			}
		}
		Field.OcrLanguageCombo.SetOptionsFilter(filteredValues)
		Field.OcrLanguageCombo.ShowCompletion()
	}
	Field.OcrWindowCombo.UpdateBeforeOpenFunc = func() {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type: "get_windows_list",
		}
		sendMessage.SendMessage()
	}

	Field.OscLimitHint.TextSize = theme.TextSize()

	Field.TranscriptionInputHint.TextSize = theme.CaptionTextSize()
	Field.TranscriptionInputHint.Alignment = fyne.TextAlignLeading
	Field.TranscriptionInputHintOnTxtTranslate.TextSize = theme.CaptionTextSize()
	Field.TranscriptionInputHintOnTxtTranslate.Alignment = fyne.TextAlignLeading
	Field.TranscriptionSpeechToTextInput.OnChanged = func(value string) {
		Field.TranscriptionInputHint.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionInputHint.Refresh()
		Field.TranscriptionInputHintOnTxtTranslate.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionInputHintOnTxtTranslate.Refresh()

		OscLimitHintUpdateFunc()
	}
	Field.TranscriptionTranslationInputHint.TextSize = theme.CaptionTextSize()
	Field.TranscriptionTranslationInputHint.Alignment = fyne.TextAlignLeading
	Field.TranscriptionTranslationInputHintOnTxtTranslate.TextSize = theme.CaptionTextSize()
	Field.TranscriptionTranslationInputHintOnTxtTranslate.Alignment = fyne.TextAlignLeading
	Field.TranscriptionTranslationSpeechToTextInput.OnChanged = func(value string) {
		Field.TranscriptionTranslationInputHint.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionTranslationInputHint.Refresh()
		Field.TranscriptionTranslationInputHintOnTxtTranslate.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionTranslationInputHintOnTxtTranslate.Refresh()

		OscLimitHintUpdateFunc()
	}

}
