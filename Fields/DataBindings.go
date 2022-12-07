package Fields

import (
	"fyne.io/fyne/v2/data/binding"
)

var DataBindings = struct {
	WhisperResultsDataBinding      binding.ExternalUntypedList
	TextToSpeechEnabledDataBinding binding.Bool
	OSCEnabledDataBinding          binding.Bool
}{
	WhisperResultsDataBinding: binding.BindUntypedList(
		&[]interface{}{},
	),
	TextToSpeechEnabledDataBinding: binding.NewBool(),
	OSCEnabledDataBinding:          binding.NewBool(),
}
