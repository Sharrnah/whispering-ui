package Fields

import (
	"fyne.io/fyne/v2/widget"
	"log"
)

var Field = struct {
	TargetLanguageCombo *widget.Select
	TtsEnabled          *widget.Check
	OscEnabled          *widget.Check
}{
	TargetLanguageCombo: widget.NewSelect([]string{"Option 1", "Option 2"}, func(value string) {

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
