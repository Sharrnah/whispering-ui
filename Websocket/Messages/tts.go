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

func ttsModelDisplayValue(selection []string) string {
	if len(selection) < 2 {
		return ""
	}
	for displayValue, canonicalValue := range Fields.TtsModelSelectionValues {
		if len(canonicalValue) == 2 && canonicalValue[0] == selection[0] && canonicalValue[1] == selection[1] {
			return displayValue
		}
	}
	return ""
}

func (res TtsLanguagesListing) Update() *TtsLanguagesListing {
	Fields.Field.TtsModelCombo.Options = nil
	Fields.TtsModelSelectionValues = make(map[string][]string)
	for _, languageItem := range res.Languages {
		//elementName := languageItem.Language
		for _, modelItem := range languageItem.Models {
			if Settings.Config.Tts_type == "silero" {
				if strings.Contains(modelItem, "v3") || strings.Contains(modelItem, "v4") {
					Fields.Field.TtsModelCombo.Options = append(Fields.Field.TtsModelCombo.Options, modelItem)
					Fields.TtsModelSelectionValues[modelItem] = []string{languageItem.Language, modelItem}
				}
			} else {
				modelEntry := modelItem + " (" + lang.L(languageItem.Language) + ")"
				Fields.Field.TtsModelCombo.Options = append(Fields.Field.TtsModelCombo.Options, modelEntry)
				Fields.TtsModelSelectionValues[modelEntry] = []string{languageItem.Language, modelItem}
			}
		}
	}
	if selected := ttsModelDisplayValue(Settings.Config.Tts_model); selected != "" {
		Fields.Field.TtsModelCombo.Selected = selected
		Fields.Field.TtsModelCombo.Refresh()
		if Fields.TtsModelSelectionChanged != nil {
			Fields.TtsModelSelectionChanged(selected)
		}
	} else if len(Fields.Field.TtsModelCombo.Options) > 0 {
		// A profile can retain the previous TTS family's model. Select and send
		// the new family's first backend-provided model so the UI never displays
		// an empty value and the profile receives a canonical [group, model].
		Fields.Field.TtsModelCombo.SetSelected(Fields.Field.TtsModelCombo.Options[0])
	} else {
		Fields.Field.TtsModelCombo.Refresh()
		if Fields.TtsModelSelectionChanged != nil {
			Fields.TtsModelSelectionChanged("")
		}
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

func hasSelectableTTSVoice(options []CustomWidget.TextValueOption) bool {
	return len(options) > 0 && !(len(options) == 1 && options[0].Value == "")
}

func (res TtsVoicesListing) Update() *TtsVoicesListing {
	lastSelectedVoice := ""
	if selected := Fields.Field.TtsVoiceCombo.GetSelected(); selected != nil {
		lastSelectedVoice = selected.Value
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
	if !voicesListContainsSelectedVoice && hasSelectableTTSVoice(Fields.Field.TtsVoiceCombo.Options) {
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
	fileSaveDialog.Show()
	// FileDialog.Resize calls MinSize and requires Show to initialize its
	// internal dialog first in the vendored Fyne version.
	fileSaveDialog.Resize(dialogSize)
}
