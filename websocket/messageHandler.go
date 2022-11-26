package websocket

import (
	"encoding/json"
	"log"
	"whispering-tiger-ui/websocket/Messages"
)

func (c *MessageStruct) HandleMessage() {
	var err error = nil

	switch c.Type {
	case "installed_languages":
		err = json.Unmarshal(c.Data, &Messages.InstalledLanguages)
	case "available_tts_models":
		err = json.Unmarshal(c.Data, &Messages.TtsLanguages)
	case "available_tts_voices":
		err = json.Unmarshal(c.Data, &Messages.TtsVoices)
	case "translate_settings":
		err = json.Unmarshal(c.Data, &Messages.TranslateSettings)
	case "transcript":
		Messages.WhisperResults = append([]Messages.WhisperResult{
			{
				Text:     c.Text,
				Language: c.Language,
			},
		}, Messages.WhisperResults...)
		Messages.WhisperResultsDataBinding.Prepend(c.Text)
	case "windows_list":
		err = json.Unmarshal(c.Data, &Messages.WindowsList)
	case "translate_result":
		Messages.LastTranslationResult = c.TranslateResult
	}
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
