package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Websocket/Messages"
)

func CreateSpeechToTextWindow() fyne.CanvasObject {
	languageRow := container.New(layout.NewVBoxLayout(),
		container.New(layout.NewFormLayout(),
			widget.NewLabel("Speech Language:"),
			Fields.Field.TranscriptionSpeakerLanguageCombo,

			widget.NewLabel("Speech Task:"),
			Fields.Field.TranscriptionTaskCombo,
		),
	)

	transcriptionRow := container.New(
		layout.NewGridLayout(2),
		container.NewBorder(nil, Fields.Field.TranscriptionInputHint, nil, nil, Fields.Field.TranscriptionInput),
		container.NewBorder(nil, Fields.Field.TranscriptionTranslationInputHint, nil, nil, Fields.Field.TranscriptionTranslationInput),
	)

	// quick options row
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
		Fields.Field.SttEnabled,
		Fields.Field.TextTranslateEnabled,
		Fields.Field.TtsEnabled,
		container.NewHBox(
			container.NewBorder(nil, nil, nil, Fields.Field.OscLimitHint, Fields.Field.OscEnabled),
		),
	)

	// main layout
	leftVerticalLayout := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			languageRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			container.New(layout.NewHBoxLayout(), quickOptionsRow),
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

	if !Settings.Config.Realtime {
		Fields.Field.RealtimeResultLabel.Hide()
	}
	realtimeWhisperResultBlock := container.NewBorder(
		nil, container.NewVBox(widget.NewSeparator(), widget.NewSeparator()), nil, nil,
		Fields.Field.RealtimeResultLabel,
	)
	whisperResultContainer := container.NewMax(
		container.NewBorder(
			realtimeWhisperResultBlock, Fields.Field.ProcessingStatus, nil, nil,
			Fields.Field.WhisperResultList,
		),
	)

	mainContent := container.NewHSplit(
		leftVerticalLayout,
		container.NewMax(
			whisperResultContainer,
		),
	)

	Fields.Field.ProcessingStatus.Stop()

	mainContent.SetOffset(0.6)

	return mainContent
}
