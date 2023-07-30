package Fields

import (
	"fyne.io/fyne/v2/data/binding"
)

var DataBindings = struct {
	WhisperResultsDataBinding       binding.ExternalUntypedList
	SpeechToTextEnabledDataBinding  binding.Bool
	TextTranslateEnabledDataBinding binding.Bool
	TextToSpeechEnabledDataBinding  binding.Bool
	OSCEnabledDataBinding           binding.Bool
}{
	WhisperResultsDataBinding: binding.BindUntypedList(
		&[]interface{}{},
	),
	SpeechToTextEnabledDataBinding:  binding.NewBool(),
	TextTranslateEnabledDataBinding: binding.NewBool(),
	TextToSpeechEnabledDataBinding:  binding.NewBool(),
	OSCEnabledDataBinding:           binding.NewBool(),
}
