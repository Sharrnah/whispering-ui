package Fields

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"log"
	"whispering-tiger-ui/CustomWidget"
)

var Field = struct {
	ProcessingStatus                  *widget.ProgressBarInfinite
	TranscriptionTaskCombo            *widget.Select
	TranscriptionSpeakerLanguageCombo *widget.Select
	TranscriptionInput                *CustomWidget.EntryWithPopupMenu
	TranscriptionTranslationInput     *CustomWidget.EntryWithPopupMenu
	SourceLanguageCombo               *CustomWidget.TextValueSelect
	TargetLanguageCombo               *widget.Select
	TargetLanguageTxtTranslateCombo   *widget.Select
	TtsModelCombo                     *widget.Select
	TtsVoiceCombo                     *widget.Select
	TtsEnabled                        *widget.Check
	OscEnabled                        *widget.Check
	OcrLanguageCombo                  *widget.Select
	OcrWindowCombo                    *CustomWidget.TappableSelect
}{
	ProcessingStatus: nil,
	TranscriptionTaskCombo: widget.NewSelect([]string{"transcribe", "translate (to en)"}, func(value string) {
		switch value {
		case "transcribe":
			value = "transcribe"
		case "translate (to en)":
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

		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to Text 2 Speech", func() {
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
		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to OSC", func() {
			sendMessage := SendMessageStruct{
				Type:  "send_osc",
				Value: &entry.Text,
			}
			sendMessage.SendMessage()
		}))
		return entry
	}(),
	TranscriptionTranslationInput: func() *CustomWidget.EntryWithPopupMenu {
		entry := CustomWidget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord

		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to Text 2 Speech", func() {
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
		entry.AddAdditionalMenuItem(fyne.NewMenuItem("Send to OSC", func() {
			sendMessage := SendMessageStruct{
				Type:  "send_osc",
				Value: &entry.Text,
			}
			sendMessage.SendMessage()
		}))
		return entry
	}(),
	SourceLanguageCombo: CustomWidget.NewTextValueSelect("src_lang", []CustomWidget.TextValueOption{
		{
			Text:  "Auto",
			Value: "auto",
		},
	}, func(valueObj CustomWidget.TextValueOption) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "src_lang",
			Value: valueObj.Value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", valueObj.Value)
	}, 0),
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
	TtsEnabled: widget.NewCheckWithData("Automatic Text 2 Speech", DataBindings.TextToSpeechEnabledDataBinding),
	OscEnabled: widget.NewCheckWithData("Automatic OSC", DataBindings.OSCEnabledDataBinding),

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
}

func init() {
	// Set onchange events
	Field.TtsEnabled.OnChanged = func(value bool) {
		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_answer",
			Value: value,
		}
		sendMessage.SendMessage()
	}
	Field.OscEnabled.OnChanged = func(value bool) {
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
}
