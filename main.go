package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages"
	"whispering-tiger-ui/Pages/Advanced"
	"whispering-tiger-ui/Resources"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/UpdateUtility"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Websocket"
)

var WebsocketClient = Websocket.NewClient("127.0.0.1:5000")

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
	// main application
	val, ok := os.LookupEnv("WT_SCALE")
	if ok {
		_ = os.Setenv("FYNE_SCALE", val)
	}

	lang.SetLanguageOrder([]string{"en"})
	langVal, langOk := os.LookupEnv("PREFERRED_LANGUAGE")
	if langOk && langVal != "" {
		lang.SetPreferredLocale(langVal)
	}
	lang.AddTranslationsFS(Resources.Translations, "translations")

	a := app.NewWithID("io.github.whispering-tiger")
	a.SetIcon(Resources.ResourceAppIconPng)

	a.Settings().SetTheme(&AppTheme{})

	Utilities.AppVersion = a.Metadata().Version
	Utilities.AppBuild = strconv.Itoa(a.Metadata().Build)

	w := a.NewWindow("Whispering Tiger")
	w.SetMaster()
	w.CenterOnScreen()

	// initialize global fields (so they can use initialized languages)
	Fields.InitializeGlobalFields()

	w.SetOnClosed(func() {
		fyne.CurrentApp().Preferences().SetFloat("MainWindowWidth", float64(w.Canvas().Size().Width))
		fyne.CurrentApp().Preferences().SetFloat("MainWindowHeight", float64(w.Canvas().Size().Height))
	})

	// initialize whisper process
	//var whisperProcess = RuntimeBackend.NewWhisperProcess()
	//whisperProcess.DeviceIndex = Settings.Config.Device_index.(string)
	//whisperProcess.DeviceOutIndex = Settings.Config.Device_out_index.(string)
	//whisperProcess.SettingsFile = Settings.Config.SettingsFilename
	//RuntimeBackend.BackendsList = append(RuntimeBackend.BackendsList, whisperProcess)

	// allow enabling logging via environment variable
	loggingVal, loggingOk := os.LookupEnv("ENABLE_LOGGING")
	if loggingOk {
		if loggingVal != "" {
			enableLogging, err := Utilities.ParseBoolean(loggingVal)
			if err != nil {
				log.Printf("Error parsing ENABLE_LOGGING: %v", err)
				enableLogging = true // default to true if parsing fails
			}
			fyne.CurrentApp().Preferences().SetBool("WriteLogfile", enableLogging)
		}
	}

	profileWindow := a.NewWindow(lang.L("Whispering Tiger Profiles"))

	onProfileClose := func() {

		RuntimeBackend.BackendsList = append(RuntimeBackend.BackendsList, RuntimeBackend.NewWhisperProcess())
		RuntimeBackend.BackendsList[0].DeviceIndex = strconv.Itoa(Settings.Config.Device_index.(int))
		RuntimeBackend.BackendsList[0].DeviceOutIndex = strconv.Itoa(Settings.Config.Device_out_index.(int))
		RuntimeBackend.BackendsList[0].SettingsFile = filepath.Join(Settings.GetConfProfileDir(), Settings.Config.SettingsFilename)
		// Setting this to use UTF-8 encoding for Python does not work when build using PyInstaller
		if fyne.CurrentApp().Preferences().BoolWithFallback("RunWithUTF8", true) {
			log.Printf("Running with UTF-8 encoding")
			RuntimeBackend.BackendsList[0].AttachEnvironment("PYTHONIOENCODING", "UTF-8")
			RuntimeBackend.BackendsList[0].AttachEnvironment("PYTHONLEGACYWINDOWSSTDIO", "UTF-8")
			RuntimeBackend.BackendsList[0].AttachEnvironment("PYTHONUTF8", "1")
		}
		RuntimeBackend.BackendsList[0].AttachEnvironment("CT2_CUDA_ALLOW_FP16", "1")

		// AMD ROCm support (Todo: does this work with NVIDIA?)
		RuntimeBackend.BackendsList[0].AttachEnvironment("HSA_OVERRIDE_GFX_VERSION", "10.3.0")

		// RuntimeBackend.BackendsList[0].AttachEnvironment("CUBLAS_WORKSPACE_CONFIG", ":4096:8")
		if Utilities.FileExists("ffmpeg/bin/ffmpeg.exe") {
			appExec, _ := os.Executable()
			appPath := filepath.Dir(appExec)

			RuntimeBackend.BackendsList[0].AttachEnvironment("Path", filepath.Join(appPath, "ffmpeg/bin"))
		}
		if Settings.Config.Run_backend {
			if !fyne.CurrentApp().Preferences().BoolWithFallback("DisableUiDownloads", false) {
				RuntimeBackend.BackendsList[0].UiDownload = true
			}
			if !Settings.Config.Run_backend_reconnect {
				RuntimeBackend.BackendsList[0].Start()
			}
		}

		// initialize status bar
		Fields.Field.StatusBar = widget.NewProgressBar()
		Fields.Field.StatusBar.TextFormatter = func() string {
			return ""
		}

		//activitySpacer := layout.NewSpacer()
		activitySpacer := &layout.Spacer{
			FixHorizontal: true,
		}
		//activitySpacer.FixHotWidth(10)
		Fields.Field.StatusRow = container.NewStack(Fields.Field.StatusBar, container.NewBorder(nil, nil, nil, container.NewBorder(nil, nil, nil, activitySpacer, Fields.Field.ProcessingStatus), Fields.Field.StatusText))

		// initialize main window
		appTabs := container.NewAppTabs(
			container.NewTabItemWithIcon(lang.L("Speech-to-Text"), theme.NewThemedResource(Resources.ResourceSpeechToTextIconSvg), Pages.CreateSpeechToTextWindow()),
			container.NewTabItemWithIcon(lang.L("Text-Translate"), theme.NewThemedResource(Resources.ResourceTranslateIconSvg), Pages.CreateTextTranslateWindow()),
			container.NewTabItemWithIcon(lang.L("Text-to-Speech"), theme.NewThemedResource(Resources.ResourceTextToSpeechIconSvg), Pages.CreateTextToSpeechWindow()),
			container.NewTabItemWithIcon(lang.L("Image-to-Text"), theme.NewThemedResource(Resources.ResourceImageRecognitionIconSvg), Pages.CreateOcrWindow()),
			container.NewTabItemWithIcon(lang.L("Plugins"), theme.NewThemedResource(Resources.ResourcePluginsIconSvg), Advanced.CreatePluginSettingsPage()),
			container.NewTabItemWithIcon(lang.L("Settings"), theme.SettingsIcon(), Pages.CreateSettingsWindow()),
			container.NewTabItemWithIcon(lang.L("Advanced"), theme.MoreVerticalIcon(), Pages.CreateAdvancedWindow()),
		)
		appTabs.SetTabLocation(container.TabLocationTop)

		appTabs.OnSelected = func(tab *container.TabItem) {
			if tab.Text == lang.L("Text-to-Speech") {
				Pages.OnOpenTextToSpeechWindow(tab.Content)
			} else {
				Pages.OnCloseTextToSpeechWindow(tab.Content)
			}
			if tab.Text == lang.L("Settings") {
				tab.Content = Pages.CreateSettingsWindow()
				tab.Content.Refresh()
			}
			if tab.Text == lang.L("Plugins") {
				tab.Content.(*container.Scroll).Content = Advanced.CreatePluginSettingsPage()
				tab.Content.(*container.Scroll).Content.Refresh()
				tab.Content.(*container.Scroll).Refresh()
			}
			if tab.Text == lang.L("Advanced") {
				// check if tab content is of type container.AppTabs
				if tabContent, ok := tab.Content.(*container.AppTabs); ok {
					if tabContent.Selected().Text == lang.L("Advanced Settings") {
						// force trigger onselect for (Advanced -> Settings) Tab
						tabContent.OnSelected(tabContent.Items[tabContent.SelectedIndex()])
					}
					if tabContent.Selected().Text == lang.L("Logs") {
						Fields.Field.LogText.SetText("")
						Fields.Field.LogText.Write([]byte(strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\r\n") + "\r\n"))
					}
				}
			}
		}

		Fields.Field.StatusText.Wrapping = fyne.TextTruncate
		w.SetContent(container.NewBorder(nil, Fields.Field.StatusRow, nil, nil, appTabs))

		// set main window size
		mainWindowWidth := fyne.CurrentApp().Preferences().FloatWithFallback("MainWindowWidth", 1200)
		mainWindowHeight := fyne.CurrentApp().Preferences().FloatWithFallback("MainWindowHeight", 600)

		w.Resize(fyne.NewSize(float32(mainWindowWidth), float32(mainWindowHeight)))

		// show main window
		w.Show()

		// set websocket client to configured ip+port
		WebsocketClient.Addr = Settings.Config.Websocket_ip + ":" + strconv.Itoa(Settings.Config.Websocket_port)

		go WebsocketClient.Start()

		fyne.CurrentApp().Preferences().SetFloat("ProfileWindowWidth", float64(profileWindow.Canvas().Size().Width))
		fyne.CurrentApp().Preferences().SetFloat("ProfileWindowHeight", float64(profileWindow.Canvas().Size().Height))

		// close profile window
		profileWindow.Close()
	}

	profilePage := Pages.CreateProfileWindow(onProfileClose)
	profileWindow.SetContent(profilePage)

	// set profile window size
	profileWindowWidth := fyne.CurrentApp().Preferences().FloatWithFallback("ProfileWindowWidth", 1400)
	profileWindowHeight := fyne.CurrentApp().Preferences().FloatWithFallback("ProfileWindowHeight", 600)
	profileWindow.Resize(fyne.NewSize(float32(profileWindowWidth), float32(profileWindowHeight)))

	profileWindow.CenterOnScreen()
	profileWindow.Show()

	if !fyne.CurrentApp().Preferences().BoolWithFallback("SendErrorsToServerInit", false) {
		confirmErrorReportingDialog := dialog.NewConfirm(lang.L("Automatically Report Errors"), lang.L("Do you want to automatically report errors?"),
			func(b bool) {
				Logging.EnableReporting(b)
				fyne.CurrentApp().Preferences().SetBool("SendErrorsToServerInit", true) // Set Init to mark that it was already shown
			},
			profileWindow)
		confirmErrorReportingDialog.Show()
	}
	SendErrorsToServerSetting := fyne.CurrentApp().Preferences().BoolWithFallback("SendErrorsToServer", false)
	Logging.EnableReporting(SendErrorsToServerSetting)

	Logging.ErrorHandlerInit(Utilities.AppVersion+"."+Utilities.AppBuild, UpdateUtility.GetCurrentPlatformVersion())
	defer Logging.ErrorHandlerRecover()

	// check for updates
	if fyne.CurrentApp().Preferences().BoolWithFallback("CheckForUpdateAtStartup", true) || (!Utilities.FileExists("audioWhisper/audioWhisper.exe") && !Utilities.FileExists("audioWhisper.py")) {
		go func() {
			if len(fyne.CurrentApp().Driver().AllWindows()) == 2 {
				UpdateUtility.VersionCheck(fyne.CurrentApp().Driver().AllWindows()[1], false)
			}
		}()
	}
	if fyne.CurrentApp().Preferences().BoolWithFallback("CheckForPluginUpdatesAtStartup", true) {
		go func() {
			lastCheckTimestamp := fyne.CurrentApp().Preferences().IntWithFallback("CheckForPluginUpdatesAtStartupLastTime", 0)
			lastCheckTime := time.Unix(int64(lastCheckTimestamp), 0)
			currentTime := time.Now()

			// make sure to check for plugin updates only every day
			if lastCheckTime.Year() != currentTime.Year() || lastCheckTime.YearDay() != currentTime.YearDay() {
				fyne.CurrentApp().Preferences().SetInt("CheckForPluginUpdatesAtStartupLastTime", int(currentTime.Unix()))

				if UpdateUtility.PluginsUpdateAvailable() {
					dialog.ShowConfirm(lang.L("New Plugin updates available"), lang.L("Whispering Tiger has new Plugin updates available. Go to Plugin List now?"), func(b bool) {
						if b {
							Advanced.CreatePluginListWindow(nil, false)
						}
					}, fyne.CurrentApp().Driver().AllWindows()[1])
				}
			}
		}()
	}

	a.Lifecycle().SetOnStopped(func() {
		// after run (app exit), send whisper process signal to stop
		if len(RuntimeBackend.BackendsList) > 0 {
			RuntimeBackend.BackendsList[0].Stop()
			RuntimeBackend.BackendsList[0].WriterBackend.Close()
			RuntimeBackend.BackendsList[0].ReaderBackend.Close()
		}
	})

	a.Run()
}
