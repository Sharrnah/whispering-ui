package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/Pages"
	"whispering-tiger-ui/websocket"
)

func main() {
	a := app.NewWithID("tiger.whispering")
	a.SetIcon(resourceAppIconPng)
	w := a.NewWindow("Whispering Tiger")
	w.SetMaster()

	tabs := container.NewAppTabs(
		container.NewTabItem("Main", Pages.CreateMainWindow()),
		container.NewTabItem("Speech 2 Text", widget.NewLabel("Hello")),
		container.NewTabItem("Text Translate", widget.NewLabel("Hello")),
		container.NewTabItem("Text 2 Speech", widget.NewLabel("Hello")),
		container.NewTabItem("OCR", widget.NewLabel("Hello")),
		container.NewTabItem("Settings", widget.NewLabel("World!")),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	//w.SetContent(widget.NewLabel("Hello World!"))
	w.SetContent(tabs)

	w.Resize(fyne.NewSize(1200, 600))

	go websocket.Start()

	w.ShowAndRun()
}
