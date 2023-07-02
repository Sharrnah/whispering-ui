package Websocket

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"log"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
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

	// only in case of LLM message
	LlmAnswer string `json:"llm_answer,omitempty"`
}

var (
	resetRealtimeLabelHideTimer = make(chan bool)
	realtimeLabelTimer          *time.Timer
	realtimeLabelTimerMutex     sync.Mutex
)

func realtimeLabelHideTimer() {
	for {
		select {
		case <-resetRealtimeLabelHideTimer:
			realtimeLabelTimerMutex.Lock()
			if realtimeLabelTimer != nil {
				realtimeLabelTimer.Stop()
			}
			realtimeLabelTimer = time.AfterFunc(5*time.Second, func() {
				//Fields.Field.RealtimeResultLabel.Hide()
				Fields.Field.RealtimeResultLabel.SetText("")
			})
			realtimeLabelTimerMutex.Unlock()
		}
	}
}

var (
	resetProcessingStopTimer = make(chan bool)
	processingTimer          *time.Timer
	processingStopTimerMutex sync.Mutex
)

func processingStopTimer() {
	for {
		select {
		case <-resetProcessingStopTimer:
			processingStopTimerMutex.Lock()
			if processingTimer != nil {
				processingTimer.Stop()
			}
			processingTimer = time.AfterFunc(10*time.Second, func() {
				Fields.Field.ProcessingStatus.Stop()
			})
			processingStopTimerMutex.Unlock()
		}
	}
}

func (c *MessageStruct) GetMessage(messageData []byte) *MessageStruct {
	// no message data
	if messageData == nil {
		return nil
	}
	c.Raw = messageData
	msgStruct, err := messageLoader(c, messageData)
	if err != nil {
		log.Println(err)
		return nil
	}
	return msgStruct.(*MessageStruct)
}

// Handle the different receiving message types

func (c *MessageStruct) HandleReceiveMessage() {
	var err error = nil

	switch c.Type {
	case "error":
		errorMessage := Messages.ExceptionMessage{}
		err = json.Unmarshal(c.Raw, &errorMessage)
		if err != nil {
			log.Println(err)
			return
		}
		errorMessage.ShowError(fyne.CurrentApp().Driver().AllWindows()[0])
	case "info":
		errorMessage := Messages.ExceptionMessage{}
		err = json.Unmarshal(c.Raw, &errorMessage)
		if err != nil {
			log.Println(err)
			return
		}
		errorMessage.ShowInfo(fyne.CurrentApp().Driver().AllWindows()[0])
	case "installed_languages":
		err = json.Unmarshal(c.Raw, &Messages.InstalledLanguages)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.InstalledLanguages.Update()
	case "available_tts_models":
		err = json.Unmarshal(c.Raw, &Messages.TtsLanguages)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.TtsLanguages.Update()
	case "available_tts_voices":
		err = json.Unmarshal(c.Raw, &Messages.TtsVoices)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.TtsVoices.Update()
	case "available_img_languages":
		err = json.Unmarshal(c.Raw, &Messages.OcrLanguagesList)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.OcrLanguagesList.Update()
	case "windows_list":
		err = json.Unmarshal(c.Raw, &Messages.WindowsList)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.WindowsList.Update()
	case "settings_values":
		var (
			i  interface{}
			ok bool
		)
		err = json.Unmarshal(c.Data, &i)
		if err != nil {
			log.Println(err)
			return
		}
		if Settings.ConfigValues, ok = i.(map[string]interface{}); !ok {
			log.Println("failed to type assert data")
		}
	case "translate_settings":
		// skip received run_backend value from receiving
		var runBackend = true
		var websocketIp string
		var websocketPort int
		if !Messages.TranslateSettings.Run_backend {
			runBackend = false
			websocketIp = Messages.TranslateSettings.Websocket_ip
			websocketPort = Messages.TranslateSettings.Websocket_port
		}

		err = json.Unmarshal(c.Data, &Messages.TranslateSettings)

		if !runBackend {
			Messages.TranslateSettings.Run_backend = runBackend
			Messages.TranslateSettings.Websocket_ip = websocketIp
			Messages.TranslateSettings.Websocket_port = websocketPort
		}

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

		// stop processing status
		Fields.Field.ProcessingStatus.Stop()
		Fields.Field.ProcessingStatus.Refresh()

		Fields.Field.RealtimeResultLabel.SetText(c.Text)
		Fields.Field.RealtimeResultLabel.Refresh()

		select {
		// reset processing status timer
		case resetRealtimeLabelHideTimer <- true:
		default:
		}
	case "translate_result":
		Messages.LastTranslationResult = c.TranslateResult
		Fields.Field.TranscriptionTranslationInput.SetText(c.TranslateResult)
		if c.OriginalText != "" {
			Fields.Field.TranscriptionInput.SetText(c.OriginalText)
		}
		if Fields.Field.SourceLanguageCombo.GetSelected() != nil && Fields.Field.SourceLanguageCombo.GetSelected().Value == "Auto" {
			langName := Utilities.LanguageMapList.GetName(c.TxtFromLang)
			Settings.Config.Last_auto_txt_translate_lang = c.TxtFromLang
			if langName == "" {
				langName = c.TxtFromLang
			}
			Fields.Field.SourceLanguageCombo.Options[0].Text = "Auto [detected: " + langName + "]"
			Fields.Field.SourceLanguageCombo.SetSelected(Fields.Field.SourceLanguageCombo.Options[0].Value)
			Fields.Field.SourceLanguageCombo.Refresh()
		}
		//case "tts_save":
		//	var audioData = Audio.TtsResultRaw{}
		//	err = json.Unmarshal(c.Raw, &audioData)
		//	go func() {
		//		err := audioData.PlayWAVFromBase64()
		//		if err != nil {
		//			log.Println(err)
		//		}
		//	}()

		//	byteData, _ := b64.StdEncoding.DecodeString(audioData.WavData)
		//	audioData.WavBinary = byteData
		//	audioData.WavData = ""
		//	Audio.LastFile = audioData
		//	go Audio.LastFile.Play()
	case "ocr_result":
		err = json.Unmarshal(c.Data, &Messages.OcrResult)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.OcrResult.Update()

	// special case for LLM plugin
	case "llm_answer":
		c.Text = strings.TrimSpace(c.Text)
		c.TxtTranslation = strings.TrimSpace(c.LlmAnswer)
		whisperResultMessage := Messages.WhisperResult{
			Text:                 c.Text,
			Language:             c.Language,
			TxtTranslation:       c.TxtTranslation,
			TxtTranslationTarget: c.TxtTranslationTarget,
		}

		whisperResultMessage.Update()

		// stop processing status
		Fields.Field.ProcessingStatus.Stop()
		Fields.Field.ProcessingStatus.Refresh()

	case "processing_start":
		var processingStarted = false
		err = json.Unmarshal(c.Data, &processingStarted)
		if err != nil {
			log.Println(err)
			return
		}
		if processingStarted {
			Fields.Field.ProcessingStatus.Start()
			Fields.Field.ProcessingStatus.Refresh()
			select {
			// reset processing status timer
			case resetProcessingStopTimer <- true:
			default:
			}
		} else {
			Fields.Field.ProcessingStatus.Stop()
			Fields.Field.ProcessingStatus.Refresh()
		}
	case "processing_data":
		var processingData = ""
		err = json.Unmarshal(c.Data, &processingData)
		if err != nil {
			log.Println(err)
			return
		}
		if processingData != "" {
			Fields.Field.ProcessingStatus.Start()
			Fields.Field.ProcessingStatus.Refresh()
			Fields.Field.RealtimeResultLabel.Show()
			Fields.Field.RealtimeResultLabel.SetText(processingData)
			Fields.Field.RealtimeResultLabel.Refresh()
			select {
			// reset hide realtime label timer
			case resetRealtimeLabelHideTimer <- true:
			// reset processing status timer
			case resetProcessingStopTimer <- true:
			default:
			}
		}
	case "loading_state":
		if c.Raw == nil {
			return
		}
		err = json.Unmarshal(c.Raw, &Messages.CurrentLoadingState)
		if err != nil {
			Messages.LoadingStateContainer.RemoveAll()
			Messages.LoadingStateDialog.Hide()
			return
		}
		Messages.CurrentLoadingState.Update()
		fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Content().Refresh()
	case "tts_save":
		ttsSpeechAudio := Messages.TtsSpeechAudio{}
		err = json.Unmarshal(c.Raw, &ttsSpeechAudio)
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		}
		if err == nil && len(ttsSpeechAudio.WavData) > 0 {
			ttsSpeechAudio.SaveWav()
		}
	case "download":
		download := Messages.DownloadMessage{}
		err = json.Unmarshal(c.Raw, &download)
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		go func() {
			err = download.StartDownload()
			if err != nil {
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}
			fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Content().Refresh()
		}()
	}

	// set focus to main window
	if fyne.CurrentApp().Preferences().BoolWithFallback("AutoRefocusWindow", false) {
		fyne.CurrentApp().Driver().AllWindows()[0].RequestFocus()
	}

	// refresh window
	//fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Content().Refresh()

	if err != nil {
		log.Printf("Unmarshal: %v", err)
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
			if langCode != "" {
				sendMessage.Value = langCode
			} else {
				sendMessage.Value = SkipMessage
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
		case "tts_model":
			selectedModel := sendMessage.Value.(string)
			var voiceLanguage = ""
			for _, language := range Messages.TtsLanguages.Languages {
				for _, model := range language.Models {
					if model == selectedModel {
						voiceLanguage = language.Language
						break
					}
				}
			}
			sendMessage.Value = []string{voiceLanguage, selectedModel}
		case "ocr_lang":
			//langCode := Messages.OcrLanguagesList.GetCodeByName(sendMessage.Value.(string))
			langCode := sendMessage.Value.(string)
			if langCode != "" && Messages.TranslateSettings.Ocr_lang != langCode {
				sendMessage.Value = langCode
			} else {
				sendMessage.Value = SkipMessage
			}
		}
	}
}
