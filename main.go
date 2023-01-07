package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"whispering-tiger-ui/Pages"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Updater"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket"
)

var WebsocketClient = Websocket.NewClient("127.0.0.1:5000")

var updateInfoUrl = "https://s3.libs.space:9000/projects/whispering/latest.yaml"

func versionDownload(updater Updater.UpdatePackages, packageName, filename string, inPlaceUpdate bool) error {
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)
	dialog.ShowCustom("Update in progress...", "Close", statusBarContainer, fyne.CurrentApp().Driver().AllWindows()[1])
	downloadingLabel := widget.NewLabel("Downloading... ")
	randomUrlIndex := rand.Int() % len(updater.Packages["app"].Urls)
	downloader := Updater.Download{
		Url:      updater.Packages[packageName].Urls[randomUrlIndex],
		Filepath: filename,
	}
	downloader.WriteCounter.OnProgress = func(progress, total uint64) {
		if int64(total) == -1 {
			statusBarContainer.Remove(statusBar)
			statusBarContainer.Add(widget.NewProgressBarInfinite())
			statusBarContainer.Refresh()
		} else {
			statusBar.Max = float64(total)
			statusBar.SetValue(float64(progress))

			downloadingLabel.SetText("Downloading... " + humanize.Bytes(total))
		}
	}

	statusBarContainer.Add(downloadingLabel)
	statusBarContainer.Refresh()
	err := downloader.DownloadFile()
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		return err
	}
	appExec, _ := os.Executable()
	if inPlaceUpdate {
		err = os.Rename(appExec, appExec+".old")
		if err != nil {
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		}
	}

	statusBarContainer.Add(widget.NewLabel("Extracting..."))
	statusBarContainer.Refresh()
	err = Updater.Unzip(filename, filepath.Dir(appExec))
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
	}
	err = os.Remove(filename)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
	}

	if err == nil {
		statusBarContainer.Add(widget.NewLabel("Finished."))

		if inPlaceUpdate {
			dialog.ShowConfirm("Update finished", "Restart the Application now?", func(b bool) {
				cmd := exec.Command(appExec)
				cmd.Start()

				os.Exit(0)
			}, fyne.CurrentApp().Driver().AllWindows()[1])
		}
	}

	statusBarContainer.Refresh()

	return nil
}

func versionCheck() {
	updater := Updater.UpdatePackages{}
	err := updater.GetUpdateInfo(updateInfoUrl)
	if err != nil {
		return
	}

	// check platform version
	platformFileWithoutVersion := !Utilities.FileExists(".current_platform.yaml") && (Utilities.FileExists("audioWhisper/audioWhisper.exe") || Utilities.FileExists("audioWhisper.py"))
	platformRequiresUpdate := false
	if Utilities.FileExists(".current_platform.yaml") {
		currentPlatformVersion := Updater.UpdateInfo{}
		data, err := os.ReadFile(".current_platform.yaml")
		if err == nil {
			_ = currentPlatformVersion.ReadYaml(data)
			if currentPlatformVersion.Version != updater.Packages["ai_platform"].Version {
				platformRequiresUpdate = true
			}
		}
	}
	if !Utilities.FileExists("audioWhisper/audioWhisper.exe") && !Utilities.FileExists("audioWhisper.py") {
		platformRequiresUpdate = true
	}

	if platformRequiresUpdate || platformFileWithoutVersion {
		dialog.ShowConfirm("Platform Update available", "There is a new Update of the Platform available. Update to "+updater.Packages["ai_platform"].Version+" now?", func(b bool) {
			if b {
				go func() {
					err = versionDownload(updater, "ai_platform", "audioWhisper_platform.zip", false)
					if err == nil {
						packageInfo := updater.Packages["ai_platform"]
						packageInfo.WriteYaml(".current_platform.yaml")
					}
				}()
			} else {
				if platformFileWithoutVersion {
					packageInfo := updater.Packages["ai_platform"]
					packageInfo.WriteYaml(".current_platform.yaml")
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[1])
	}

	// check app version
	currentAppVersion := fyne.CurrentApp().Metadata().Version + "." + strconv.Itoa(fyne.CurrentApp().Metadata().Build)
	if updater.Packages["app"].Version != currentAppVersion {
		dialog.ShowConfirm("App Update available", "There is a new Update of the App available. Update to "+updater.Packages["app"].Version+" now?", func(b bool) {
			if b {
				go func() {
					err = versionDownload(updater, "app", "whispering-tiger-ui.zip", true)
					if err == nil {
						packageInfo := updater.Packages["app"]
						packageInfo.WriteYaml(".current_app.yaml")
					}
				}()
			}
		}, fyne.CurrentApp().Driver().AllWindows()[1])
	}

}

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
	a := app.NewWithID("tiger.whispering")
	a.SetIcon(Resources.ResourceAppIconPng)

	a.Settings().SetTheme(&AppTheme{})

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
		if Utilities.FileExists("ffmpeg/bin/ffmpeg.exe") {
			appExec, _ := os.Executable()
			appPath := filepath.Dir(appExec)

			RuntimeBackend.BackendsList[0].AttachEnvironment("Path", filepath.Join(appPath, "ffmpeg/bin"))
		}
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
	profileWindow.Resize(fyne.NewSize(1400, 600))

	profileWindow.CenterOnScreen()
	profileWindow.Show()

	// delete old app version
	appExec, _ := os.Executable()
	if _, err := os.Stat(appExec + ".old"); err == nil {
		err = os.Remove(appExec + ".old")
		if err != nil {
			dialog.NewError(err, fyne.CurrentApp().Driver().AllWindows()[1])
		}
	}

	// check for updates
	versionCheck()

	a.Run()

	// after run (app exit), send whisper process signal to stop
	if len(RuntimeBackend.BackendsList) > 0 {
		RuntimeBackend.BackendsList[0].Stop()
		RuntimeBackend.BackendsList[0].WriterBackend.Close()
		RuntimeBackend.BackendsList[0].ReaderBackend.Close()
	}
}
