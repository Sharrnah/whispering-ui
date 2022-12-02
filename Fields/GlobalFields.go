package Fields

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"log"
)

var Field = struct {
	TranscriptionTaskCombo            *widget.Select
	TranscriptionSpeakerLanguageCombo *widget.Select
	TranscriptionInput                *widget.Entry
	TranscriptionTranslationInput     *widget.Entry
	SourceLanguageCombo               *widget.Select
	TargetLanguageCombo               *widget.Select
	TtsEnabled                        *widget.Check
	OscEnabled                        *widget.Check
}{
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
	TranscriptionInput: func() *widget.Entry {
		entry := widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		return entry
	}(),
	TranscriptionTranslationInput: func() *widget.Entry {
		entry := widget.NewMultiLineEntry()
		entry.Wrapping = fyne.TextWrapWord
		return entry
	}(),
	SourceLanguageCombo: widget.NewSelect([]string{"Auto"}, func(value string) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "src_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}),
	TargetLanguageCombo: widget.NewSelect([]string{"None"}, func(value string) {

		sendMessage := SendMessageStruct{
			Type:  "setting_change",
			Name:  "trg_lang",
			Value: value,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}),
	TtsEnabled: widget.NewCheckWithData("Text 2 Speech", DataBindings.TextToSpeechEnabledDataBinding),
	OscEnabled: widget.NewCheckWithData("OSC", DataBindings.OSCEnabledDataBinding),
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
}
