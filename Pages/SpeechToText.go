package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Websocket/Messages"
)

func CreateSpeechToTextWindow() fyne.CanvasObject {
	//widget.NewSeparator()
	//layout.NewSpacer()

	//lastDetectedLanguage := widget.NewLabel("Last detected Language: ")

	//LanguageRow := container.New(layout.NewHBoxLayout(), widget.NewLabel("Target Language: "), Fields.Field.TargetLanguageCombo)

	languageRow := container.New(layout.NewFormLayout(), widget.NewLabel("Speech Task:"), container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionTaskCombo, Fields.Field.TranscriptionSpeakerLanguageCombo), widget.NewLabel("Target Language:"), Fields.Field.TargetLanguageCombo)

	transcriptionRow := container.New(layout.NewGridLayout(2), Fields.Field.TranscriptionInput, Fields.Field.TranscriptionTranslationInput)

	// quick options row
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
		Fields.Field.TtsEnabled,
		Fields.Field.OscEnabled,
	)

	// main layout
	leftVerticalLayout := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			languageRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.New(layout.NewVBoxLayout(), quickOptionsRow),
		),
	)

	Fields.Field.ProcessingStatus = widget.NewProgressBarInfinite()

	// whisper Result list
	Fields.Field.WhisperResultList = widget.NewListWithData(Fields.DataBindings.WhisperResultsDataBinding,
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
			value := i.(binding.Untyped)
			whisperMessage, _ := value.Get()

			result := whisperMessage.(Messages.WhisperResult)

			translateResultBind := binding.NewString()
			translateResultBind.Set(result.TxtTranslation)

			translateResultLanguageBind := binding.NewString()
			translateResultLanguageBind.Set("[" + result.TxtTranslationTarget + "]")

			originalTranscriptBind := binding.NewString()
			originalTranscriptBind.Set(result.Text)

			originalTranscriptLanguageBind := binding.NewString()
			originalTranscriptLanguageBind.Set("[" + result.Language + "]")

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

			// bind data to elements if no translation is generated (sets transcription to top label)
			if result.TxtTranslation == "" {
				translateResultLabel.Bind(originalTranscriptBind)
				translateResultLanguageLabel.Bind(originalTranscriptLanguageBind)

				originalTranscriptionLabel.SetText("")
				originalTranscriptionLanguageLabel.SetText("")
			} else { // bind data to elements if translation was generated
				translateResultLabel.Bind(translateResultBind)
				translateResultLanguageLabel.Bind(translateResultLanguageBind)

				originalTranscriptionLabel.Bind(originalTranscriptBind)
				originalTranscriptionLanguageLabel.Bind(originalTranscriptLanguageBind)
			}

			// resize
			//mainContainer.Resize(fyne.NewSize(mainContainer.Size().Width, translateResultLabel.Size().Height+originalTranscriptionLabel.Size().Height+10))
		})

	Fields.Field.WhisperResultList.OnSelected = func(id widget.ListItemID) {
		whisperMessage, _ := Fields.DataBindings.WhisperResultsDataBinding.GetValue(id)

		result := whisperMessage.(Messages.WhisperResult)

		Fields.Field.TranscriptionInput.SetText(result.Text)
		if result.TxtTranslation != "" {
			Fields.Field.TranscriptionTranslationInput.SetText(result.TxtTranslation)
		} else {
			Fields.Field.TranscriptionTranslationInput.SetText(result.Text)
		}

		go func() {
			time.Sleep(200 * time.Millisecond)
			Fields.Field.WhisperResultList.Unselect(id)
		}()
	}

	mainContent := container.NewHSplit(
		leftVerticalLayout,
		container.NewMax(
			container.NewBorder(nil, Fields.Field.ProcessingStatus, nil, nil, Fields.Field.WhisperResultList),
		),
	)

	Fields.Field.ProcessingStatus.Stop()

	mainContent.SetOffset(0.6)

	return mainContent
}
