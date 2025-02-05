package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"github.com/getsentry/sentry-go"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Pages/SettingsMappings"
)

func CreateSettingsWindow() fyne.CanvasObject {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Settings->CreateSettingsWindow")
	})

	settingsFormTabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Application Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.ApplicationSettingsMapping)),
		container.NewTabItem(lang.L("Speech-to-Text Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.SpeechToTextSettingsMapping)),
		container.NewTabItem(lang.L("Text-Translate Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextTranslateSettingsMapping)),
		container.NewTabItem(lang.L("Text-to-Speech Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextToSpeechSettingsMapping)),
		container.NewTabItem(lang.L("OSC (VRChat) Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.OSCSettingsMapping)),
		container.NewTabItem(lang.L("Experimental Options"), SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.ExperimentalSettingsMapping)),
	)
	settingsFormTabs.SetTabLocation(container.TabLocationLeading)

	return settingsFormTabs
}
