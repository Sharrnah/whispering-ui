package Pages

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gen2brain/malgo"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
	"whispering-tiger-ui/Utilities/AudioAPI"
)

type liveAudioInputState struct {
	input       CustomWidget.TextValueOption
	application *CustomWidget.TextValueOption
}

type liveAudioInputSwitchRequest struct {
	RequestID           string `json:"request_id"`
	AudioAPI            string `json:"audio_api"`
	AudioInputDevice    string `json:"audio_input_device"`
	AudioInputProcess   string `json:"audio_input_process"`
	AudioInputProcessID int    `json:"audio_input_process_id"`
}

type liveAudioOutputSwitchRequest struct {
	RequestID         string `json:"request_id"`
	AudioAPI          string `json:"audio_api"`
	AudioOutputDevice string `json:"audio_output_device"`
}

func buildLiveAudioInputSwitchRequest(requestID, audioAPI string, input, application *CustomWidget.TextValueOption) (liveAudioInputSwitchRequest, bool) {
	if input == nil {
		return liveAudioInputSwitchRequest{}, false
	}
	request := liveAudioInputSwitchRequest{
		RequestID:        requestID,
		AudioAPI:         audioAPI,
		AudioInputDevice: input.Text,
	}
	if !Utilities.IsAudioApplicationOptionValue(input.Value) {
		return request, true
	}
	if application == nil {
		return liveAudioInputSwitchRequest{}, false
	}
	processID, executable, ok := Utilities.ParseAudioProcessOptionValue(application.Value)
	if !ok {
		return liveAudioInputSwitchRequest{}, false
	}
	request.AudioInputDevice = application.Text
	request.AudioInputProcess = executable
	request.AudioInputProcessID = int(processID)
	return request, true
}

func buildLiveAudioOutputSwitchRequest(requestID, audioAPI string, output *CustomWidget.TextValueOption) (liveAudioOutputSwitchRequest, bool) {
	if output == nil {
		return liveAudioOutputSwitchRequest{}, false
	}
	return liveAudioOutputSwitchRequest{
		RequestID:         requestID,
		AudioAPI:          audioAPI,
		AudioOutputDevice: output.Text,
	}, true
}

func liveAudioInputOptions(backend malgo.Backend) []CustomWidget.TextValueOption {
	options, _, _ := GetAudioDevices(
		backend,
		[]malgo.DeviceType{malgo.Capture, malgo.Loopback},
		0,
		"",
		"",
	)
	if len(options) == 0 {
		options = []CustomWidget.TextValueOption{{Text: "Default", Value: "-1"}}
	}
	return appendApplicationAudioInputOption(options, backend)
}

func liveAudioOutputOptions(backend malgo.Backend) []CustomWidget.TextValueOption {
	options, _, _ := GetAudioDevices(
		backend,
		[]malgo.DeviceType{malgo.Playback},
		0,
		"",
		"",
	)
	if len(options) == 0 {
		return []CustomWidget.TextValueOption{{Text: "Default", Value: "-1"}}
	}
	return options
}

func selectConfiguredApplication(selection *CustomWidget.TextValueSelect) {
	if selection == nil || strings.TrimSpace(Settings.Config.Audio_input_process) == "" {
		return
	}
	executable := strings.TrimSpace(Settings.Config.Audio_input_process)
	processID := uint32(0)
	if Settings.Config.Audio_input_process_id > 0 {
		processID = uint32(Settings.Config.Audio_input_process_id)
	}

	exactValue := Utilities.FormatAudioProcessOptionValue(processID, executable)
	matchingValues := make([]string, 0, 1)
	for _, option := range selection.Options {
		pid, optionExecutable, ok := Utilities.ParseAudioProcessOptionValue(option.Value)
		if !ok || !strings.EqualFold(optionExecutable, executable) {
			continue
		}
		matchingValues = append(matchingValues, option.Value)
		if pid == processID {
			selection.SetSelected(option.Value)
			return
		}
	}
	if len(matchingValues) == 1 {
		selection.SetSelected(matchingValues[0])
		return
	}

	label := strings.TrimSpace(Settings.Config.Audio_input_device)
	if label == "" {
		label = fmt.Sprintf("%s (PID %d)", executable, processID)
	}
	options := append([]CustomWidget.TextValueOption(nil), selection.Options...)
	options = append(options, CustomWidget.TextValueOption{Text: label, Value: exactValue})
	selection.SetValueOptions(options)
	selection.SetSelected(exactValue)
}

func selectConfiguredLiveAudioInput(selection *CustomWidget.TextValueSelect) {
	if selection == nil {
		return
	}
	if strings.TrimSpace(Settings.Config.Audio_input_process) != "" {
		selection.SetSelected(Utilities.AudioApplicationOptionValue)
		return
	}
	deviceName := strings.TrimSpace(Settings.Config.Audio_input_device)
	if deviceName != "" {
		for _, option := range selection.Options {
			if strings.EqualFold(strings.TrimSpace(option.Text), deviceName) {
				selection.SetSelected(option.Value)
				return
			}
		}
	}
	if Settings.Config.Device_index != nil {
		selection.SetSelected(fmt.Sprint(Settings.Config.Device_index))
		if selection.GetSelected() != nil {
			return
		}
	}
	selection.SetSelected("-1")
}

func selectConfiguredLiveAudioOutput(selection *CustomWidget.TextValueSelect) {
	if selection == nil {
		return
	}
	deviceName := strings.TrimSpace(Settings.Config.Audio_output_device)
	if deviceName != "" {
		for _, option := range selection.Options {
			if strings.EqualFold(strings.TrimSpace(option.Text), deviceName) {
				selection.SetSelected(option.Value)
				return
			}
		}
	}
	if Settings.Config.Device_out_index != nil {
		selection.SetSelected(fmt.Sprint(Settings.Config.Device_out_index))
		if selection.GetSelected() != nil {
			return
		}
	}
	selection.SetSelected("-1")
}

func captureLiveAudioInputState(input, application *CustomWidget.TextValueSelect) liveAudioInputState {
	state := liveAudioInputState{}
	if selected := input.GetSelected(); selected != nil {
		state.input = *selected
	}
	if selected := application.GetSelected(); selected != nil {
		selectedCopy := *selected
		state.application = &selectedCopy
	}
	return state
}

func ensureLiveAudioOption(selection *CustomWidget.TextValueSelect, option CustomWidget.TextValueOption) {
	if selection.ContainsEntry(&option, CustomWidget.CompareValue) {
		return
	}
	options := append([]CustomWidget.TextValueOption(nil), selection.Options...)
	selection.SetValueOptions(append(options, option))
}

func createLiveAudioInputControl() fyne.CanvasObject {
	backend := AudioAPI.GetAudioBackendByName(Settings.Config.Audio_api)
	inputSelection := CustomWidget.NewTextValueSelect(
		"live_audio_input",
		liveAudioInputOptions(backend.Backend),
		nil,
		-1,
	)
	applicationSelection := CustomWidget.NewTextValueSelect(
		"live_audio_application",
		applicationCaptureOptions(),
		nil,
		-1,
	)
	applicationSelection.Hide()

	selectConfiguredLiveAudioInput(inputSelection)
	selectConfiguredApplication(applicationSelection)
	if selected := inputSelection.GetSelected(); selected != nil && Utilities.IsAudioApplicationOptionValue(selected.Value) {
		applicationSelection.Show()
	}

	progress := widget.NewProgressBarInfinite()
	progress.Stop()
	progress.Hide()

	suppressChanges := false
	pendingRequestID := ""
	committedState := captureLiveAudioInputState(inputSelection, applicationSelection)

	setPending := func(pending bool) {
		if pending {
			inputSelection.Disable()
			applicationSelection.Disable()
			progress.Show()
			progress.Start()
			return
		}
		progress.Stop()
		progress.Hide()
		inputSelection.Enable()
		applicationSelection.Enable()
	}

	restoreCommittedState := func() {
		suppressChanges = true
		defer func() { suppressChanges = false }()
		ensureLiveAudioOption(inputSelection, committedState.input)
		inputSelection.SetSelected(committedState.input.Value)
		if Utilities.IsAudioApplicationOptionValue(committedState.input.Value) {
			applicationSelection.Show()
			if committedState.application != nil {
				ensureLiveAudioOption(applicationSelection, *committedState.application)
				applicationSelection.SetSelected(committedState.application.Value)
			}
		} else {
			applicationSelection.Hide()
		}
	}

	sendSelection := func() {
		if suppressChanges || pendingRequestID != "" {
			return
		}
		selectedInput := inputSelection.GetSelected()
		request, ok := buildLiveAudioInputSwitchRequest(
			fmt.Sprintf("%d", time.Now().UnixNano()),
			backend.Name,
			selectedInput,
			applicationSelection.GetSelected(),
		)
		if !ok {
			return
		}

		pendingRequestID = request.RequestID
		setPending(true)
		SendMessageChannel.SendMessageStruct{
			Type:  "audio_input_switch",
			Value: request,
		}.SendMessage()
	}

	inputSelection.OnChanged = func(option CustomWidget.TextValueOption) {
		if suppressChanges {
			return
		}
		if Utilities.IsAudioApplicationOptionValue(option.Value) {
			applicationSelection.Show()
			if applicationSelection.GetSelected() == nil {
				return
			}
		} else {
			applicationSelection.Hide()
		}
		sendSelection()
	}
	applicationSelection.OnChanged = func(_ CustomWidget.TextValueOption) {
		if suppressChanges {
			return
		}
		if selected := inputSelection.GetSelected(); selected != nil && Utilities.IsAudioApplicationOptionValue(selected.Value) {
			sendSelection()
		}
	}

	inputSelection.BeforeTapped = func() {
		previous := inputSelection.GetSelected()
		suppressChanges = true
		defer func() { suppressChanges = false }()
		inputSelection.SetValueOptions(liveAudioInputOptions(backend.Backend))
		if previous != nil {
			if inputSelection.ContainsEntry(previous, CustomWidget.CompareValue) {
				inputSelection.SetSelected(previous.Value)
				return
			}
			for _, option := range inputSelection.Options {
				if strings.EqualFold(option.Text, previous.Text) {
					inputSelection.SetSelected(option.Value)
					return
				}
			}
		}
		inputSelection.SetSelected("-1")
	}
	applicationSelection.BeforeTapped = func() {
		refreshApplicationCaptureOptions(applicationSelection)
	}

	Fields.AudioInputSwitchResult = func(requestID string, success bool, errorMessage string) {
		if requestID == "" || requestID != pendingRequestID {
			return
		}
		pendingRequestID = ""
		setPending(false)
		if success {
			committedState = captureLiveAudioInputState(inputSelection, applicationSelection)
			return
		}
		restoreCommittedState()
		if strings.TrimSpace(errorMessage) == "" {
			errorMessage = lang.L("Error")
		}
		if app := fyne.CurrentApp(); app != nil && len(app.Driver().AllWindows()) > 0 {
			dialog.ShowError(errors.New(errorMessage), app.Driver().AllWindows()[0])
		}
	}

	return container.NewVBox(inputSelection, applicationSelection, progress)
}

func createLiveAudioOutputControl() fyne.CanvasObject {
	backend := AudioAPI.GetAudioBackendByName(Settings.Config.Audio_api)
	outputSelection := CustomWidget.NewTextValueSelect(
		"live_audio_output",
		liveAudioOutputOptions(backend.Backend),
		nil,
		-1,
	)
	selectConfiguredLiveAudioOutput(outputSelection)

	progress := widget.NewProgressBarInfinite()
	progress.Stop()
	progress.Hide()

	suppressChanges := false
	pendingRequestID := ""
	committedOutput := CustomWidget.TextValueOption{Text: "Default", Value: "-1"}
	if selected := outputSelection.GetSelected(); selected != nil {
		committedOutput = *selected
	}

	setPending := func(pending bool) {
		if pending {
			outputSelection.Disable()
			progress.Show()
			progress.Start()
			return
		}
		progress.Stop()
		progress.Hide()
		outputSelection.Enable()
	}

	restoreCommittedOutput := func() {
		suppressChanges = true
		defer func() { suppressChanges = false }()
		ensureLiveAudioOption(outputSelection, committedOutput)
		outputSelection.SetSelected(committedOutput.Value)
	}

	outputSelection.OnChanged = func(option CustomWidget.TextValueOption) {
		if suppressChanges || pendingRequestID != "" {
			return
		}
		request, ok := buildLiveAudioOutputSwitchRequest(
			fmt.Sprintf("%d", time.Now().UnixNano()),
			backend.Name,
			&option,
		)
		if !ok {
			return
		}
		pendingRequestID = request.RequestID
		setPending(true)
		SendMessageChannel.SendMessageStruct{
			Type:  "audio_output_switch",
			Value: request,
		}.SendMessage()
	}

	outputSelection.BeforeTapped = func() {
		previous := outputSelection.GetSelected()
		suppressChanges = true
		defer func() { suppressChanges = false }()
		outputSelection.SetValueOptions(liveAudioOutputOptions(backend.Backend))
		if previous != nil {
			if outputSelection.ContainsEntry(previous, CustomWidget.CompareValue) {
				outputSelection.SetSelected(previous.Value)
				return
			}
			for _, option := range outputSelection.Options {
				if strings.EqualFold(option.Text, previous.Text) {
					outputSelection.SetSelected(option.Value)
					return
				}
			}
		}
		outputSelection.SetSelected("-1")
	}

	Fields.AudioOutputSwitchResult = func(requestID string, success bool, errorMessage string) {
		if requestID == "" || requestID != pendingRequestID {
			return
		}
		pendingRequestID = ""
		setPending(false)
		if success {
			if selected := outputSelection.GetSelected(); selected != nil {
				committedOutput = *selected
			}
			return
		}
		restoreCommittedOutput()
		if strings.TrimSpace(errorMessage) == "" {
			errorMessage = lang.L("Error")
		}
		if app := fyne.CurrentApp(); app != nil && len(app.Driver().AllWindows()) > 0 {
			dialog.ShowError(errors.New(errorMessage), app.Driver().AllWindows()[0])
		}
	}

	return container.NewVBox(outputSelection, progress)
}

func createLiveAudioDeviceSettings() fyne.CanvasObject {
	return container.New(
		layout.NewFormLayout(),
		widget.NewLabelWithStyle(
			lang.L("Audio Input (mic)")+":",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		createLiveAudioInputControl(),
		widget.NewLabelWithStyle(
			lang.L("Audio Output (speaker)")+":",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		createLiveAudioOutputControl(),
	)
}
