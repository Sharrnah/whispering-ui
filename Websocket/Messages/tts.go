package Messages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/storage"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
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
			if Settings.Config.Tts_type == "silero" {
				if strings.Contains(modelItem, "v3") || strings.Contains(modelItem, "v4") {
					Fields.Field.TtsModelCombo.Options = append(Fields.Field.TtsModelCombo.Options, modelItem)
				}
			} else {
				modelEntry := modelItem + " (" + languageItem.Language + ")"
				Fields.Field.TtsModelCombo.Options = append(Fields.Field.TtsModelCombo.Options, modelEntry)
			}
		}
	}
	if len(Settings.Config.Tts_model) > 0 {
		Fields.Field.TtsModelCombo.Selected = Settings.Config.Tts_model[1]
		Fields.Field.TtsModelCombo.Refresh()
	}
	return &res
}

// TTS Voices

type TtsVoicesListing struct {
	//Voices []string `json:"data"`
	Voices []Voice `json:"data"`
}
type Voice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var TtsVoices TtsVoicesListing

func (res TtsVoicesListing) Update() *TtsVoicesListing {
	lastSelectedVoice := ""
	if len(Fields.Field.TtsVoiceCombo.Options) > 0 {
		lastSelectedVoice = Fields.Field.TtsVoiceCombo.Selected
	}
	Fields.Field.TtsVoiceCombo.Options = nil
	for _, voice := range res.Voices {
		text := voice.Name
		if text == "open_voice_dir" {
			text = lang.L(text)
		}
		Fields.Field.TtsVoiceCombo.Options = append(Fields.Field.TtsVoiceCombo.Options, CustomWidget.TextValueOption{
			Text:  text,
			Value: voice.Value,
		})
	}
	//Fields.Field.TtsVoiceCombo.Options = append(Fields.Field.TtsVoiceCombo.Options, res.Voices...)

	// set first voice if selection is not in list
	voicesListContainsSelectedVoice := false
	for _, voice := range res.Voices {
		if voice.Value == Settings.Config.Tts_voice {
			voicesListContainsSelectedVoice = true
			break
		}
	}
	voicesListContainsComboboxSelectedVoice := false
	for _, voiceOption := range Fields.Field.TtsVoiceCombo.Options {
		if voiceOption.Value == lastSelectedVoice {
			voicesListContainsComboboxSelectedVoice = true
			break
		}
	}
	// only set new tts voice if select is not received tts_voice and
	// if select is not empty and does not contain only one empty element
	if !voicesListContainsSelectedVoice && Fields.Field.TtsVoiceCombo.Options != nil && (len(Fields.Field.TtsVoiceCombo.Options) > 0 &&
		(len(Fields.Field.TtsVoiceCombo.Options) == 1 && Fields.Field.TtsVoiceCombo.Options[0].Value != "")) {
		Fields.Field.TtsVoiceCombo.SetSelectedIndex(0)
	}
	if Settings.Config.Tts_voice != "" && voicesListContainsSelectedVoice {
		Fields.Field.TtsVoiceCombo.SetSelected(Settings.Config.Tts_voice)
	}
	if lastSelectedVoice != "" && voicesListContainsComboboxSelectedVoice {
		Fields.Field.TtsVoiceCombo.SetSelected(lastSelectedVoice)
	}

	return &res
}

// TTS Speech Audio

type TtsSpeechAudio struct {
	Type    string `json:"type"`
	WavData []byte `json:"wav_data"`
}

func (res TtsSpeechAudio) SaveWav() {
	currentMainWindow, _ := Utilities.GetCurrentMainWindow("Save TTS Wav")
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if writer == nil {
			return
		}
		if err != nil {
			log.Println("Error saving file:", err)
			return
		}
		defer writer.Close()
		writer.Write(res.WavData) // write wav data to file

		fyne.CurrentApp().Preferences().SetString("LastTTSSavePath", filepath.Dir(writer.URI().Path()))

	}, currentMainWindow)

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
}
