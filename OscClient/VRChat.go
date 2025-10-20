package OscClient

import (
	"fyne.io/fyne/v2"
	"whispering-tiger-ui/Settings"
)

func SendChatboxTyping(value bool) {
	// Check if OSC is enabled and typing indicator is enabled
	SendTypingIndicatorOscSetting := fyne.CurrentApp().Preferences().BoolWithFallback("SendTypingIndicatorOsc", false)
	if Settings.Config.Osc_ip != "" && Settings.Config.Osc_port != 0 &&
		SendTypingIndicatorOscSetting &&
		Settings.Config.Osc_typing_indicator &&
		(Settings.Config.Osc_auto_processing_enabled || Settings.Config.Osc_force_activity_indication) {
		go func(valueSend bool) {
			SendBool("/chatbox/typing", valueSend)
		}(value)
	}
}
