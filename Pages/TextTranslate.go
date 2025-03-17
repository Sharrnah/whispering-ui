package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/AdditionalTextTranslations"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

var additionalTranslationWindow dialog.Dialog

func CreateTextTranslateWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\TextTranslate->CreateTextTranslateWindow")
	})

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

		sourceField, _ := Fields.DataBindings.TranscriptionInputBinding.Get()
		targetField, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
		Fields.DataBindings.TranscriptionInputBinding.Set(targetField)
		Fields.DataBindings.TranscriptionTranslationInputBinding.Set(sourceField)
	})
	switchButton.Importance = widget.LowImportance
	switchButton.Alignment = widget.ButtonAlignCenter
	switchButton.IconPlacement = widget.ButtonIconLeadingText
	switchButtonAligner := container.NewCenter(switchButton)

	numOfAdditionalLanguagesLabelText := Fields.AdditionalLanguagesCountString("", "()")
	additionalLanguagesMenuButton := widget.NewButtonWithIcon(numOfAdditionalLanguagesLabelText, theme.ListIcon(), nil)
	additionalLanguagesMenuButton.OnTapped = func() {
		if additionalTranslationWindow != nil {
			additionalTranslationWindow.Hide()
		}
		additionalTranslationWindow = AdditionalTextTranslations.CreateLanguagesListWindow(additionalLanguagesMenuButton)
	}
	languageRow := container.NewBorder(nil, nil, nil, additionalLanguagesMenuButton, container.New(layout.NewGridLayout(2), sourceLanguageRow, targetLanguageRow))

	transcriptionRow := container.New(layout.NewGridLayout(2),
		container.NewBorder(nil, Fields.Field.TranscriptionInputHintOnTxtTranslate, nil, nil, Fields.Field.TranscriptionTextTranslationInput),
		container.NewBorder(nil, Fields.Field.TranscriptionTranslationInputHintOnTxtTranslate, nil, nil, Fields.Field.TranscriptionTranslationTextTranslationInput),
	)

	// special case for models that only allow setting target language
	if Settings.Config.Txt_translator == "phi4" {
		Fields.Field.SourceLanguageCombo.Disable()
		Fields.Field.SourceLanguageCombo.Text = "auto"
		Fields.Field.SourceLanguageCombo.Refresh()

		switchButton.Disable()
	}

	translateOnlyFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageCombo.Text)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageCombo.Text)
		text, _ := Fields.DataBindings.TranscriptionInputBinding.Get()
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				To_romaji           bool   `json:"to_romaji"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                text,
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
		text, _ := Fields.DataBindings.TranscriptionInputBinding.Get()
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type: "translate_req",
			Value: struct {
				Text                string `json:"text"`
				From_lang           string `json:"from_lang"`
				To_lang             string `json:"to_lang"`
				To_romaji           bool   `json:"to_romaji"`
				Ignore_send_options bool   `json:"ignore_send_options"`
			}{
				Text:                text,
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
		Fields.Field.TtsEnabledOnTxtTranslate,
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
	Fields.Field.TranscriptionTextTranslationInput.AddCustomShortcut(translateShortcut)

	translateOnlyShortcut := CustomWidget.ShortcutEntrySubmit{
		KeyName:  fyne.KeyReturn,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierAlt,
		Handler: func() {
			if mainContent.Visible() {
				translateOnlyFunction()
			}
		},
	}
	Fields.Field.TranscriptionTextTranslationInput.AddCustomShortcut(translateOnlyShortcut)

	// add shortcuts to target text field
	Fields.Field.TranscriptionTranslationTextTranslationInput.AddCustomShortcut(translateOnlyShortcut)

	return mainContent
}
