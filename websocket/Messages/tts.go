package Messages

// TTS Languages

type TtsLanguage struct {
	Language string   `json:"language"`
	Models   []string `json:"models"`
}

type TtsLanguagesListing struct {
	Languages []TtsLanguage `json:"data"`
}

var TtsLanguages TtsLanguagesListing

func (res TtsLanguagesListing) Update() *TtsLanguagesListing {
	return &res
}

// TTS Voices

type TtsVoicesListing struct {
	Voices []string `json:"data"`
}

var TtsVoices TtsVoicesListing

func (res TtsVoicesListing) Update() *TtsVoicesListing {
	return &res
}
