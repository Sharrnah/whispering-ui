package Fields

import (
	"fyne.io/fyne/v2/data/binding"
)

type WhisperResult struct {
	Text                 string `json:"text"`
	Language             string `json:"language"`
	TxtTranslation       string `json:"txt_translation,omitempty"`
	TxtTranslationTarget string `json:"txt_translation_target,omitempty"`
}

var DataBindings = struct {
	WhisperResultsData                   []WhisperResult
	WhisperResultIntermediateResult      binding.String
	SpeechToTextEnabledDataBinding       binding.Bool
	TextTranslateEnabledDataBinding      binding.Bool
	TextToSpeechEnabledDataBinding       binding.Bool
	OSCEnabledDataBinding                binding.Bool
	StatusTextBinding                    binding.String
	TranscriptionInputBinding            binding.String
	TranscriptionTranslationInputBinding binding.String
}{
	WhisperResultIntermediateResult:      binding.NewString(),
	SpeechToTextEnabledDataBinding:       binding.NewBool(),
	TextTranslateEnabledDataBinding:      binding.NewBool(),
	TextToSpeechEnabledDataBinding:       binding.NewBool(),
	OSCEnabledDataBinding:                binding.NewBool(),
	StatusTextBinding:                    binding.NewString(),
	TranscriptionInputBinding:            binding.NewString(),
	TranscriptionTranslationInputBinding: binding.NewString(),
}
