package Pages

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/websocket/Messages"
)

func CreateMainWindow() fyne.CanvasObject {
	title := container.New(layout.NewCenterLayout(), widget.NewLabel("Main"))

	//widget.NewSeparator()
	//layout.NewSpacer()

	//lastDetectedLanguage := widget.NewLabel("Last detected Language: ")

	//LanguageRow := container.New(layout.NewHBoxLayout(), widget.NewLabel("Target Language: "), Fields.Field.TargetLanguageCombo)

	LanguageRow := container.New(layout.NewFormLayout(), widget.NewLabel("Speech Task:"), container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionTaskCombo, Fields.Field.TranscriptionSpeakerLanguageCombo), widget.NewLabel("Target Language:"), Fields.Field.TargetLanguageCombo)

	transcriptionRow := container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionInput, Fields.Field.TranscriptionTranslation)

	// quick options row
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
		Fields.Field.TtsEnabled,
		Fields.Field.OscEnabled,
	)

	// whisper results row
	//Messages.WhisperResultsStringList = []string{}
	resultList := widget.NewListWithData(Fields.DataBindings.WhisperResultsDataBinding,
		func() fyne.CanvasObject {
			return container.New(layout.NewGridLayout(1),
				container.NewBorder(
					nil,
					nil,
					nil,
					widget.NewLabelWithStyle("[ResultLang]", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
					widget.NewLabelWithStyle("TranslateResult", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				),
				container.NewBorder(
					nil,
					nil,
					nil,
					widget.NewLabelWithStyle("[ResultLang]", fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
					widget.NewLabel("Transcription"),
				),
			)
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			//o.(*widget.Label).Bind(i.(binding.String))
			value := i.(binding.String)
			jsonStringValue, _ := value.Get()

			var jsonResult = Messages.WhisperResult{}
			json.Unmarshal([]byte(jsonStringValue), &jsonResult)

			//values := strings.Split(stringValue, "###")

			translateResultBind := binding.NewString()
			translateResultBind.Set(jsonResult.TxtTranslation)

			translateResultLanguageBind := binding.NewString()
			translateResultLanguageBind.Set("[" + jsonResult.TxtTranslationTarget + "]")

			originalTranscriptBind := binding.NewString()
			originalTranscriptBind.Set(jsonResult.Text)

			originalTranscriptLanguageBind := binding.NewString()
			originalTranscriptLanguageBind.Set("[" + jsonResult.Language + "]")

			// get all template elements
			mainContainer := o.(*fyne.Container)
			finalTranslationContainer := mainContainer.Objects[0].(*fyne.Container)
			originalTranscriptionContainer := mainContainer.Objects[1].(*fyne.Container)

			translateResultLabel := finalTranslationContainer.Objects[0].(*widget.Label)
			translateResultLabel.Wrapping = fyne.TextWrapWord
			translateResultLanguageLabel := finalTranslationContainer.Objects[1].(*widget.Label)

			originalTranscriptionLabel := originalTranscriptionContainer.Objects[0].(*widget.Label)
			originalTranscriptionLabel.Wrapping = fyne.TextWrapWord
			originalTranscriptionLanguageLabel := originalTranscriptionContainer.Objects[1].(*widget.Label)

			// bind data to elements
			translateResultLabel.Bind(translateResultBind)
			translateResultLanguageLabel.Bind(translateResultLanguageBind)

			originalTranscriptionLabel.Bind(originalTranscriptBind)
			originalTranscriptionLanguageLabel.Bind(originalTranscriptLanguageBind)

			// set to top label if text translation is empty
			if jsonResult.TxtTranslation == "" {
				translateResultLabel.Bind(originalTranscriptBind)
				translateResultLanguageLabel.Bind(originalTranscriptLanguageBind)

				originalTranscriptionLabel.Unbind()
				originalTranscriptionLabel.Bind(binding.NewString())
				originalTranscriptionLanguageLabel.Unbind()
				originalTranscriptionLanguageLabel.Bind(binding.NewString())
			}

			// resize
			//fyne.MeasureText(jsonResult.TxtTranslation, 12, fyne.TextStyle{Bold: true})
			//mainContainer.Resize(fyne.NewSize(mainContainer.Size().Width, translateResultLabel.Size().Height+originalTranscriptionLabel.Size().Height+10))
		})

	resultList.OnSelected = func(id widget.ListItemID) {
		selectedJsonData, _ := Fields.DataBindings.WhisperResultsDataBinding.GetValue(id)

		var jsonResult = Messages.WhisperResult{}
		json.Unmarshal([]byte(selectedJsonData), &jsonResult)

		Fields.Field.TranscriptionInput.SetText(jsonResult.Text)
		Fields.Field.TranscriptionTranslation.SetText(jsonResult.TxtTranslation)
	}

	// split between transcription + options row
	splitTranscriptionOptions := container.NewVSplit(transcriptionRow, quickOptionsRow)
	//splitTranscriptionOptions.Offset = 0.5

	// main layout
	verticalLayout := container.New(layout.NewVBoxLayout(),
		title,
		LanguageRow,
		splitTranscriptionOptions,
	)

	mainContent := container.NewHSplit(
		verticalLayout,
		container.NewMax(resultList),
	)

	return mainContent
}
