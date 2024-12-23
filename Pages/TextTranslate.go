package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Pages/AdditionalTextTranslations"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

var additionalTranslationWindow fyne.Window

func CreateTextTranslateWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	sourceLanguageRow := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Source Language")+":"), Fields.Field.SourceLanguageCombo)
	targetLanguageRow := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Target Language")+":"), Fields.Field.TargetLanguageCombo)

	switchButton := widget.NewButtonWithIcon(lang.L("Swap languages"), theme.NewThemedResource(Resources.ResourceSwapHorizontalSvg), func() {
		sourceLanguage := Fields.Field.SourceLanguageCombo.Text
		// use last detected language when switching between source and target language
		if strings.HasPrefix(strings.ToLower(sourceLanguage), "auto") && Settings.Config.Last_auto_txt_translate_lang != "" {
			sourceLanguage = Utilities.LanguageMapList.GetName(Settings.Config.Last_auto_txt_translate_lang)
		}

		targetLanguage := Fields.Field.TargetLanguageCombo.Text
		if targetLanguage == "None" {
			targetLanguage = "Auto"
		}

		Fields.Field.SourceLanguageCombo.Text = targetLanguage
		Fields.Field.SourceLanguageCombo.Refresh()
		Fields.Field.TargetLanguageCombo.Text = sourceLanguage
		Fields.Field.TargetLanguageCombo.Refresh()

		sourceField := Fields.Field.TranscriptionInput.Text
		targetField := Fields.Field.TranscriptionTranslationInput.Text
		Fields.Field.TranscriptionInput.SetText(targetField)
		Fields.Field.TranscriptionTranslationInput.SetText(sourceField)
	})
	switchButton.Importance = widget.LowImportance
	switchButton.Alignment = widget.ButtonAlignCenter
	switchButton.IconPlacement = widget.ButtonIconLeadingText
	switchButtonAligner := container.NewCenter(switchButton)

	additionalLanguagesMenuButton := widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		if additionalTranslationWindow != nil {
			additionalTranslationWindow.Close()
		}
		additionalTranslationWindow = AdditionalTextTranslations.CreateLanguagesListWindow()
	})
	languageRow := container.NewBorder(nil, nil, nil, additionalLanguagesMenuButton, container.New(layout.NewGridLayout(2), sourceLanguageRow, targetLanguageRow))

	transcriptionRow := container.New(layout.NewGridLayout(2),
		container.NewBorder(nil, Fields.Field.TranscriptionInputHint, nil, nil, Fields.Field.TranscriptionInput),
		container.NewBorder(nil, Fields.Field.TranscriptionTranslationInputHint, nil, nil, Fields.Field.TranscriptionTranslationInput),
	)

	translateOnlyFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Text)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Text)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				To_romaji           bool   `json:"to_romaji"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                Fields.Field.TranscriptionInput.Text,
				From_lang:           fromLang,
				To_lang:             toLang,
				To_romaji:           Settings.Config.Txt_romaji,
				Ignore_send_options: true,
			},
		}
		sendMessage.SendMessage()
	}
	translateOnlyButton := widget.NewButtonWithIcon(lang.L("Translate Only[CTRL+ALT+Enter]"), theme.MenuExpandIcon(), translateOnlyFunction)

	translateFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Text)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Text)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := Fields.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				To_romaji           bool   `json:"to_romaji"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                Fields.Field.TranscriptionInput.Text,
				From_lang:           fromLang,
				To_lang:             toLang,
				To_romaji:           Settings.Config.Txt_romaji,
				Ignore_send_options: false,
			},
		}
		sendMessage.SendMessage()
	}
	translateButton := widget.NewButtonWithIcon(lang.L("Translate (and send)[CTRL+Enter]"), theme.ConfirmIcon(), translateFunction)
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
			switchButtonAligner,
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
