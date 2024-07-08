package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Utilities"
)

func ShowSaveTTSWindow(saveFunc func(string)) {
	// find active window
	window := fyne.CurrentApp().Driver().AllWindows()[0]
	if len(fyne.CurrentApp().Driver().AllWindows()) == 1 && fyne.CurrentApp().Driver().AllWindows()[0] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[0]
	} else if len(fyne.CurrentApp().Driver().AllWindows()) == 2 && fyne.CurrentApp().Driver().AllWindows()[1] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[1]
		// more general fallbacks in case more than 1 or 2 windows
	} else if len(fyne.CurrentApp().Driver().AllWindows()) > 0 && fyne.CurrentApp().Driver().AllWindows()[0] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[0]
	} else if len(fyne.CurrentApp().Driver().AllWindows()) > 0 && fyne.CurrentApp().Driver().AllWindows()[1] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[1]
	} else {
		return
	}

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
	//mainContent := container.(*fyne.Container)
	//clipboardText := Fields.Field.TranscriptionTranslationInput.Text
	if Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("sep-tts") == nil {
		seperatorMenuItem := fyne.NewMenuItemSeparator()
		seperatorMenuItem.Label = "sep-tts"
		Fields.Field.TranscriptionTranslationInput.AddAdditionalMenuItem(seperatorMenuItem)
	}
	if Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("Send to TTS from Clipboard") == nil {
		Fields.Field.TranscriptionTranslationInput.AddAdditionalMenuItem(fyne.NewMenuItem("Send to TTS from Clipboard", func() {
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
			sendMessage := Fields.SendMessageStruct{
				Type:  "tts_req",
				Value: valueData,
			}
			sendMessage.SendMessage()
		}))
	}
	if Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("Export .wav from Clipboard") == nil {
		Fields.Field.TranscriptionTranslationInput.AddAdditionalMenuItem(fyne.NewMenuItem("Export .wav from Clipboard", func() {
			clipboardText := GetClipboardText()
			if clipboardText == "" {
				return
			}
			ShowSaveTTSWindow(func(s string) {
				sendMessage := Fields.SendMessageStruct{
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
	seperatorMenuItem := Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("sep-tts")
	if seperatorMenuItem == nil {
		Fields.Field.TranscriptionTranslationInput.RemoveAdditionalMenuItem(seperatorMenuItem)
	}
	playFromClipboardMenuItem := Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("Send to TTS from Clipboard")
	if playFromClipboardMenuItem != nil {
		Fields.Field.TranscriptionTranslationInput.RemoveAdditionalMenuItem(playFromClipboardMenuItem)
	}
	exportFromClipboardMenuItem := Fields.Field.TranscriptionTranslationInput.FindAdditionalMenuItemByLabel("Export .wav from Clipboard")
	if exportFromClipboardMenuItem != nil {
		Fields.Field.TranscriptionTranslationInput.RemoveAdditionalMenuItem(exportFromClipboardMenuItem)
	}
}

func CreateTextToSpeechWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	ttsModels := container.New(layout.NewFormLayout(), widget.NewLabel("Model:"), Fields.Field.TtsModelCombo)

	saveRandomVoiceButton := widget.NewButtonWithIcon("Save Random Voice", theme.DocumentSaveIcon(), func() {
		sendMessage := Fields.SendMessageStruct{
			Type: "tts_voice_save_req",
		}
		sendMessage.SendMessage()
		Fields.Field.TtsVoiceCombo.SetSelected("last")
	})
	ttsVoices := container.New(layout.NewFormLayout(), widget.NewLabel("Voice:"), Fields.Field.TtsVoiceCombo)

	ttsVoicesSaveBtnLayout := container.NewBorder(nil, nil, nil, saveRandomVoiceButton, ttsVoices)

	ttsModelVoiceRow := container.New(layout.NewGridLayout(2), ttsModels, ttsVoicesSaveBtnLayout)

	transcriptionRow := container.New(layout.NewGridLayout(1), Fields.Field.TranscriptionTranslationInput)

	exportSpeechButton := widget.NewButtonWithIcon("Export .wav", theme.DocumentSaveIcon(), func() {
		ShowSaveTTSWindow(func(s string) {
			sendMessage := Fields.SendMessageStruct{
				Type: "tts_req",
				Value: struct {
					Text     string `json:"text"`
					ToDevice bool   `json:"to_device"`
					Download bool   `json:"download"`
					Path     string `json:"path,omitempty"`
				}{
					Text:     Fields.Field.TranscriptionTranslationInput.Text,
					ToDevice: false,
					Download: true,
					Path:     s,
				},
			}
			sendMessage.SendMessage()
		})
	})

	sendFunction := func() {
		valueData := struct {
			Text     string `json:"text"`
			ToDevice bool   `json:"to_device"`
			Download bool   `json:"download"`
		}{
			Text:     Fields.Field.TranscriptionTranslationInput.Text,
			ToDevice: true,
			Download: false,
		}
		sendMessage := Fields.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
	}
	sendButton := widget.NewButtonWithIcon("Send to Text-to-Speech", theme.MediaPlayIcon(), sendFunction)
	sendButton.Importance = widget.HighImportance

	testButton := widget.NewButton("Test the Voice", func() {
		valueData := struct {
			Text        string      `json:"text"`
			ToDevice    bool        `json:"to_device"`
			DeviceIndex interface{} `json:"device_index"`
			Download    bool        `json:"download"`
		}{
			Text:        Fields.Field.TranscriptionTranslationInput.Text,
			ToDevice:    true,
			DeviceIndex: nil, // send to default device
			Download:    false,
		}
		sendMessage := Fields.SendMessageStruct{
			Type:  "tts_req",
			Value: valueData,
		}
		sendMessage.SendMessage()
	})

	stopPlayButton := widget.NewButtonWithIcon("Stop playing", theme.MediaStopIcon(), func() {
		sendMessage := Fields.SendMessageStruct{
			Type:  "audio_stop",
			Value: "tts",
		}
		sendMessage.SendMessage()
	})

	buttonRow := container.NewHBox(
		exportSpeechButton,
		layout.NewSpacer(),
		testButton,
		sendButton,
		stopPlayButton,
	)

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			ttsModelVoiceRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.New(layout.NewVBoxLayout(), buttonRow),
		),
	)

	// add shortcuts
	sendShortcut := CustomWidget.ShortcutEntrySubmit{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl,
		Handler: func() {
			if mainContent.Visible() {
				sendFunction()
			} else {
				if Fields.Field.TtsEnabled.Checked {
					sendFunction()
				}
				if Fields.Field.OscEnabled.Checked {
					sendMessage := Fields.SendMessageStruct{
						Type: "send_osc",
						Value: struct {
							Text *string `json:"text"`
						}{
							Text: &Fields.Field.TranscriptionTranslationInput.Text,
						},
					}
					sendMessage.SendMessage()
				}
			}
		},
	}
	Fields.Field.TranscriptionTranslationInput.AddCustomShortcut(sendShortcut)

	return mainContent
}
