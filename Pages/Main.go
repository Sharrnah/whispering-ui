package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"whispering-tiger-ui/websocket/Messages"
)

func CreateMainWindow() fyne.CanvasObject {
	title := container.New(layout.NewCenterLayout(), widget.NewLabel("Main"))

	//widget.NewSeparator()
	//layout.NewSpacer()

	// language row
	languageCombo := widget.NewSelect([]string{"Option 1", "Option 2"}, func(value string) {
		log.Println("Select set to", value)
	})

	lastDetectedLanguage := widget.NewLabel("Last detected Language: ")

	LanguageRow := container.New(layout.NewHBoxLayout(), widget.NewLabel("Current Language: "), languageCombo, layout.NewSpacer(), lastDetectedLanguage)

	// transcription row
	transcriptionInput := widget.NewMultiLineEntry()
	transcriptionInput.Wrapping = fyne.TextWrapWord

	transcriptionTranslation := widget.NewMultiLineEntry()
	transcriptionTranslation.Wrapping = fyne.TextWrapWord

	transcriptionRow := container.New(layout.NewGridLayout(2), transcriptionInput, transcriptionTranslation)

	// quick options row
	TTSCheck := widget.NewCheck("Text 2 Speech", func(value bool) {
		log.Println("Check set to", value)
	})
	OSCCheck := widget.NewCheck("OSC", func(value bool) {
		log.Println("Check set to", value)
	})
	quickOptionsRow := container.New(
		layout.NewVBoxLayout(),
		TTSCheck,
		OSCCheck,
	)

	// whisper results row
	//Messages.WhisperResultsStringList = []string{}
	resultList := widget.NewListWithData(Messages.WhisperResultsDataBinding,
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		})

	// split between transcription + options row
	splitTranscriptionOptions := container.NewVSplit(transcriptionRow, quickOptionsRow)
	splitTranscriptionOptions.Offset = 0.2

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
