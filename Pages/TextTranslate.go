package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

func CreateTextTranslateWindow() fyne.CanvasObject {

	sourceLanguageRow := container.New(layout.NewFormLayout(), widget.NewLabel("Source Language:"), Fields.Field.SourceLanguageCombo)
	targetLanguageRow := container.New(layout.NewFormLayout(), widget.NewLabel("Target Language:"), Fields.Field.TargetLanguageCombo)

	switchButton := container.NewCenter(widget.NewButton("<==>", func() {
		sourceLanguage := Fields.Field.SourceLanguageCombo.Selected
		// use last detected language when switching between source and target language
		if strings.HasPrefix(sourceLanguage, "Auto") && Settings.Config.Last_auto_txt_translate_lang != "" {
			sourceLanguage = Utilities.LanguageMapList.GetName(Settings.Config.Last_auto_txt_translate_lang)
		}

		targetLanguage := Fields.Field.TargetLanguageCombo.Selected
		if targetLanguage == "None" {
			targetLanguage = "Auto"
		}

		Fields.Field.SourceLanguageCombo.SetSelected(targetLanguage)
		Fields.Field.TargetLanguageCombo.SetSelected(sourceLanguage)

		sourceField := Fields.Field.TranscriptionInput.Text
		targetField := Fields.Field.TranscriptionTranslationInput.Text
		Fields.Field.TranscriptionInput.SetText(targetField)
		Fields.Field.TranscriptionTranslationInput.SetText(sourceField)
	}))

	languageRow := container.New(layout.NewGridLayout(2), sourceLanguageRow, targetLanguageRow)

	transcriptionRow := container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionInput, Fields.Field.TranscriptionTranslationInput)

	translateButton := widget.NewButtonWithIcon("translate", theme.ConfirmIcon(), func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Selected)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Selected)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text      string `json:"text"`
				From_lang string `json:"from_lang"`
				To_lang   string `json:"to_lang"`
			}{
				Text:      Fields.Field.TranscriptionInput.Text,
				From_lang: fromLang,
				To_lang:   toLang,
			},
		}
		sendMessage.SendMessage()
	})
	translateButton.Importance = widget.HighImportance

	translateButtonRow := container.NewHBox(layout.NewSpacer(),
		translateButton,
	)

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			languageRow,
			switchButton,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.New(layout.NewVBoxLayout(), translateButtonRow),
		),
	)

	return mainContent
}
