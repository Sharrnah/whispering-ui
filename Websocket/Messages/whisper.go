package Messages

import (
	"fyne.io/fyne/v2"
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
	FieldsWhisperResultData := Fields.WhisperResult{
		Text:                 res.Text,
		Language:             res.Language,
		TxtTranslation:       res.TxtTranslation,
		TxtTranslationTarget: res.TxtTranslationTarget,
	}

	// prepend to slice Fields.DataBindings.WhisperResultsData

	fyne.Do(func() {
		Fields.DataBindings.WhisperResultsData = append([]Fields.WhisperResult{FieldsWhisperResultData}, Fields.DataBindings.WhisperResultsData...)
		Fields.Field.WhisperResultList.Refresh()
	})
}
