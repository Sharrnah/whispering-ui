package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

func CreateOcrWindow() fyne.CanvasObject {

	Fields.Field.OcrLanguageCombo.OnChanged = func(value string) {
		valueIso := Messages.OcrLanguagesList.GetCodeByName(value)
		srcTargetLangCode := Utilities.LanguageMapList.GetISO1(valueIso)
		if srcTargetLangCode == "" && len(Utilities.LanguageMapList.GetISO3(valueIso)) > 0 {
			srcTargetLangCode = Utilities.LanguageMapList.GetISO3(valueIso)[0]
		}
		if srcTargetLangCode == "" {
			srcTargetLangCode = valueIso
		}

		Fields.Field.SourceLanguageCombo.SetSelected(srcTargetLangCode)

		sendMessage := Fields.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_lang",
			Value: valueIso,
		}
		sendMessage.SendMessage()

		log.Println("Select set to", value)
	}

	container.New(layout.NewMaxLayout())
	ocrLanguageWindowForm := container.New(layout.NewFormLayout(), widget.NewLabel("OCR Language:"), Fields.Field.OcrLanguageCombo, widget.NewLabel("Window:"), Fields.Field.OcrWindowCombo)
	//ocrWindowForm := container.New(layout.NewFormLayout(), widget.NewLabel("Window:"), Fields.Field.OcrWindowCombo)
	ocrSettingsRow := container.New(layout.NewGridLayout(1), ocrLanguageWindowForm)

	ocrButton := widget.NewButtonWithIcon("process window with OCR", theme.ConfirmIcon(), func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Selected)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageTxtTranslateCombo.Selected)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "ocr_req",
			Value: struct {
				Ocr_lang  string `json:"ocr_lang"`
				From_lang string `json:"from_lang"`
				To_lang   string `json:"to_lang"`
			}{
				Ocr_lang:  Messages.OcrLanguagesList.GetCodeByName(Fields.Field.OcrLanguageCombo.Selected),
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
		widget.NewLabel("Text Translation of OCR Result:"),
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
