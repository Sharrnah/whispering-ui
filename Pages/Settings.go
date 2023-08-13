package Pages

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"whispering-tiger-ui/Pages/SettingsMappings"
)

func CreateSettingsWindow() fyne.CanvasObject {

	settingsFormTabs := container.NewAppTabs(
		container.NewTabItem("Speech-to-Text Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.SpeechToTextSettingsMapping)),
		container.NewTabItem("Text-Translate Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextTranslateSettingsMapping)),
		container.NewTabItem("Text-to-Speech Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.TextToSpeechSettingsMapping)),
		container.NewTabItem("OSC (VRChat) Options", SettingsMappings.CreateSettingsFormByMapping(SettingsMappings.OSCSettingsMapping)),
	)
	settingsFormTabs.SetTabLocation(container.TabLocationLeading)

	return settingsFormTabs
}
