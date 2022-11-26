package Messages

import "fyne.io/fyne/v2/data/binding"

type WhisperResult struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

func (res WhisperResult) String() string {
	return res.Text
}

var WhisperResults []WhisperResult

var WhisperResultsDataBinding = binding.BindStringList(
	&[]string{},
)

var LastTranslationResult string
