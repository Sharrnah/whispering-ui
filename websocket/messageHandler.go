package websocket

import (
	"encoding/json"
	"log"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/websocket/Messages"
)

// receiving message

const (
	SkipMessage = 85746964687
)

type MessageStruct struct {
	Raw  []byte          // Raw data representation of this struct
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`

	// only in case of whisper message
	Text                 string `json:"text,omitempty"`
	Language             string `json:"language,omitempty"` // speaker language
	TxtTranslation       string `json:"txt_translation,omitempty"`
	TxtTranslationSource string `json:"txt_translation_source,omitempty"`
	TxtTranslationTarget string `json:"txt_translation_target,omitempty"`

	// only in case of text translate message
	TranslateResult string `json:"translate_result,omitempty"`
	OriginalText    string `json:"original_text,omitempty"`
	TxtFromLang     string `json:"txt_from_lang,omitempty"`

	// only in case of FLAN message
	FlanAnswer string `json:"flan_answer,omitempty"`
}

func (c *MessageStruct) GetMessage(messageData []byte) *MessageStruct {
	c.Raw = messageData
	return messageLoader(c, messageData).(*MessageStruct)
}

// Handle the different receiving message types

func (c *MessageStruct) HandleReceiveMessage() {
	var err error = nil

	switch c.Type {
	case "installed_languages":
		err = json.Unmarshal(c.Raw, &Messages.InstalledLanguages)
		Messages.InstalledLanguages.Update()
	case "available_tts_models":
		err = json.Unmarshal(c.Raw, &Messages.TtsLanguages)
		Messages.TtsLanguages.Update()
	case "available_tts_voices":
		err = json.Unmarshal(c.Raw, &Messages.TtsVoices)
		Messages.TtsVoices.Update()
	case "translate_settings":
		err = json.Unmarshal(c.Data, &Messages.TranslateSettings)
		Messages.TranslateSettings.Update()
	case "transcript":
		c.Text = strings.TrimSpace(c.Text)
		c.TxtTranslation = strings.TrimSpace(c.TxtTranslation)
		whisperResultMessage := Messages.WhisperResult{
			Text:                 c.Text,
			Language:             c.Language,
			TxtTranslation:       c.TxtTranslation,
			TxtTranslationTarget: c.TxtTranslationTarget,
		}

		whisperResultMessage.Update()
	case "windows_list":
		err = json.Unmarshal(c.Raw, &Messages.WindowsList)
		Messages.WindowsList.Update()
	case "translate_result":
		Messages.LastTranslationResult = c.TranslateResult
		Fields.Field.TranscriptionTranslationInput.SetText(c.TranslateResult)
	}
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

}

func HandleSendMessage(sendMessage *Fields.SendMessageStruct) {
	switch sendMessage.Type {
	case "setting_change":
		switch sendMessage.Name {
		case "src_lang":
			langCode := Messages.InstalledLanguages.GetCodeByName(sendMessage.Value.(string))
			if langCode == "" {
				langCode = "auto"
			}
			if langCode != "" && Messages.TranslateSettings.Src_lang != langCode {
				sendMessage.Value = langCode
			} else {
				sendMessage.Value = SkipMessage
			}
		case "trg_lang":
			langCode := Messages.InstalledLanguages.GetCodeByName(sendMessage.Value.(string))
			if langCode != "" && Messages.TranslateSettings.Trg_lang != langCode {
				sendMessage.Value = langCode
				txtTranslateSendMessage := Fields.SendMessageStruct{
					Type:  "setting_change",
					Name:  "txt_translate",
					Value: true,
				}
				go txtTranslateSendMessage.SendMessage()
			} else {
				sendMessage.Value = SkipMessage
			}
			if langCode == "" {
				txtTranslateSendMessage := Fields.SendMessageStruct{
					Type:  "setting_change",
					Name:  "txt_translate",
					Value: false,
				}
				go txtTranslateSendMessage.SendMessage()
			}

		case "current_language":
			langCode := Messages.TranslateSettings.GetWhisperLanguageCodeByName(sendMessage.Value.(string))
			if Messages.TranslateSettings.Current_language != langCode {
				sendMessage.Value = langCode
				if langCode == "" {
					sendMessage.Value = nil
				}
				Messages.TranslateSettings.Current_language = langCode
			} else {
				sendMessage.Value = SkipMessage
			}
		}
	}
}
