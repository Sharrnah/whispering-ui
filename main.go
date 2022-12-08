package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"os"
	"path/filepath"
	"runtime"
	"whispering-tiger-ui/Pages"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/websocket"
)

func overwriteFyneFont() {
	pwd, _ := filepath.Abs("./")
	if _, err := os.Stat(pwd + "\\" + "GoNoto.ttf"); err == nil {
		if err := os.Setenv("FYNE_FONT", pwd+"\\"+"GoNoto.ttf"); err != nil {
			fmt.Printf("WARNING: failed to set FYNE_FONT=%s: %v\n", pwd+"\\"+"GoNoto.ttf", err)
		}
		return
	}

	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" {
		winDir := os.Getenv("WINDIR")
		if len(winDir) == 0 {
			return
		}
		fontPath := determineWindowsFont(winDir + "\\Fonts")
		if err := os.Setenv("FYNE_FONT", fontPath); err != nil {
			fmt.Printf("WARNING: failed to set FYNE_FONT=%s: %v\n", fontPath, err)
		}
	}
}

func determineWindowsFont(fontsDir string) string {
	font := "YuGothM.ttc"
	if _, err := os.Stat(fontsDir + "\\" + font); err == nil {
		return fontsDir + "\\" + font
	}
	font = "meiryo.ttc"
	if _, err := os.Stat(fontsDir + "\\" + font); err == nil {
		return fontsDir + "\\" + font
	}
	font = "msgothic.ttc"
	if _, err := os.Stat(fontsDir + "\\" + font); err == nil {
		return fontsDir + "\\" + font
	}
	font = "segoeui.ttf"
	if _, err := os.Stat(fontsDir + "\\" + font); err == nil {
		return fontsDir + "\\" + font
	}
	return ""
}

func main() {
	overwriteFyneFont()
	a := app.NewWithID("tiger.whispering")
	a.SetIcon(resourceAppIconPng)
	w := a.NewWindow("Whispering Tiger")

	w.SetMaster()

	var WhisperProcess = RuntimeBackend.WhisperProcessConfig{
		DeviceIndex:  "5",
		SettingsFile: "settings.yaml",
	}
	WhisperProcess.StartWhisper()

	//Pages.AppTabs.SetTabLocation(container.TabLocationTop)

	appTabs := container.NewAppTabs(
		container.NewTabItem("Speech 2 Text", Pages.CreateSpeechToTextWindow()),
		container.NewTabItem("Text Translate", Pages.CreateTextTranslateWindow()),
		container.NewTabItem("Text 2 Speech", Pages.CreateTextToSpeechWindow()),
		container.NewTabItem("OCR", Pages.CreateOcrWindow()),
		container.NewTabItem("Advanced", Pages.CreateAdvancedWindow()),
	)
	appTabs.SetTabLocation(container.TabLocationTop)

	w.SetContent(appTabs)

	w.Resize(fyne.NewSize(1200, 600))

	go websocket.Start()

	w.ShowAndRun()
}
