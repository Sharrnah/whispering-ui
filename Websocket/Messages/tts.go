package Messages

import (
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

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
	Fields.Field.TtsModelCombo.Options = nil
	for _, languageItem := range res.Languages {
		//elementName := languageItem.Language
		for _, modelItem := range languageItem.Models {
			if strings.Contains(modelItem, "v3") {
				Fields.Field.TtsModelCombo.Options = append(Fields.Field.TtsModelCombo.Options, modelItem)
			}

		}
	}
	return &res
}

// TTS Voices

type TtsVoicesListing struct {
	Voices []string `json:"data"`
}

var TtsVoices TtsVoicesListing

func (res TtsVoicesListing) Update() *TtsVoicesListing {
	Fields.Field.TtsVoiceCombo.Options = nil
	Fields.Field.TtsVoiceCombo.Options = append(Fields.Field.TtsVoiceCombo.Options, res.Voices...)

	// set first voice if selection is not in list
	voicesListContainsSelectedVoice := false
	for _, voice := range res.Voices {
		if voice == Settings.Config.Tts_voice {
			voicesListContainsSelectedVoice = true
			break
		}
	}
	if !voicesListContainsSelectedVoice {
		Fields.Field.TtsVoiceCombo.SetSelectedIndex(0)
	}

	return &res
}
