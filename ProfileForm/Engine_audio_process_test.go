package ProfileForm

import (
	"testing"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

func newAudioInputEngine(applicationOptions []CustomWidget.TextValueOption) (*FormEngine, *CustomWidget.TextValueSelect, *CustomWidget.TextValueSelect) {
	controls := &AllProfileControls{}
	engine := NewFormEngine(controls, nil)
	inputSelection := CustomWidget.NewTextValueSelect("device_index", []CustomWidget.TextValueOption{
		{Text: "Default", Value: "-1"},
		{Text: "Application Audio", Value: Utilities.AudioApplicationOptionValue},
	}, nil, 0)
	applicationSelection := CustomWidget.NewTextValueSelect("audio_input_application", applicationOptions, nil, 0)
	controls.AudioInput = inputSelection
	controls.AudioApplication = applicationSelection
	engine.Register("device_index", inputSelection)
	return engine, inputSelection, applicationSelection
}

func TestSaveApplicationAudioInputKeepsDeviceIndexNumeric(t *testing.T) {
	processValue := Utilities.FormatAudioProcessOptionValue(4242, "player.exe")
	engine, inputSelection, applicationSelection := newAudioInputEngine([]CustomWidget.TextValueOption{
		{Text: "player.exe (PID 4242)", Value: processValue},
	})
	inputSelection.SetSelected(Utilities.AudioApplicationOptionValue)
	applicationSelection.SetSelected(processValue)
	conf := Settings.Conf{}

	engine.SaveToSettings(&conf)

	if conf.Device_index != -1 {
		t.Fatalf("device_index = %#v, want numeric -1", conf.Device_index)
	}
	if conf.Audio_input_process != "player.exe" {
		t.Fatalf("audio_input_process = %q", conf.Audio_input_process)
	}
	if conf.Audio_input_process_id != 4242 {
		t.Fatalf("audio_input_process_id = %d", conf.Audio_input_process_id)
	}
}

func TestLoadingApplicationAudioInputRecoversAfterPIDChange(t *testing.T) {
	currentValue := Utilities.FormatAudioProcessOptionValue(9000, "player.exe")
	engine, inputSelection, applicationSelection := newAudioInputEngine([]CustomWidget.TextValueOption{
		{Text: "player.exe (PID 9000)", Value: currentValue},
	})
	conf := Settings.Conf{
		Device_index:           -1,
		Audio_input_device:     "player.exe (PID 1234)",
		Audio_input_process:    "player.exe",
		Audio_input_process_id: 1234,
	}

	engine.LoadFromSettings(&conf)

	selectedInput := inputSelection.GetSelected()
	if selectedInput == nil || selectedInput.Value != Utilities.AudioApplicationOptionValue {
		t.Fatalf("input selection = %#v, want Application Audio", selectedInput)
	}
	selectedApplication := applicationSelection.GetSelected()
	if selectedApplication == nil || selectedApplication.Value != currentValue {
		t.Fatalf("application selection = %#v, want current process option %q", selectedApplication, currentValue)
	}
	engine.SaveToSettings(&conf)
	if conf.Audio_input_process_id != 9000 {
		t.Fatalf("saved PID = %d, want refreshed PID 9000", conf.Audio_input_process_id)
	}
}

func TestLoadingApplicationAudioInputDoesNotGuessBetweenInstances(t *testing.T) {
	firstValue := Utilities.FormatAudioProcessOptionValue(9000, "player.exe")
	secondValue := Utilities.FormatAudioProcessOptionValue(9001, "player.exe")
	engine, inputSelection, applicationSelection := newAudioInputEngine([]CustomWidget.TextValueOption{
		{Text: "player.exe (PID 9000)", Value: firstValue},
		{Text: "player.exe (PID 9001)", Value: secondValue},
	})
	conf := Settings.Conf{
		Device_index:           -1,
		Audio_input_device:     "player.exe (PID 1234)",
		Audio_input_process:    "player.exe",
		Audio_input_process_id: 1234,
	}

	engine.LoadFromSettings(&conf)

	selectedInput := inputSelection.GetSelected()
	if selectedInput == nil || selectedInput.Value != Utilities.AudioApplicationOptionValue {
		t.Fatalf("input selection = %#v, want Application Audio", selectedInput)
	}
	want := Utilities.FormatAudioProcessOptionValue(1234, "player.exe")
	selectedApplication := applicationSelection.GetSelected()
	if selectedApplication == nil || selectedApplication.Value != want {
		t.Fatalf("application selection = %#v, want unresolved saved option %q", selectedApplication, want)
	}
}

func TestSavingDeviceAudioInputClearsOldApplicationTarget(t *testing.T) {
	engine, inputSelection, _ := newAudioInputEngine(nil)
	inputSelection.SetValueOptions([]CustomWidget.TextValueOption{
		{Text: "Default", Value: "-1"},
		{Text: "Microphone", Value: "7"},
		{Text: "Application Audio", Value: Utilities.AudioApplicationOptionValue},
	})
	inputSelection.SetSelected("7")
	conf := Settings.Conf{
		Audio_input_process:    "old.exe",
		Audio_input_process_id: 1234,
	}

	engine.SaveToSettings(&conf)

	if conf.Device_index != 7 {
		t.Fatalf("device_index = %#v, want 7", conf.Device_index)
	}
	if conf.Audio_input_process != "" || conf.Audio_input_process_id != 0 {
		t.Fatalf("old application target was retained: %#v", conf)
	}
}
