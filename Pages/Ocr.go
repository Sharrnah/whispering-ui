package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

func CreateOcrWindow() fyne.CanvasObject {

	Fields.Field.OcrLanguageCombo.OnSubmitted = func(value string) {
		for i := 0; i < len(Fields.Field.OcrLanguageCombo.Options); i++ {
			if strings.Contains(strings.ToLower(Fields.Field.OcrLanguageCombo.Options[i]), strings.ToLower(value)) {
				Fields.Field.OcrLanguageCombo.SelectItemByValue(Fields.Field.OcrLanguageCombo.Options[i])
				value = Fields.Field.OcrLanguageCombo.Options[i]
				Fields.Field.OcrLanguageCombo.Text = value
				Fields.Field.OcrLanguageCombo.Entry.CursorColumn = len(Fields.Field.OcrLanguageCombo.Text)
				Fields.Field.OcrLanguageCombo.Refresh()
				break
			}
		}

		valueIso := Messages.OcrLanguagesList.GetCodeByName(value)
		if valueIso == "" {
			valueObj := Fields.Field.SourceLanguageCombo.GetValueOptionEntryByText(value)
			value = valueObj.Value

			valueIso = Messages.OcrLanguagesList.GetCodeByName(value)
		}

		sendMessage := Fields.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_lang",
			Value: valueIso,
		}
		sendMessage.SendMessage()

		log.Println("ocr Select set to", value)
	}

	container.New(layout.NewMaxLayout())
	ocrLanguageWindowForm := container.New(layout.NewFormLayout(), widget.NewLabel("Text in Image Language:"), Fields.Field.OcrLanguageCombo, widget.NewLabel("Window:"), Fields.Field.OcrWindowCombo)

	ocrSettingsRow := container.New(layout.NewGridLayout(1), ocrLanguageWindowForm)

	ocrButton := widget.NewButtonWithIcon("process window with OCR", theme.ConfirmIcon(), func() {

		ocrLanguageCode := Messages.OcrLanguagesList.GetCodeByName(Fields.Field.OcrLanguageCombo.Text)

		fromLang := ""
		if len(Fields.Field.SourceLanguageCombo.OptionsTextValue) > 0 {
			fromLang = Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.GetValueOptionEntryByText(Fields.Field.SourceLanguageCombo.Text).Value)
		} else {
			fromLang = Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Text)
		}
		if fromLang == "" || fromLang == "Auto" {
			fromLang = "auto"
		}
		if fromLang == "auto" {
			guessedSrcLangByOCRLang := ""
			// try to guess the language from the OCR language selection if auto to lessen the language guessing
			guessedSrcLangByOCRLang = Messages.InstalledLanguages.GetCodeByName(Fields.Field.OcrLanguageCombo.Text)
			if guessedSrcLangByOCRLang == "" {
				if Utilities.LanguageMapList.GetName(ocrLanguageCode) != "" {
					guessedSrcLangByOCRLang = ocrLanguageCode
				}
			}
			if guessedSrcLangByOCRLang != "" {
				fromLang = guessedSrcLangByOCRLang
				println("guessedSrcLangByOCRLang", guessedSrcLangByOCRLang)
			}
		}

		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageTxtTranslateCombo.Text)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "ocr_req",
			Value: struct {
				Ocr_lang  string `json:"ocr_lang"`
				From_lang string `json:"from_lang"`
				To_lang   string `json:"to_lang"`
			}{
				Ocr_lang:  ocrLanguageCode,
				From_lang: fromLang,
				To_lang:   toLang,
			},
		}
		sendMessage.SendMessage()
	})
	ocrButton.Importance = widget.HighImportance

	buttonRow := container.NewHBox(layout.NewSpacer(),
		ocrButton,
	)

	sourceLanguageForm := container.New(layout.NewFormLayout(), widget.NewLabel("Source Language:"), Fields.Field.SourceLanguageCombo)
	targetLanguageForm := container.New(layout.NewFormLayout(), widget.NewLabel("Target Language:"), Fields.Field.TargetLanguageTxtTranslateCombo)
	languageRow := container.New(layout.NewGridLayout(2), sourceLanguageForm, targetLanguageForm)

	transcriptionRow := container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionInput, Fields.Field.TranscriptionTranslationInput)

	ocrContent := container.New(layout.NewVBoxLayout(),
		ocrSettingsRow,
		container.New(layout.NewPaddedLayout(), buttonRow),
		widget.NewSeparator(),
		widget.NewLabel("Text-Translation of OCR Result:"),
		languageRow,
	)

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			ocrContent,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			Fields.Field.OcrImageContainer,
		),
	)

	return mainContent
}
