package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"golang.design/x/clipboard"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/SpecialTextToSpeechSettings"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

func ShowSaveTTSWindow(saveFunc func(string)) {
	// find active window
	window, _ := Utilities.GetCurrentMainWindow(lang.L("Save TTS File"))

	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if writer == nil {
			return
		}
		if err != nil {
			log.Println("Error saving file:", err)
			return
		}
		defer writer.Close()

		uri := writer.URI().String()
		// replace "file://" from the beginning of the string
		uri, _ = strings.CutPrefix(uri, "file://")

		saveFunc(uri)

		fyne.CurrentApp().Preferences().SetString("LastTTSSavePath", filepath.Dir(writer.URI().Path()))

	}, window)

	fileSaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".wav"}))
	fileSaveDialog.SetFileName("tts_" + time.Now().Format("2006-01-02_15-04-05") + ".wav")

	saveStartingPath := fyne.CurrentApp().Preferences().StringWithFallback("LastTTSSavePath", "")
	if saveStartingPath != "" {
		// check if folder exists
		folderExists := false
		if _, err := os.Stat(saveStartingPath); !os.IsNotExist(err) {
			folderExists = true
		}
		if folderExists {
			fileURI := storage.NewFileURI(saveStartingPath)
			fileLister, _ := storage.ListerForURI(fileURI)

			fileSaveDialog.SetLocation(fileLister)
		}
	}

	dialogSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
	dialogSize.Height = dialogSize.Height - 50
	dialogSize.Width = dialogSize.Width - 50
	fileSaveDialog.Resize(dialogSize)

	fileSaveDialog.Show()

	return
}

func GetClipboardText() string {
	clipboardText := ""
	err := clipboard.Init()
	if err == nil {
		clipboardBinary := clipboard.Read(clipboard.FmtText)
		if clipboardBinary != nil {
			clipboardText = string(clipboardBinary)
		}
	}
	if len(fyne.CurrentApp().Driver().AllWindows()) > 0 && clipboardText == "" {
		clipboardText = fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().Content()
	}
	return clipboardText
}

func OnOpenTextToSpeechWindow(container fyne.CanvasObject) {
	if Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel("sep-tts") == nil {
		seperatorMenuItem := fyne.NewMenuItemSeparator()
		seperatorMenuItem.Label = "sep-tts"
		Fields.Field.TranscriptionTranslationTextToSpeechInput.AddAdditionalMenuItem(seperatorMenuItem)
	}
	if Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel(lang.L("Send to TTS from Clipboard")) == nil {
		Fields.Field.TranscriptionTranslationTextToSpeechInput.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Send to TTS from Clipboard"), func() {
			clipboardText := GetClipboardText()
			if clipboardText == "" {
				return
			}
			valueData := struct {
				Text     string `json:"text"`
				ToDevice bool   `json:"to_device"`
				Download bool   `json:"download"`
			}{
				Text:     clipboardText,
				ToDevice: true,
				Download: false,
			}
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
	}
	if Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel(lang.L("Export .wav from Clipboard")) == nil {
		Fields.Field.TranscriptionTranslationTextToSpeechInput.AddAdditionalMenuItem(fyne.NewMenuItem(lang.L("Export .wav from Clipboard"), func() {
			clipboardText := GetClipboardText()
			if clipboardText == "" {
				return
			}
			ShowSaveTTSWindow(func(s string) {
				sendMessage := SendMessageChannel.SendMessageStruct{
					Type: "tts_req",
					Value: struct {
						Text     string `json:"text"`
						ToDevice bool   `json:"to_device"`
						Download bool   `json:"download"`
						Path     string `json:"path,omitempty"`
					}{
						Text:     clipboardText,
						ToDevice: false,
						Download: true,
						Path:     s,
					},
				}
				sendMessage.SendMessage()
			})
		}))
	}
}

func OnCloseTextToSpeechWindow(container fyne.CanvasObject) {
	seperatorMenuItem := Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel("sep-tts")
	if seperatorMenuItem == nil {
		Fields.Field.TranscriptionTranslationTextToSpeechInput.RemoveAdditionalMenuItem(seperatorMenuItem)
	}
	playFromClipboardMenuItem := Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel(lang.L("Send to TTS from Clipboard"))
	if playFromClipboardMenuItem != nil {
		Fields.Field.TranscriptionTranslationTextToSpeechInput.RemoveAdditionalMenuItem(playFromClipboardMenuItem)
	}
	exportFromClipboardMenuItem := Fields.Field.TranscriptionTranslationTextToSpeechInput.FindAdditionalMenuItemByLabel(lang.L("Export .wav from Clipboard"))
	if exportFromClipboardMenuItem != nil {
		Fields.Field.TranscriptionTranslationTextToSpeechInput.RemoveAdditionalMenuItem(exportFromClipboardMenuItem)
	}
}

func CreateTextToSpeechWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\TextToSpeech->CreateTextToSpeechWindow")
	})

	ttsModels := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Model")+":"), Fields.Field.TtsModelCombo)

	var saveRandomVoiceButton fyne.CanvasObject
	var advancedSettings fyne.CanvasObject
	switch Settings.Config.Tts_type {
	case "silero":
		saveRandomVoiceButton = widget.NewButtonWithIcon(lang.L("Save Random Voice"), theme.DocumentSaveIcon(), func() {
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "tts_voice_save_req",
			}
			sendMessage.SendMessage()
			Fields.Field.TtsVoiceCombo.SetSelected("last")
		})
	case "f5_e2", "zonos":
		saveRandomVoiceButton = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "tts_voice_reload_req",
			}
			sendMessage.SendMessage()
		})

		if Settings.Config.Tts_type == "zonos" {
			advancedSettings = SpecialTextToSpeechSettings.BuildZonosSpecialSettings()
		}
	}
	ttsVoices := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Voice")+":"), Fields.Field.TtsVoiceCombo)

	ttsVoicesSaveBtnLayout := container.NewBorder(nil, nil, nil, saveRandomVoiceButton, ttsVoices)

	ttsModelVoiceRow := container.New(layout.NewGridLayout(2), ttsModels, ttsVoicesSaveBtnLayout)

	transcriptionRow := container.New(layout.NewGridLayout(1), Fields.Field.TranscriptionTranslationTextToSpeechInput)

	exportLastSpeechButton := widget.NewButton(lang.L("Export Last Generation"), func() {
		ShowSaveTTSWindow(func(s string) {
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "tts_req_last",
				Value: struct {
					ToDevice bool   `json:"to_device"`
					Download bool   `json:"download"`
					Path     string `json:"path,omitempty"`
				}{
					ToDevice: false,
					Download: true,
					Path:     s,
				},
			}
			sendMessage.SendMessage()
		})
	})
	exportLastSpeechButton.Disable()

	exportSpeechButton := widget.NewButtonWithIcon(lang.L("Export .wav"), theme.DocumentSaveIcon(), func() {
		ShowSaveTTSWindow(func(s string) {
			text, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "tts_req",
				Value: struct {
					Text     string `json:"text"`
					ToDevice bool   `json:"to_device"`
					Download bool   `json:"download"`
					Path     string `json:"path,omitempty"`
				}{
					Text:     text,
					ToDevice: false,
					Download: true,
					Path:     s,
				},
			}
			sendMessage.SendMessage()

			exportLastSpeechButton.Enable()
		})
	})

	sendFunction := func() {
		text, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
		valueData := struct {
			Text     string `json:"text"`
			ToDevice bool   `json:"to_device"`
			Download bool   `json:"download"`
		}{
			Text:     text,
			ToDevice: true,
			Download: false,
		}
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
		exportLastSpeechButton.Enable()
	}
	sendButton := widget.NewButtonWithIcon(lang.L("Send to Text-to-Speech"), theme.MediaPlayIcon(), sendFunction)
	sendButton.Importance = widget.HighImportance

	testButton := widget.NewButton(lang.L("Test the Voice"), func() {
		text, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
		valueData := struct {
			Text        string      `json:"text"`
			ToDevice    bool        `json:"to_device"`
			DeviceIndex interface{} `json:"device_index"`
			Download    bool        `json:"download"`
		}{
			Text:        text,
			ToDevice:    true,
			DeviceIndex: nil, // send to default device
			Download:    false,
		}
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
		exportLastSpeechButton.Enable()
	})

	stopPlayButton := widget.NewButtonWithIcon(lang.L("Stop playing"), theme.MediaStopIcon(), func() {
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "audio_stop",
			Value: "tts",
		}
		sendMessage.SendMessage()
	})

	buttonRow := container.NewHBox(
		exportSpeechButton,
		exportLastSpeechButton,
		layout.NewSpacer(),
		testButton,
		sendButton,
		stopPlayButton,
	)

	bottomButtonPart := container.New(layout.NewVBoxLayout(), buttonRow)
	if advancedSettings != nil {
		bottomButtonPart.Add(advancedSettings)
	}

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			ttsModelVoiceRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			bottomButtonPart,
		),
	)

	// add shortcuts
	sendShortcut := CustomWidget.ShortcutEntrySubmit{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl,
		Handler: func() {
			text, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
			if mainContent.Visible() {
				sendFunction()
			} else {
				ttsEnabled, _ := Fields.DataBindings.TextToSpeechEnabledDataBinding.Get()
				oscEnabled, _ := Fields.DataBindings.OSCEnabledDataBinding.Get()
				if ttsEnabled {
					sendFunction()
				}
				if oscEnabled {
					sendMessage := SendMessageChannel.SendMessageStruct{
						Type: "send_osc",
						Value: struct {
							Text string `json:"text"`
						}{
							Text: text,
						},
					}
					sendMessage.SendMessage()
				}
			}
		},
	}
	Fields.Field.TranscriptionTranslationTextToSpeechInput.AddCustomShortcut(sendShortcut)

	return mainContent
}
