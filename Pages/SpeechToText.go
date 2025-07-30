package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func CreateSpeechToTextWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\SpeechToText->CreateSpeechToTextWindow")
	})

	var additionalWidgets fyne.CanvasObject

	var originalTranscriptionSpeakerLanguageComboEntries = Fields.Field.TranscriptionSpeakerLanguageCombo.OptionsTextValue
	// remove "auto" option from speaker language combo, as it is not supported by all models
	var originalTranscriptionSpeakerLanguageComboEntriesWithoutAuto = []CustomWidget.TextValueOption{}
	for _, entry := range originalTranscriptionSpeakerLanguageComboEntries {
		if entry.Value != "auto" && entry.Value != "" {
			originalTranscriptionSpeakerLanguageComboEntriesWithoutAuto = append(originalTranscriptionSpeakerLanguageComboEntriesWithoutAuto, entry)
		}
	}

	speechLanguageLabel := widget.NewLabel(lang.L("Speech Language") + ":")

	speechTaskWidgetLabel := widget.NewLabel(lang.L("Speech Task") + ":")
	var speechTaskWidget fyne.CanvasObject = Fields.Field.TranscriptionTaskCombo
	// disable task for seamless_m4t model, as it always translates to target language (Speech Language)
	if Settings.Config.Stt_type == "seamless_m4t" || Settings.Config.Stt_type == "nemo_canary" {
		speechTaskWidgetLabel.SetText(lang.L("Target Language") + ":")
		speechTaskWidget = Fields.Field.TranscriptionTargetLanguageCombo
	}
	if Settings.Config.Stt_type == "phi4" || Settings.Config.Stt_type == "phi4-onnx" {
		speechLanguageLabel.SetText(lang.L("Target Language") + ":")

		speechTaskWidget.(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{{
			Text:  lang.L("transcribe"),
			Value: "transcribe",
		}, {
			Text:  lang.L("translate"),
			Value: "translate",
		}, {
			Text:  lang.L("transcribe & translate"),
			Value: "transcribe_translate",
		}, {
			Text:  lang.L("question & answering"),
			Value: "question_answering",
		}, {
			Text:  lang.L("function calling"),
			Value: "function_calling",
		}}

		// set initial task to value of loaded profile configuration
		settingsTask := speechTaskWidget.(*CustomWidget.TextValueSelect).GetEntry(&CustomWidget.TextValueOption{
			Value: Settings.Config.Whisper_task,
		}, CustomWidget.CompareValue)
		if settingsTask != nil {
			speechTaskWidget.(*CustomWidget.TextValueSelect).Selected = settingsTask.Text
		}

		oldOnChangeFunc := speechTaskWidget.(*CustomWidget.TextValueSelect).OnChanged
		speechTaskWidget.(*CustomWidget.TextValueSelect).OnChanged = func(s CustomWidget.TextValueOption) {
			if s.Value == "transcribe" || s.Value == "question_answering" || s.Value == "function_calling" {
				Fields.Field.TranscriptionSpeakerLanguageCombo.Disable()
			} else {
				Fields.Field.TranscriptionSpeakerLanguageCombo.Enable()
			}
			additionalWidgets.Hide()
			if s.Value == "question_answering" || s.Value == "chat" || s.Value == "function_calling" {
				additionalWidgets.Show()
			}
			oldOnChangeFunc(s)
		}

		additionalWidgets = container.New(
			layout.NewGridLayout(2),
			widget.NewButtonWithIcon(lang.L("Chat"), theme.MailSendIcon(), func() {
				text, _ := Fields.DataBindings.TranscriptionInputBinding.Get()
				task := speechTaskWidget.(*CustomWidget.TextValueSelect).GetSelected().Value
				sendMessage := SendMessageChannel.SendMessageStruct{
					Type: "chat_req",
					Value: struct {
						Text *string `json:"text"`
						Task *string `json:"task"`
					}{
						Text: &text,
						Task: &task,
					},
				}
				sendMessage.SendMessage()
			}),
		)
		if Settings.Config.Whisper_task != "question_answering" && Settings.Config.Whisper_task != "chat" && Settings.Config.Whisper_task != "function_calling" {
			additionalWidgets.Hide()
		}
	}
	if Settings.Config.Stt_type == "wav2vec_bert" {
		speechTaskWidgetLabel.SetText("")
		speechTaskWidget.Hide()
	}
	if Settings.Config.Stt_type == "mms" {
		speechTaskWidgetLabel.SetText("")
		speechTaskWidget.Hide()
	}
	if (Settings.Config.Stt_type == "faster_whisper" || Settings.Config.Stt_type == "transformer_whisper" || Settings.Config.Stt_type == "original_whisper") &&
		strings.HasSuffix(Settings.Config.Model, "-turbo") {
		speechTaskWidgetLabel.SetText("")
		speechTaskWidget.(*CustomWidget.TextValueSelect).Selected = "transcribe"
		Settings.Config.Whisper_task = "transcribe"
		speechTaskWidget.(*CustomWidget.TextValueSelect).Hide()
	}
	if Settings.Config.Stt_type == "voxtral" {
		speechTaskWidget.(*CustomWidget.TextValueSelect).Options = []CustomWidget.TextValueOption{{
			Text:  lang.L("transcribe"),
			Value: "transcribe",
		}, {
			Text:  lang.L("translate"),
			Value: "translate",
		}, {
			Text:  lang.L("question & answering"),
			Value: "question_answering",
		}}

		// set initial task to value of loaded profile configuration
		settingsTask := speechTaskWidget.(*CustomWidget.TextValueSelect).GetEntry(&CustomWidget.TextValueOption{
			Value: Settings.Config.Whisper_task,
		}, CustomWidget.CompareValue)
		if settingsTask != nil {
			speechTaskWidget.(*CustomWidget.TextValueSelect).Selected = settingsTask.Text
		}

		oldOnChangeFunc := speechTaskWidget.(*CustomWidget.TextValueSelect).OnChanged
		speechTaskWidget.(*CustomWidget.TextValueSelect).OnChanged = func(s CustomWidget.TextValueOption) {
			speechLanguageLabel.SetText(lang.L("Speech Language") + ":")
			Fields.Field.TranscriptionSpeakerLanguageCombo.SetValueOptions(originalTranscriptionSpeakerLanguageComboEntries)

			if s.Value == "question_answering" {
				Fields.Field.TranscriptionSpeakerLanguageCombo.Disable()
			} else if s.Value == "translate" {
				Fields.Field.TranscriptionSpeakerLanguageCombo.Enable()
				Fields.Field.TranscriptionSpeakerLanguageCombo.SetValueOptions(originalTranscriptionSpeakerLanguageComboEntriesWithoutAuto)
				speechLanguageLabel.SetText(lang.L("Target Language") + ":")
			} else {
				Fields.Field.TranscriptionSpeakerLanguageCombo.Enable()
			}
			additionalWidgets.Hide()
			if s.Value == "question_answering" {
				additionalWidgets.Show()
			}
			oldOnChangeFunc(s)
		}

		chatEntry := widget.NewEntry()
		chatEntry.OnChanged = func(s string) {
			task := "update_llm_prompt"
			allowEmpty := true
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "chat_req",
				Value: struct {
					Text       *string `json:"text"`
					Task       *string `json:"task"`
					AllowEmpty *bool   `json:"allow_empty"`
				}{
					Text:       &s,
					Task:       &task,
					AllowEmpty: &allowEmpty, // Also update when the text is empty
				},
			}
			sendMessage.SendMessageDebounced()
		}
		chatButton := widget.NewButtonWithIcon(lang.L("Chat"), theme.MailSendIcon(), func() {
			text := chatEntry.Text
			task := speechTaskWidget.(*CustomWidget.TextValueSelect).GetSelected().Value
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "chat_req",
				Value: struct {
					Text *string `json:"text"`
					Task *string `json:"task"`
				}{
					Text: &text,
					Task: &task,
				},
			}
			sendMessage.SendMessage()
		})
		presetSelection := widget.NewSelect([]string{
			"",
			"Answer very shortly.",
			"Only Translate audio into Japanese. Just write the translation without explanations.",
			"Transcribe the audio. respond with both original language and French separated by linebreak. Just write both translations without explanations.",
			"Describe the emotional tone of what you hear in 1 or 2 words. nothing else. do not answer.",
			"Write what you hear as some emojis. Just write the emojis without explanations. at most 1 - 5 emojis.",
			"Describe the speakers you hear.",
			"Write a summary of what you hear.",
			"You are Taira, a wise, charismatic, fully-sentient female tiger who can speak perfect human languages while retaining unmistakably feline instincts. You start each sentence with 'Roar!'. Keep the Answers short and concise.",
		}, func(s string) {
			chatEntry.SetText(s)
		})
		additionalWidgets = container.NewVBox(
			container.NewBorder(nil, nil, widget.NewLabel("Prompt Presets"), nil,
				presetSelection,
			),
			container.NewBorder(
				nil, nil, nil, chatButton,
				chatEntry,
			),
		)
		chatEntry.SetText(Settings.Config.Stt_llm_prompt) // Set initial chat entry text from configuration
		if Settings.Config.Whisper_task == "question_answering" {
			Fields.Field.TranscriptionSpeakerLanguageCombo.Disable()
		} else if Settings.Config.Whisper_task == "translate" {
			Fields.Field.TranscriptionSpeakerLanguageCombo.Enable()
			Fields.Field.TranscriptionSpeakerLanguageCombo.SetValueOptions(originalTranscriptionSpeakerLanguageComboEntriesWithoutAuto)
			speechLanguageLabel.SetText(lang.L("Target Language") + ":")
		} else {
			Fields.Field.TranscriptionSpeakerLanguageCombo.Enable()
			additionalWidgets.Hide()
		}
	}

	languageRow := container.New(layout.NewVBoxLayout(),
		container.New(layout.NewFormLayout(),
			speechLanguageLabel,
			Fields.Field.TranscriptionSpeakerLanguageCombo,

			speechTaskWidgetLabel,
			speechTaskWidget,
		),
	)

	transcriptionRow := container.New(
		layout.NewGridLayout(2),
		container.NewBorder(nil, Fields.Field.TranscriptionInputHint, nil, nil, Fields.Field.TranscriptionSpeechToTextInput),
		container.NewBorder(nil, Fields.Field.TranscriptionTranslationInputHint, nil, nil, Fields.Field.TranscriptionTranslationSpeechToTextInput),
	)

	beginLine := canvas.NewHorizontalGradient(&color.NRGBA{R: 198, G: 123, B: 0, A: 255}, &color.NRGBA{R: 198, G: 123, B: 0, A: 0})
	beginLine.Resize(fyne.NewSize(Fields.Field.SttEnabled.Size().Width, 2))

	// quick options row
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
	)

	quickOptionsRow.Add(Fields.Field.SttEnabled)
	quickOptionsRow.Add(container.NewGridWithColumns(2, beginLine))
	quickOptionsRow.Add(Fields.Field.TextTranslateEnabled)
	quickOptionsRow.Add(Fields.Field.TtsEnabledOnStt)
	quickOptionsRow.Add(container.NewHBox(
		container.NewBorder(nil, nil, nil, Fields.Field.OscLimitHint, Fields.Field.OscEnabled),
	))

	leftVerticalBottomLayout := container.New(layout.NewVBoxLayout())
	if additionalWidgets != nil {
		leftVerticalBottomLayout.Add(additionalWidgets)
	}
	leftVerticalBottomLayout.Add(quickOptionsRow)

	// main layout
	leftVerticalLayout := container.NewBorder(
		container.New(layout.NewVBoxLayout(),
			languageRow,
		),
		nil, nil, nil,
		container.NewVSplit(
			transcriptionRow,
			leftVerticalBottomLayout,
		),
	)

	Fields.Field.WhisperResultList = widget.NewList(
		func() int {
			return len(Fields.DataBindings.WhisperResultsData)
		},
		func() fyne.CanvasObject {
			return container.New(layout.NewGridLayoutWithRows(2),
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
		func(i widget.ListItemID, o fyne.CanvasObject) {
			whisperMessage := Fields.DataBindings.WhisperResultsData[i]

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
			if whisperMessage.TxtTranslation == "" {
				translateResultLabel.SetText(whisperMessage.Text)
				translateResultLanguageLabel.SetText("[" + whisperMessage.Language + "]")

				originalTranscriptionLabel.SetText("")
				originalTranscriptionLanguageLabel.SetText("")
			} else { // bind data to elements if translation was generated
				translateResultLabel.SetText(whisperMessage.TxtTranslation)
				translateResultLanguageLabel.SetText("[" + whisperMessage.TxtTranslationTarget + "]")

				originalTranscriptionLabel.SetText(whisperMessage.Text)
				originalTranscriptionLanguageLabel.SetText("[" + whisperMessage.Language + "]")
			}

			// resize
			Fields.Field.WhisperResultList.SetItemHeight(i, translateResultLabel.MinSize().Height+originalTranscriptionLabel.MinSize().Height+theme.Padding())
		},
	)

	Fields.Field.WhisperResultList.OnSelected = func(id widget.ListItemID) {
		whisperMessage := Fields.DataBindings.WhisperResultsData[id]

		Fields.DataBindings.TranscriptionInputBinding.Set(whisperMessage.Text)
		if whisperMessage.TxtTranslation != "" {
			Fields.Field.TranscriptionTranslationSpeechToTextInput.SetText(whisperMessage.TxtTranslation)
		} else {
			Fields.Field.TranscriptionTranslationSpeechToTextInput.SetText(whisperMessage.Text)
		}

		go func() {
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() {
				Fields.Field.WhisperResultList.Unselect(id)
			})
		}()
	}

	if !Settings.Config.Realtime {
		Fields.Field.RealtimeResultLabel.Hide()
	}
	realtimeWhisperResultBlock := container.NewBorder(
		nil, container.NewVBox(widget.NewSeparator(), widget.NewSeparator()), nil, nil,
		Fields.Field.RealtimeResultLabel,
	)

	clearResultListButton := widget.NewButtonWithIcon(lang.L("Clear"), theme.ContentClearIcon(), func() {
		Fields.DataBindings.WhisperResultsData = []Fields.WhisperResult{}
		Fields.Field.WhisperResultList.Refresh()
		sendMessage := SendMessageChannel.SendMessageStruct{
			Type:  "clear_transcription",
			Value: true,
		}
		sendMessage.SendMessage()
	})

	saveCsvButton := widget.NewButton(lang.L("Save CSV"), func() {
		dialogSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
		dialogSize.Height = dialogSize.Height - 80
		dialogSize.Width = dialogSize.Width - 80

		saveStartingPath := fyne.CurrentApp().Preferences().StringWithFallback("LastCSVTranscriptionSavePath", "")

		fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err == nil && writer != nil {
				filePath := writer.URI().Path()
				sendMessage := SendMessageChannel.SendMessageStruct{
					Type:  "save_transcription",
					Value: filePath,
				}
				sendMessage.SendMessage()
				fyne.CurrentApp().Preferences().SetString("LastCSVTranscriptionSavePath", filepath.Dir(filePath))
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fileDialog.Resize(dialogSize)

		if saveStartingPath != "" {
			// check if folder exists
			folderExists := false
			if _, err := os.Stat(saveStartingPath); !os.IsNotExist(err) {
				folderExists = true
			}
			if folderExists {
				fileURI := storage.NewFileURI(saveStartingPath)
				fileLister, _ := storage.ListerForURI(fileURI)

				fileDialog.SetLocation(fileLister)
			}
		}

		fileDialog.SetFileName("transcription_" + time.Now().Format("2006-01-02_15-04-05") + ".csv")

		fileDialog.Show()
	})
	lastResultLine := container.NewBorder(nil, nil, saveCsvButton, clearResultListButton)

	whisperResultContainer := container.NewStack(
		container.NewBorder(
			realtimeWhisperResultBlock, lastResultLine, nil, nil,
			Fields.Field.WhisperResultList,
		),
	)

	mainContent := container.NewHSplit(
		leftVerticalLayout,
		container.NewStack(
			whisperResultContainer,
		),
	)

	Fields.Field.ProcessingStatus.Stop()

	mainContent.SetOffset(0.6)

	return mainContent
}
