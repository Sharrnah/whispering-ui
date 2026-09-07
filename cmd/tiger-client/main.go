package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/lang"
	"os"
	"whispering-tiger-ui/RemoteAudioView"
	"whispering-tiger-ui/Resources"
)

func main() {
	lang.SetLanguageOrder([]string{"en"})
	if locale := os.Getenv("PREFERRED_LANGUAGE"); locale != "" {
		lang.SetPreferredLocale(locale)
	}
	_ = lang.AddTranslationsFS(Resources.Translations, "translations")
	a := app.NewWithID("com.whispering-tiger.remote-client")
	w := a.NewWindow(lang.L("Remote audio client"))
	content, closeClient := RemoteAudioView.Client(w)
	w.SetContent(content)
	w.SetOnClosed(closeClient)
	w.Resize(fyne.NewSize(640, 720))
	w.ShowAndRun()
}
