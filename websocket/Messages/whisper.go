package Messages

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
	//
}

var WhisperResults []WhisperResult

var LastTranslationResult string
