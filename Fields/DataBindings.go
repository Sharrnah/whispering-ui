package Fields

import (
	"fyne.io/fyne/v2/data/binding"
)

var DataBindings = struct {
	WhisperResultsDataBinding       binding.ExternalUntypedList
	WhisperResultIntermediateResult binding.String
	SpeechToTextEnabledDataBinding  binding.Bool
	TextTranslateEnabledDataBinding binding.Bool
	TextToSpeechEnabledDataBinding  binding.Bool
	OSCEnabledDataBinding           binding.Bool
	StatusTextBinding               binding.String
}{
	WhisperResultsDataBinding: binding.BindUntypedList(
		&[]interface{}{},
	),
	WhisperResultIntermediateResult: binding.NewString(),
	SpeechToTextEnabledDataBinding:  binding.NewBool(),
	TextTranslateEnabledDataBinding: binding.NewBool(),
	TextToSpeechEnabledDataBinding:  binding.NewBool(),
	OSCEnabledDataBinding:           binding.NewBool(),
	StatusTextBinding:               binding.NewString(),
}
