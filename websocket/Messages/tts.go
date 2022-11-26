package Messages

type TtsLanguage struct {
	Language string   `json:"language"`
	Models   []string `json:"models"`
}

var TtsLanguages []TtsLanguage

var TtsVoices []string
