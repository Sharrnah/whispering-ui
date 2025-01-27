package AdditionalTextTranslations

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	dialog2 "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strconv"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket/Messages"
)

func AdditionalLanguagesCountString() string {
	// count additional languages
	numOfAdditionalLanguages := 0
	numOfAdditionalLanguagesLabelText := ""
	for _, language := range strings.Split(Settings.Config.Txt_second_translation_languages, ",") {
		if language != "" {
			numOfAdditionalLanguages++
		}
	}
	if Settings.Config.Txt_second_translation_enabled && numOfAdditionalLanguages > 0 {
		numOfAdditionalLanguagesLabelText = "(+" + strconv.Itoa(numOfAdditionalLanguages) + ")"
	}
	return numOfAdditionalLanguagesLabelText
}

func CreateLanguagesListWindow(button *widget.Button) *dialog2.CustomDialog {
	defer Utilities.PanicLogger()

	var activeLanguagesList []string

	for _, language := range strings.Split(Settings.Config.Txt_second_translation_languages, ",") {
		if language != "" {
			activeLanguagesList = append(activeLanguagesList, language)
		}
	}

	var activeLanguagesListWidget *widget.List
	activeLanguagesListWidget = widget.NewList(
		func() int {
			return len(activeLanguagesList)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				nil,
				widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
				}),
				widget.NewLabel("template"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			mainContainer := o.(*fyne.Container)
			languageLabel := mainContainer.Objects[0].(*widget.Label)
			removeButton := mainContainer.Objects[1].(*widget.Button)

			languageLabel.SetText(Messages.InstalledLanguages.GetNameByCode(activeLanguagesList[i]))
			removeButton.OnTapped = func() {
				// Remove language from activeLanguagesList
				activeLanguagesList = append(activeLanguagesList[:i], activeLanguagesList[i+1:]...)
				activeLanguagesListWidget.Refresh()

				Settings.Config.Txt_second_translation_languages = strings.Join(activeLanguagesList, ",")
				// send new list
				sendMessage := Fields.SendMessageStruct{
					Type:  "setting_change",
					Name:  "txt_second_translation_languages",
					Value: Settings.Config.Txt_second_translation_languages,
				}
				sendMessage.SendMessage()

				button.SetText(AdditionalLanguagesCountString())
			}
		},
	)

	// Create window content
	enableAdditionalTranslationCheckbox := widget.NewCheck(lang.L("Enable Additional Translations"), func(checked bool) {
		Settings.Config.Txt_second_translation_enabled = checked
		sendMessage := Fields.SendMessageStruct{
			Type:  "setting_change",
			Name:  "txt_second_translation_enabled",
			Value: checked,
		}
		sendMessage.SendMessage()
		button.SetText(AdditionalLanguagesCountString())
	})
	enableAdditionalTranslationCheckbox.Checked = Settings.Config.Txt_second_translation_enabled

	languageListWidget := CustomWidget.NewCompletionEntry(Fields.Field.TargetLanguageTxtTranslateCombo.Options)
	languageListWidget.OptionsTextValue = Fields.Field.TargetLanguageTxtTranslateCombo.OptionsTextValue

	languageListWidget.ResetOptionsFilter()

	languageListWidget.ShowAllEntryText = lang.L("... show all")
	languageListWidget.Entry.PlaceHolder = lang.L("Select target language")
	languageListWidget.OnChanged = func(value string) {
		// filter out the values of Options that do not contain the value
		var filteredValues []string
		for i := 0; i < len(languageListWidget.Options); i++ {
			if len(languageListWidget.Options) > i && strings.Contains(strings.ToLower(languageListWidget.Options[i]), strings.ToLower(value)) {
				filteredValues = append(filteredValues, languageListWidget.Options[i])
			}
		}
		languageListWidget.SetOptionsFilter(filteredValues)
		languageListWidget.ShowCompletion()
	}
	languageListWidget.OnSubmitted = func(value string) {
		// check if value is not in Options
		value = Fields.UpdateCompletionEntryBasedOnValue(languageListWidget, value)
		value = Messages.InstalledLanguages.GetCodeByName(value)
		if value == "" {
			return
		}

		// only append if not already in activeLanguagesList
		if !Utilities.Contains(activeLanguagesList, value) {
			activeLanguagesList = append(activeLanguagesList, value)
			Settings.Config.Txt_second_translation_languages = strings.Join(activeLanguagesList, ",")

			activeLanguagesListWidget.Refresh()

			sendMessage := Fields.SendMessageStruct{
				Type:  "setting_change",
				Name:  "txt_second_translation_languages",
				Value: Settings.Config.Txt_second_translation_languages,
			}
			sendMessage.SendMessage()
			button.SetText(AdditionalLanguagesCountString())
		}
	}

	targetLanguageListRow := container.New(layout.NewFormLayout(), widget.NewLabel(lang.L("Additional Translation")+":"), languageListWidget)

	beginLine := canvas.NewHorizontalGradient(&color.NRGBA{R: 198, G: 123, B: 0, A: 255}, &color.NRGBA{R: 198, G: 123, B: 0, A: 0})
	beginLine.Resize(fyne.NewSize(Fields.Field.SttEnabled.Size().Width, 2))

	content := container.NewBorder(container.NewVBox(enableAdditionalTranslationCheckbox, targetLanguageListRow, container.NewGridWithColumns(2, beginLine)), nil, nil, nil, activeLanguagesListWidget)

	mainWindow, _ := Utilities.GetCurrentMainWindow("")
	dialog := dialog2.NewCustom(lang.L("Additional Translation Languages"), lang.L("Close"), content, mainWindow)

	windowSize := Utilities.GetInlineDialogSize(fyne.NewSize(300, 150), fyne.NewSize(100, 200), fyne.NewSize(700, 500))
	dialog.Resize(windowSize)

	dialog.Show()

	content.Refresh()

	return dialog
}
