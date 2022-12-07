package Messages

import (
	"whispering-tiger-ui/Fields"
)

type WhisperResult struct {
	Text                 string `json:"text"`
	Language             string `json:"language"`
	TxtTranslation       string `json:"txt_translation,omitempty"`
	TxtTranslationTarget string `json:"txt_translation_target,omitempty"`
}

func (res WhisperResult) String() string {
	return res.Text
}
func (res WhisperResult) Update() {
	//Fields.Field.TranscriptionInput.SetText(res.Text)
	//Fields.Field.TranscriptionTranslationInput.SetText(res.TxtTranslation)

	WhisperResults = append([]WhisperResult{res}, WhisperResults...)

	//if Fields.DataBindings.WhisperResultsDataBinding.Length() >= 200 {
	//	whisperResultsPart, _ := Fields.DataBindings.WhisperResultsDataBinding.Get()
	//	Fields.DataBindings.WhisperResultsDataBinding.Set(whisperResultsPart[:199])
	//}
	Fields.DataBindings.WhisperResultsDataBinding.Prepend(res)
}

var WhisperResults []WhisperResult

var LastTranslationResult string
