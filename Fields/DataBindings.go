package Fields

import (
	"fyne.io/fyne/v2/data/binding"
)

var DataBindings = struct {
	WhisperResultsDataBinding      binding.ExternalStringList
	TextToSpeechEnabledDataBinding binding.Bool
	OSCEnabledDataBinding          binding.Bool
}{
	WhisperResultsDataBinding: binding.BindStringList(
		&[]string{},
	),
	TextToSpeechEnabledDataBinding: binding.NewBool(),
	OSCEnabledDataBinding:          binding.NewBool(),
}
