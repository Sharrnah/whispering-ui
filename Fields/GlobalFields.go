package Fields

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
	"log"
	"whispering-tiger-ui/CustomWidget"
)

const SttTextTranslateLabelConst = "Automatic Text-Translate from %s to %s"
const OscLimitLabelConst = "[%d / %d]"

var OscLimitHintUpdateFunc = func() {}

var Field = struct {
	RealtimeResultLabel               *widget.Label // only displayed if realtime is enabled
	ProcessingStatus                  *widget.ProgressBarInfinite
	WhisperResultList                 *widget.List
	TranscriptionTaskCombo            *widget.Select
	TranscriptionSpeakerLanguageCombo *widget.Select
	TranscriptionInput                *CustomWidget.EntryWithPopupMenu
	TranscriptionInputHint            *canvas.Text
	TranscriptionTranslationInput     *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationInputHint *canvas.Text
	SourceLanguageCombo               *CustomWidget.TextValueSelect
	TargetLanguageCombo               *widget.Select
	TargetLanguageTxtTranslateCombo   *widget.Select
	TtsModelCombo                     *widget.Select
	TtsVoiceCombo                     *widget.Select
	TextTranslateEnabled              *widget.Check
	SttEnabled                        *widget.Check
	TtsEnabled                        *widget.Check
	OscEnabled                        *widget.Check
	OscLimitHint                      *canvas.Text
	OcrLanguageCombo                  *widget.Select
	OcrWindowCombo                    *CustomWidget.TappableSelect
	OcrImageContainer                 *fyne.Container
	LogText                           *terminal.Terminal
}{
	RealtimeResultLabel: widget.NewLabel(""),
	ProcessingStatus:    nil,
	WhisperResultList:   nil,
	TranscriptionTaskCombo: widget.NewSelect([]string{"transcribe", "translate (to English)"}, func(value string) {
		switch value {
		case "transcribe":
			value = "transcribe"
		case "translate (to English)":
			value = "translate"
		}
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "whisper_task",
			Value: value,
		}
		sendMessage.SendMessage()
	}),
	TranscriptionSpeakerLanguageCombo: widget.NewSelect([]string{"Auto"}, func(value string) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "current_language",
			Value: value,
		}
		sendMessage.SendMessage()
	}),
	TranscriptionInput: func() *CustomWidget.EntryWithPopupMenu {
		entry := CustomWidget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord

		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to Text-to-Speech", func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessage := SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to OSC (VRChat)", func() {
			sendMessage := SendMessageStruct{
				Type: "send_osc",
				//Value: &entry.Text,
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessage.SendMessage()
		}))
		return entry
	}(),
	TranscriptionInputHint: canvas.NewText("0", theme.PlaceHolderColor()),
	TranscriptionTranslationInput: func() *CustomWidget.EntryWithPopupMenu {
		entry := CustomWidget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord

		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to Text-to-Speech", func() {
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     entry.Text,
				ToDevice: true,
				Download: false,
			}
			sendMessage := SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to OSC (VRChat)", func() {

			sendMessage := SendMessageStruct{
				Type: "send_osc",
				//Value: &entry.Text,
				Value: struct {
					Text *string `json:"text"`
				}{
					Text: &entry.Text,
				},
			}
			sendMessage.SendMessage()
		}))
		return entry
	}(),
	TranscriptionTranslationInputHint: canvas.NewText("0", theme.PlaceHolderColor()),
	SourceLanguageCombo: CustomWidget.NewTextValueSelect("src_lang", []CustomWidget.TextValueOption{
		{
			Text:  "Auto",
			Value: "auto",
		},
	}, func(valueObj CustomWidget.TextValueOption) {}, 0),
	TargetLanguageCombo: widget.NewSelect([]string{"None"}, func(value string) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "trg_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}),
	TargetLanguageTxtTranslateCombo: widget.NewSelect([]string{"None"}, func(value string) {}),
	TtsModelCombo: widget.NewSelect([]string{}, func(value string) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_model",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}),
	TtsVoiceCombo: widget.NewSelect([]string{}, func(value string) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_voice",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}),
	TextTranslateEnabled: widget.NewCheckWithData(fmt.Sprintf(SttTextTranslateLabelConst, "?", "?"), DataBindings.TextTranslateEnabledDataBinding),
	SttEnabled:           widget.NewCheckWithData("Speech-to-Text Enabled", DataBindings.SpeechToTextEnabledDataBinding),
	TtsEnabled:           widget.NewCheckWithData("Automatic Text-to-Speech", DataBindings.TextToSpeechEnabledDataBinding),
	OscEnabled:           widget.NewCheckWithData("Automatic OSC (VRChat)", DataBindings.OSCEnabledDataBinding),
	OscLimitHint:         canvas.NewText(fmt.Sprintf(OscLimitLabelConst, 0, 0), theme.PlaceHolderColor()),

	OcrLanguageCombo: widget.NewSelect([]string{}, func(value string) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_lang",
			Value: value,
		}
		sendMessage.SendMessage()
	}),
	OcrWindowCombo: CustomWidget.NewSelect([]string{}, func(value string) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_window_name",
			Value: value,
		}
		sendMessage.SendMessage()
	}),
	OcrImageContainer: container.NewMax(),
	LogText:           terminal.New(),
}

func init() {
	Field.RealtimeResultLabel.Wrapping = fyne.TextWrapWord
	Field.RealtimeResultLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Set onchange events
	Field.SourceLanguageCombo.OnChanged = func(valueObj CustomWidget.TextValueOption) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "src_lang",
			Value: valueObj.Value,
		}
		sendMessage.SendMessage()

		Field.TextTranslateEnabled.Text = fmt.Sprintf(SttTextTranslateLabelConst, valueObj.Value, Field.TargetLanguageCombo.Selected)
		Field.TextTranslateEnabled.Refresh()

		log.Println("Select set to", valueObj.Value)
	}
	Field.TargetLanguageCombo.OnChanged = func(value string) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "trg_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		Field.TextTranslateEnabled.Text = fmt.Sprintf(SttTextTranslateLabelConst, Field.SourceLanguageCombo.GetSelected().Text, value)
		Field.TextTranslateEnabled.Refresh()

		log.Println("Select set to", value)
	}

	Field.TextTranslateEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "txt_translate",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.SttEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "stt_enabled",
			Value: value,
		}
		sendMessage.SendMessage()
	}
	Field.TtsEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageStruct{
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
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "osc_auto_processing_enabled",
			Value: value,
		}
		sendMessage.SendMessage()
	}

	Field.OcrWindowCombo.UpdateBeforeOpenFunc = func() {
		sendMessage := SendMessageStruct{
			Type: "get_windows_list",
		}
		sendMessage.SendMessage()
	}

	Field.OscLimitHint.TextSize = theme.TextSize()

	Field.TranscriptionInputHint.TextSize = theme.CaptionTextSize()
	Field.TranscriptionInputHint.Alignment = fyne.TextAlignLeading
	Field.TranscriptionInput.OnChanged = func(value string) {
		Field.TranscriptionInputHint.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionInputHint.Refresh()

		OscLimitHintUpdateFunc()
	}
	Field.TranscriptionTranslationInputHint.TextSize = theme.CaptionTextSize()
	Field.TranscriptionTranslationInputHint.Alignment = fyne.TextAlignLeading
	Field.TranscriptionTranslationInput.OnChanged = func(value string) {
		Field.TranscriptionTranslationInputHint.Text = fmt.Sprintf("%d", len([]rune(value)))
		Field.TranscriptionTranslationInputHint.Refresh()

		OscLimitHintUpdateFunc()
	}

}
