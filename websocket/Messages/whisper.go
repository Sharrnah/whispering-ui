package Messages

import (
	"encoding/json"
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
	Fields.Field.TranscriptionInput.SetText(res.Text)
	Fields.Field.TranscriptionTranslation.SetText(res.TxtTranslation)

	WhisperResults = append([]WhisperResult{res}, WhisperResults...)

	jsonBytes, _ := json.Marshal(res)
	jsonResult := string(jsonBytes[:])

	//whisperResult := strings.Join([]string{c.TxtTranslation, c.Text}, "###")
	Fields.DataBindings.WhisperResultsDataBinding.Prepend(jsonResult)
}

var WhisperResults []WhisperResult

var LastTranslationResult string
