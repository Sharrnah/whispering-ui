package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"strconv"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/SettingsMappings"
	"whispering-tiger-ui/RemoteAudioView"
	"whispering-tiger-ui/Settings"
)

func CreateSettingsWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Settings->CreateSettingsWindow")
	})

	applicationSettings := container.NewBorder(
		container.NewVBox(createLiveAudioDeviceSettings(), widget.NewSeparator()),
		nil,
		nil,
		nil,
		SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.ApplicationSettingsMapping),
	)

	settingsFormTabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Application Options"), applicationSettings),
		container.NewTabItem(lang.L("Speech-to-Text Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.SpeechToTextSettingsMapping)),
		container.NewTabItem(lang.L("Text-Translate Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextTranslateSettingsMapping)),
		container.NewTabItem(lang.L("Text-to-Speech Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextToSpeechSettingsMapping)),
		container.NewTabItem(lang.L("OSC (VRChat) Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.OSCSettingsMapping)),
		container.NewTabItem(lang.L("Experimental Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.ExperimentalSettingsMapping)),
	)
	settingsFormTabs.SetTabLocation(container.TabLocationLeading)
	if Settings.Config.Run_backend {
		windows := fyne.CurrentApp().Driver().AllWindows()
		if len(windows) > 0 {
			settingsFormTabs.Append(container.NewTabItem(lang.L("Remote audio"), RemoteAudioView.Host(windows[0], "127.0.0.1:"+strconv.Itoa(Settings.Config.Websocket_port))))
		}
	} else {
		settingsFormTabs.Append(container.NewTabItem(lang.L("Remote audio"), RemoteAudioView.Integrated()))
	}

	return settingsFormTabs
}
