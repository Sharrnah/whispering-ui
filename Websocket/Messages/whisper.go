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

	//if Fields.DataBindings.WhisperResultsDataBinding.Length() >= 200 {
	//	whisperResultsPart, _ := Fields.DataBindings.WhisperResultsDataBinding.Get()
	//	Fields.DataBindings.WhisperResultsDataBinding.Set(whisperResultsPart[:199])
	//}

	Fields.DataBindings.WhisperResultsDataBinding.Prepend(res)

	//Fields.Field.WhisperResultList.ScrollToTop()
	//Fields.DataBindings.WhisperResultsDataBinding.Append(res)
	//Fields.Field.WhisperResultList.ScrollToBottom()

	//Fields.DataBindings.WhisperResultsDataBinding.Reload()

	//Fields.DataBindings.WhisperResultsDataBinding.Reload()

	//Fields.Field.WhisperResultList.Refresh()

	//go func() {
	//	// wait for 100 milliseconds
	//	time.Sleep(100 * time.Millisecond)
	//
	//	// Force reload of the data binding
	//	Fields.DataBindings.WhisperResultsDataBinding.Reload()
	//}()

	WhisperResults = append([]WhisperResult{res}, WhisperResults...)

	// resize new entry
	//txtTranslationSize := fyne.MeasureText(res.TxtTranslation, 12, fyne.TextStyle{Bold: true})
	//textSize := fyne.MeasureText(res.Text, 12, fyne.TextStyle{Bold: true})
	//Fields.Field.WhisperResultList.SetItemHeight(Fields.Field.WhisperResultList.Length()-1, txtTranslationSize.Height+textSize.Height)
}

var WhisperResults []WhisperResult

var LastTranslationResult string
