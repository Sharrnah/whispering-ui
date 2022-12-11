package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"whispering-tiger-ui/Pages"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/websocket"
)

var WebsocketClient = websocket.NewClient("127.0.0.1:5000")

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
	w.CenterOnScreen()

	// initialize whisper process
	//var whisperProcess = RuntimeBackend.NewWhisperProcess()
	//whisperProcess.DeviceIndex = Settings.Config.Device_index.(string)
	//whisperProcess.DeviceOutIndex = Settings.Config.Device_out_index.(string)
	//whisperProcess.SettingsFile = Settings.Config.SettingsFilename
	//RuntimeBackend.BackendsList = append(RuntimeBackend.BackendsList, whisperProcess)

	profileWindow := a.NewWindow("Whispering Tiger Profiles")

	onProfileClose := func() {

		RuntimeBackend.BackendsList = append(RuntimeBackend.BackendsList, RuntimeBackend.NewWhisperProcess())
		RuntimeBackend.BackendsList[0].DeviceIndex = strconv.Itoa(Settings.Config.Device_index.(int))
		RuntimeBackend.BackendsList[0].DeviceOutIndex = strconv.Itoa(Settings.Config.Device_out_index.(int))
		RuntimeBackend.BackendsList[0].SettingsFile = Settings.Config.SettingsFilename
		RuntimeBackend.BackendsList[0].Start()

		// initialize main window
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

		// set websocket client to configured ip+port
		WebsocketClient.Addr = Settings.Config.Websocket_ip + ":" + strconv.Itoa(Settings.Config.Websocket_port)
		go WebsocketClient.Start()

		// show main window
		w.Show()

		// close profile window
		profileWindow.Close()
	}

	profilePage := Pages.CreateProfileWindow(onProfileClose)
	profileWindow.SetContent(profilePage)
	profileWindow.Resize(fyne.NewSize(1200, 600))

	profileWindow.CenterOnScreen()
	profileWindow.Show()

	a.Run()

	// after run (app exit), send whisper process signal to stop
	if len(RuntimeBackend.BackendsList) > 0 {
		RuntimeBackend.BackendsList[0].Stop()
		RuntimeBackend.BackendsList[0].WriterBackend.Close()
		RuntimeBackend.BackendsList[0].ReaderBackend.Close()
	}
}
