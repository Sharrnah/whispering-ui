package Pages

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func newLiveDisplaySelect(name, value string, inherit bool, changed func(string)) *CustomWidget.TextValueSelect {
	options := []CustomWidget.TextValueOption{
		{Text: lang.L("Replace old lines"), Value: "blocks"},
		{Text: lang.L("Rolling recent words"), Value: "rolling"},
	}
	if inherit {
		options = append([]CustomWidget.TextValueOption{{Text: lang.L("Use main setting"), Value: ""}}, options...)
	}
	control := CustomWidget.NewTextValueSelect(name, options, nil, 0)
	control.SetSelected(value)
	control.OnChanged = func(option CustomWidget.TextValueOption) { changed(option.Value) }
	return control
}

func newStreamingChatLimit(value *int, inherit bool, changed func(*int)) *widget.Entry {
	limit := widget.NewEntry()
	if value != nil {
		limit.SetText(strconv.Itoa(*value))
	}
	if inherit {
		limit.SetPlaceHolder(lang.L("Use main setting") + ": " + strconv.Itoa(Settings.Config.Osc_chat_limit))
	}
	limit.Validator = func(value string) error {
		if inherit && value == "" {
			return nil
		}
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 4096 {
			return fmt.Errorf("%s", lang.L("Enter a whole number from 1 to 4096."))
		}
		return nil
	}
	limit.OnChanged = func(value string) {
		if limit.Validate() != nil {
			return
		}
		if value == "" {
			changed(nil)
			return
		}
		n, _ := strconv.Atoi(value)
		changed(&n)
	}
	return limit
}

func buildVibeVoiceStreamingSettings() fyne.CanvasObject {
	mode := newLiveDisplaySelect("streaming_display", Settings.MainStreamingDisplayMode(), false, func(value string) {
		Settings.Config.Streaming_display_mode = value
		SendMessageChannel.SendMessageStruct{Type: "setting_change", Name: "streaming_display_mode", Value: value}.SendMessage()
	})
	limit := newStreamingChatLimit(&Settings.Config.Osc_chat_limit, false, func(value *int) {
		Settings.Config.Osc_chat_limit = *value
		SendMessageChannel.SendMessageStruct{Type: "setting_change", Name: "osc_chat_limit", Value: *value}.SendMessage()
	})
	return widget.NewForm(
		widget.NewFormItem(lang.L("Live text display (main microphone)"), mode),
		widget.NewFormItem(lang.L("Maximum chatbox length"), limit))
}
