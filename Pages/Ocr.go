package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"golang.design/x/clipboard"
	"log"
	"strings"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

// Guess the language from the OCR language selection if auto to lessen the language guessing
func guessTranslationFromLanguage(ocrLanguageCode string) string {
	fromLang := ""
	if len(Fields.Field.SourceLanguageTxtTranslateCombo.OptionsTextValue) > 0 {
		fromLang = Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageTxtTranslateCombo.GetValueOptionEntryByText(Fields.Field.SourceLanguageTxtTranslateCombo.Text).Value)
	} else {
		fromLang = Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageTxtTranslateCombo.Text)
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
	return fromLang
}

func GetClipboardImage() ([]byte, clipboard.Format) {
	var clipboardBinary []byte
	err := clipboard.Init()
	if err == nil {
		Logging.CaptureException(err)
		clipboardBinary = clipboard.Read(clipboard.FmtImage)
		if clipboardBinary != nil {
			return clipboardBinary, clipboard.FmtImage
		}
		clipboardBinary = clipboard.Read(clipboard.FmtText)
		if clipboardBinary != nil {
			return clipboardBinary, clipboard.FmtText
		}
	}
	return nil, -1
}

func CreateOcrWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Ocr->CreateOcrWindow")
	})

	translateOnlyFunction := func() {
		fromLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.SourceLanguageTxtTranslateCombo.Text)
		if fromLang == "" {
			fromLang = "auto"
		}
		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageTxtTranslateCombo.Text)
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
			valueObj := Fields.Field.SourceLanguageTxtTranslateCombo.GetValueOptionEntryByText(value)
			value = valueObj.Value

			valueIso = Messages.OcrLanguagesList.GetCodeByName(value)
		}

		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "ocr_lang",
			Value: valueIso,
		}
		sendMessage.SendMessage()

		log.Println("ocr Select set to", value)
	}

	container.New(layout.NewStackLayout())
	ocrLanguageWindowForm := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Text in Image Language")+":"), Fields.Field.OcrLanguageCombo, widget.NewLabel(lang.L("Window")+":"), Fields.Field.OcrWindowCombo)

	ocrSettingsRow := container.New(layout.NewGridLayout(1), ocrLanguageWindowForm)

	ocrButton := widget.NewButtonWithIcon(lang.L("Window Scan & Translate"), theme.ConfirmIcon(), func() {

		ocrLanguageCode := Messages.OcrLanguagesList.GetCodeByName(Fields.Field.OcrLanguageCombo.Text)

		fromLang := guessTranslationFromLanguage(ocrLanguageCode)

		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageTxtTranslateCombo.Text)
		//goland:noinspection GoSnakeCaseUsage
		sendMessage := SendMessageChannel.SendMessageStruct{
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

	ocrClipboardButtonRow := widget.NewButtonWithIcon(lang.L("Clipboard Scan & Translate"), theme.ContentPasteIcon(), func() {
		clipboardData, clipboardFormat := GetClipboardImage()
		if clipboardData == nil {
			return
		}

		ocrLanguageCode := Messages.OcrLanguagesList.GetCodeByName(Fields.Field.OcrLanguageCombo.Text)

		fromLang := guessTranslationFromLanguage(ocrLanguageCode)

		toLang := Messages.InstalledLanguages.GetCodeByName(Fields.Field.TargetLanguageTxtTranslateCombo.Text)
		if clipboardFormat == clipboard.FmtImage {
			//goland:noinspection GoSnakeCaseUsage
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "ocr_req",
				Value: struct {
					Image     []byte `json:"image"`
					Ocr_lang  string `json:"ocr_lang"`
					From_lang string `json:"from_lang"`
					To_lang   string `json:"to_lang"`
				}{
					Image:     clipboardData,
					Ocr_lang:  ocrLanguageCode,
					From_lang: fromLang,
					To_lang:   toLang,
				},
			}
			sendMessage.SendMessage()
		}
		if clipboardFormat == clipboard.FmtText {
			clipboardText := string(clipboardData)
			Fields.DataBindings.TranscriptionInputBinding.Set(clipboardText)
			translateOnlyFunction()
			Fields.Field.OcrImageContainer.RemoveAll()
		}
	})

	buttonRow := container.NewHBox(layout.NewSpacer(),
		ocrClipboardButtonRow,
		ocrButton,
	)

	switchButton := widget.NewButtonWithIcon(lang.L("Swap languages"), theme.NewThemedResource(Resources.ResourceSwapHorizontalSvg), func() {
		sourceLanguage := Fields.Field.SourceLanguageTxtTranslateCombo.Text
		// use last detected language when switching between source and target language
		if strings.HasPrefix(strings.ToLower(sourceLanguage), "auto") && Settings.Config.Last_auto_txt_translate_lang != "" {
			sourceLanguage = Utilities.LanguageMapList.GetName(Settings.Config.Last_auto_txt_translate_lang)
		}

		targetLanguage := Fields.Field.TargetLanguageTxtTranslateCombo.Text
		if targetLanguage == "None" {
			targetLanguage = "Auto"
		}

		Fields.Field.SourceLanguageTxtTranslateCombo.Text = targetLanguage
		Fields.Field.SourceLanguageTxtTranslateCombo.Refresh()
		Fields.Field.TargetLanguageTxtTranslateCombo.Text = sourceLanguage
		Fields.Field.TargetLanguageTxtTranslateCombo.Refresh()

		sourceField, _ := Fields.DataBindings.TranscriptionInputBinding.Get()
		targetField, _ := Fields.DataBindings.TranscriptionTranslationInputBinding.Get()
		Fields.DataBindings.TranscriptionInputBinding.Set(targetField)
		Fields.Field.TranscriptionTranslationSpeechToTextInput.SetText(sourceField)
	})
	switchButton.Importance = widget.LowImportance
	switchButton.Alignment = widget.ButtonAlignCenter
	switchButton.IconPlacement = widget.ButtonIconLeadingText
	switchButtonAligner := container.NewCenter(switchButton)

	sourceLanguageForm := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Source Language")+":"), Fields.Field.SourceLanguageTxtTranslateCombo)
	targetLanguageForm := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Target Language")+":"), Fields.Field.TargetLanguageTxtTranslateCombo)
	languageRow := container.New(layout.NewGridLayout(2), sourceLanguageForm, targetLanguageForm)

	transcriptionRow := container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionOcrInput, Fields.Field.TranscriptionTranslationOcrInput)

	translateOnlyButton := widget.NewButtonWithIcon(lang.L("Translate Only"), theme.MenuExpandIcon(), translateOnlyFunction)

	ocrContent := container.New(layout.NewVBoxLayout(),
		ocrSettingsRow,
		container.New(layout.NewPaddedLayout(), buttonRow),
		widget.NewSeparator(),
		widget.NewLabel(lang.L("Text-Translation of OCR Result")+":"),
		languageRow,
		switchButtonAligner,
	)

	mainContent := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			ocrContent,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.NewBorder(
				container.NewBorder(
					nil, nil, nil, translateOnlyButton,
				),
				nil, nil, nil, Fields.Field.OcrImageContainer,
			),
		),
	)

	return mainContent
}
