package Websocket

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"github.com/getsentry/sentry-go"
	"log"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/SendMessageChannel"
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

var ReceiveMessageChannelBufferSize = 100
var ReceiveMessageChannel = make(chan []byte, ReceiveMessageChannelBufferSize)

func ProcessReceiveMessageChannel() {
	for {
		select {
		case messageBytes := <-ReceiveMessageChannel:
			var msg MessageStruct
			messageStruct := msg.GetMessage(messageBytes)
			if messageStruct != nil {
				msg.HandleReceiveMessage()
			}
		}
	}
}

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
				//Fields.Field.RealtimeResultLabel.SetText("")
				Fields.DataBindings.WhisperResultIntermediateResult.Set("")
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

func messageLoader(c interface{}, message []byte) (interface{}, error) {
	err := json.Unmarshal(message, c)
	if err != nil {
		Logging.CaptureException(err)
		Logging.Flush(Logging.FlushTimeoutDefault)
		log.Fatalf("Unmarshal: %v", err)
		return nil, err
	}
	return c, nil
}

func (c *MessageStruct) GetMessage(messageData []byte) *MessageStruct {
	// no message data
	if messageData == nil {
		return nil
	}
	c.Raw = messageData
	msgStruct, err := messageLoader(c, messageData)
	if err != nil {
		Logging.CaptureException(err)
		log.Println(err)
		return nil
	}
	return msgStruct.(*MessageStruct)
}

var resultListMutex sync.Mutex
var intermediateResultListMutex sync.Mutex
var processingStatusMutex sync.Mutex

// Handle the different receiving message types

func (c *MessageStruct) HandleReceiveMessage() {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Websocket\\messageHandler->HandleReceiveMessage")
	})
	var err error = nil

	switch c.Type {
	case "error":
		errorMessage := Messages.ExceptionMessage{}
		err = json.Unmarshal(c.Raw, &errorMessage)
		if err != nil {
			Logging.CaptureException(err)
			log.Println(err)
			return
		}
		if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
			errorMessage.ShowError(fyne.CurrentApp().Driver().AllWindows()[0])
		}
	case "info":
		errorMessage := Messages.ExceptionMessage{}
		err = json.Unmarshal(c.Raw, &errorMessage)
		if err != nil {
			log.Println(err)
			return
		}
		if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
			errorMessage.ShowInfo(fyne.CurrentApp().Driver().AllWindows()[0])
		}
	case "installed_languages":
		srcLang := Settings.Config.Src_lang
		trgLang := Settings.Config.Trg_lang
		ocrSrcLang := Settings.Config.Ocr_txt_src_lang
		ocrTrgLang := Settings.Config.Ocr_txt_trg_lang
		err = json.Unmarshal(c.Raw, &Messages.InstalledLanguages)
		if err != nil {
			log.Println(err)
			return
		}
		Messages.InstalledLanguages.Update()

		if srcLang == "" {
			srcLang = "auto"
		}
		if ocrSrcLang == "" {
			ocrSrcLang = "auto"
		}

		// set txt source + target lang combo boxes
		Fields.Field.SourceLanguageCombo.Text = Messages.InstalledLanguages.GetNameByCode(srcLang)
		Fields.Field.SourceLanguageCombo.ResetOptionsFilter()
		Fields.Field.TargetLanguageCombo.Text = Messages.InstalledLanguages.GetNameByCode(trgLang)
		Fields.Field.TargetLanguageCombo.ResetOptionsFilter()

		// set ocr txt source + target lang combo boxes
		Fields.Field.SourceLanguageTxtTranslateCombo.Text = Messages.InstalledLanguages.GetNameByCode(ocrSrcLang)
		Fields.Field.SourceLanguageTxtTranslateCombo.ResetOptionsFilter()
		Fields.Field.TargetLanguageTxtTranslateCombo.Text = Messages.InstalledLanguages.GetNameByCode(ocrTrgLang)
		Fields.Field.TargetLanguageTxtTranslateCombo.ResetOptionsFilter()

		// set auto text translate checkbox label
		Fields.Field.TextTranslateEnabled.Text = lang.L("SttTextTranslateLabel", map[string]interface{}{"FromLang": Messages.InstalledLanguages.GetNameByCode(srcLang), "ToLang": Messages.InstalledLanguages.GetNameByCode(trgLang)}) + Fields.AdditionalLanguagesCountString(" ", "[]")
		Fields.Field.TextTranslateEnabled.Refresh()
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

		//go func() {
		println("Whisper Result processing update call.")

		go func(resultMsg_ Messages.WhisperResult) {
			resultListMutex.Lock()
			defer resultListMutex.Unlock()

			resultMsg_.Update()

			// stop processing status
			Fields.Field.ProcessingStatus.Stop()
			//Fields.Field.ProcessingStatus.Refresh()

			//Fields.Field.RealtimeResultLabel.SetText(whisperResultMessage.Text)
			Fields.DataBindings.WhisperResultIntermediateResult.Set(resultMsg_.Text)

			//Fields.Field.WhisperResultList.
		}(whisperResultMessage)

		// force refresh of speech-to-text tab
		//if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
		//	appTabs := fyne.CurrentApp().Driver().AllWindows()[0].Content().(*container.AppTabs)
		//	if appTabs.Selected().Text == "Speech-to-Text" {
		//		appTabs.Selected().Content.Refresh()
		//	}
		//}
		//}()

		select {
		// reset processing status timer
		case resetRealtimeLabelHideTimer <- true:
		default:
		}
	case "translate_result":
		//Messages.LastTranslationResult = c.TranslateResult
		//Fields.Field.TranscriptionTranslationSpeechToTextInput.SetText(c.TranslateResult)
		Fields.DataBindings.TranscriptionTranslationInputBinding.Set(c.TranslateResult)
		if c.OriginalText != "" {
			//Fields.Field.TranscriptionSpeechToTextInput.SetText(c.OriginalText)
			Fields.DataBindings.TranscriptionInputBinding.Set(c.OriginalText)
		}
		if Fields.Field.SourceLanguageCombo.GetCurrentValueOptionEntry() != nil && Fields.Field.SourceLanguageCombo.GetCurrentValueOptionEntry().Value == "Auto" {
			langName := Utilities.LanguageMapList.GetName(c.TxtFromLang)
			Settings.Config.Last_auto_txt_translate_lang = c.TxtFromLang
			if langName == "" {
				langName = c.TxtFromLang
			}
			Fields.Field.SourceLanguageCombo.OptionsTextValue[0].Text = "Auto [detected: " + langName + "]"
			Fields.Field.SourceLanguageCombo.Options[0] = Fields.Field.SourceLanguageCombo.OptionsTextValue[0].Text
			Fields.Field.SourceLanguageCombo.Text = Fields.Field.SourceLanguageCombo.Options[0]
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

		go func(ocrResult_ Messages.OcrResultData) {
			ocrResult_.Update()
		}(Messages.OcrResult)

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

		go func(resultMsg_ Messages.WhisperResult) {
			resultListMutex.Lock()
			defer resultListMutex.Unlock()

			resultMsg_.Update()

			// stop processing status
			Fields.Field.ProcessingStatus.Stop()
		}(whisperResultMessage)

		// stop processing status
		//Fields.Field.ProcessingStatus.Stop()
		//Fields.Field.ProcessingStatus.Refresh()

	case "processing_start":
		var processingStarted = false
		err = json.Unmarshal(c.Data, &processingStarted)
		if err != nil {
			log.Println(err)
			return
		}
		go func(processStarted_ bool) {
			processingStatusMutex.Lock()
			defer processingStatusMutex.Unlock()
			if processStarted_ {
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
		}(processingStarted)
	case "processing_data":
		var processingData = ""
		err = json.Unmarshal(c.Data, &processingData)
		if err != nil {
			log.Println(err)
			return
		}

		if processingData != "" {
			go func(procData_ string) {
				intermediateResultListMutex.Lock()
				defer intermediateResultListMutex.Unlock()
				Fields.DataBindings.WhisperResultIntermediateResult.Set(procData_)
			}(processingData)

			Fields.Field.ProcessingStatus.Start()
			//Fields.Field.ProcessingStatus.Refresh()
			Fields.Field.RealtimeResultLabel.Show()
			//Fields.Field.RealtimeResultLabel.SetText(processingData)
			//Fields.Field.RealtimeResultLabel.Refresh()

			// Attempt to send to resetRealtimeLabelHideTimer
			select {
			case resetRealtimeLabelHideTimer <- true:
			default:
			}

			// Attempt to send to resetProcessingStopTimer
			select {
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
	case "tts_save":
		ttsSpeechAudio := Messages.TtsSpeechAudio{}
		err = json.Unmarshal(c.Raw, &ttsSpeechAudio)
		if err != nil {
			if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
				currentMainWindow, _ := Utilities.GetCurrentMainWindow("")
				Logging.CaptureException(err)
				dialog.ShowError(err, currentMainWindow)
			}
		}
		if err == nil && len(ttsSpeechAudio.WavData) > 0 {
			ttsSpeechAudio.SaveWav()
		}
	case "download":
		download := Messages.DownloadMessage{}
		err = json.Unmarshal(c.Raw, &download)
		if err != nil {
			if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
				currentMainWindow, _ := Utilities.GetCurrentMainWindow("")
				Logging.CaptureException(err)
				dialog.ShowError(err, currentMainWindow)
			}
			return
		}
		go func(dl_ Messages.DownloadMessage) {
			err = dl_.StartDownload()
			if err != nil {
				if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
					currentMainWindow, _ := Utilities.GetCurrentMainWindow("")
					dialog.ShowError(err, currentMainWindow)
				}
				return
			}
			if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
				currentMainWindow, _ := Utilities.GetCurrentMainWindow("")
				currentMainWindow.Canvas().Content().Refresh()
			}
		}(download)
	}

	// set focus to main window
	if fyne.CurrentApp().Preferences().BoolWithFallback("AutoRefocusWindow", false) {
		if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
			fyne.CurrentApp().Driver().AllWindows()[0].RequestFocus()
		}
	}

	// refresh window
	//fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Content().Refresh()

	if err != nil {
		log.Printf("Unmarshal: %v", err)
	}

}

func HandleSendMessage(sendMessage *SendMessageChannel.SendMessageStruct) {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Websocket\\messageHandler->HandleSendMessage")
	})

	switch sendMessage.Type {
	case "setting_change":
		switch sendMessage.Name {
		case "src_lang", "ocr_txt_src_lang":
			langCode := Messages.InstalledLanguages.GetCodeByName(sendMessage.Value.(string))
			if langCode == "" {
				langCode = sendMessage.Value.(string)
			}
			if langCode == "" {
				langCode = "auto"
			}
			if langCode != "" && Messages.TranslateSettings.Src_lang != langCode {
				sendMessage.Value = langCode
			} else {
				sendMessage.Value = SkipMessage
			}
		case "trg_lang", "ocr_txt_trg_lang":
			langCode := Messages.InstalledLanguages.GetCodeByName(sendMessage.Value.(string))
			if langCode == "" {
				langCode = sendMessage.Value.(string)
			}
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
		case "target_language":
			langCode := Messages.TranslateSettings.GetWhisperLanguageCodeByName(sendMessage.Value.(string))
			if Messages.TranslateSettings.Target_language != langCode {
				sendMessage.Value = langCode
				if langCode == "" {
					sendMessage.Value = nil
				}
				Messages.TranslateSettings.Target_language = langCode
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
		case "txt_translate":
			val := sendMessage.Value.(bool)
			Settings.Config.Txt_translate = val
			Fields.DataBindings.TextTranslateEnabledDataBinding.Set(val)
		case "stt_enabled":
			val := sendMessage.Value.(bool)
			Settings.Config.Stt_enabled = val
			Fields.DataBindings.SpeechToTextEnabledDataBinding.Set(val)
		case "tts_answer":
			val := sendMessage.Value.(bool)
			Settings.Config.Tts_answer = val
			Fields.DataBindings.TextToSpeechEnabledDataBinding.Set(val)
		case "osc_auto_processing_enabled":
			val := sendMessage.Value.(bool)
			Settings.Config.Osc_auto_processing_enabled = val
			Fields.DataBindings.OSCEnabledDataBinding.Set(val)
		}
	}
}
