package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"whispering-tiger-ui/Pages/SettingsMappings"
	"whispering-tiger-ui/Utilities"
)

func CreateSettingsWindow() fyne.CanvasObject {
	defer Utilities.PanicLogger()

	settingsFormTabs := container.NewAppTabs(
		container.NewTabItem("Speech-to-Text Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.SpeechToTextSettingsMapping)),
		container.NewTabItem("Text-Translate Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextTranslateSettingsMapping)),
		container.NewTabItem("Text-to-Speech Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextToSpeechSettingsMapping)),
		container.NewTabItem("OSC (VRChat) Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.OSCSettingsMapping)),
		container.NewTabItem("Experimental Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.ExperimentalSettingsMapping)),
	)
	settingsFormTabs.SetTabLocation(container.TabLocationLeading)

	return settingsFormTabs
}
