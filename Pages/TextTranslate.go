package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
	"whispering-tiger-ui/CustomWidget"
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

	transcriptionRow := container.New(layout.NewGridLayout(2),
		container.NewBorder(nil, Fields.Field.TranscriptionInputHint, nil, nil, Fields.Field.TranscriptionInput),
		container.NewBorder(nil, Fields.Field.TranscriptionTranslationInputHint, nil, nil, Fields.Field.TranscriptionTranslationInput),
	)

	translateOnlyFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Selected)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Selected)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                Fields.Field.TranscriptionInput.Text,
				From_lang:           fromLang,
				To_lang:             toLang,
				Ignore_send_options: true,
			},
		}
		sendMessage.SendMessage()
	}
	translateOnlyButton := widget.NewButtonWithIcon("Translate Only\n[CTRL+ALT+Enter]", theme.MenuExpandIcon(), translateOnlyFunction)

	translateFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Selected)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Selected)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                Fields.Field.TranscriptionInput.Text,
				From_lang:           fromLang,
				To_lang:             toLang,
				Ignore_send_options: false,
			},
		}
		sendMessage.SendMessage()
	}
	translateButton := widget.NewButtonWithIcon("Translate (and send)\n[CTRL+Enter]", theme.ConfirmIcon(), translateFunction)
	translateButton.Importance = widget.HighImportance

	// quick options row
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
		Fields.Field.TtsEnabled,
		container.NewBorder(nil, nil, nil, Fields.Field.OscLimitHint, Fields.Field.OscEnabled),
	)

	translateButtonRow := container.NewHBox(container.NewBorder(nil, nil, quickOptionsRow, nil), layout.NewSpacer(),
		translateOnlyButton,
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

	// add shortcuts to source text field
	translateShortcut := CustomWidget.ShortcutEntrySubmit{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl,
		Handler:  translateFunction,
	}
	Fields.Field.TranscriptionInput.AddCustomShortcut(translateShortcut)

	translateOnlyShortcut := CustomWidget.ShortcutEntrySubmit{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierAlt,
		Handler: func() {
			if mainContent.Visible() {
				translateOnlyFunction()
			}
		},
	}
	Fields.Field.TranscriptionInput.AddCustomShortcut(translateOnlyShortcut)

	// add shortcuts to target text field
	Fields.Field.TranscriptionTranslationInput.AddCustomShortcut(translateOnlyShortcut)

	return mainContent
}
