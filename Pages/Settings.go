package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"whispering-tiger-ui/Pages/SettingsMappings"
	"whispering-tiger-ui/Utilities"
)

func CreateSettingsWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

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
