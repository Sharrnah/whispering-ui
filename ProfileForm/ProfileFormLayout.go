package ProfileForm

import (
	"errors"
	"fmt"
	"reflect"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Pages/SettingsMappings"
	"whispering-tiger-ui/Utilities/Hardwareinfo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type RowSpec struct {
	Spacer       bool
	CustomKey    string
	Label        string
	Hint         string
	ControlNames []string
	Cols         int
}

func BuildDefaultProfileLayout() []RowSpec {
	rows := make([]RowSpec, 0, 48)
	rows = append(rows,
		RowSpec{Label: lang.L("Websocket IP + Port"), Hint: lang.L("IP + Port of the websocket server the backend will start and the UI will connect to."), ControlNames: []string{"WebsocketIP", "WebsocketPort", "RunBackend"}, Cols: 3},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Audio API"), ControlNames: []string{"AudioAPI"}, Cols: 1},
		RowSpec{Label: lang.L("Audio Input (mic)"), ControlNames: []string{"AudioInput"}, Cols: 1},
		RowSpec{CustomKey: "AudioInputProgress"},
		RowSpec{Label: lang.L("Audio Output (speaker)"), ControlNames: []string{"AudioOutput"}, Cols: 1},
		RowSpec{CustomKey: "AudioOutputProgress"},
	)
	rows = append(rows,
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Speech-to-Text Type"), ControlNames: []string{"STTType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Speech-to-Text"), ControlNames: []string{"STTDevice"}, Cols: 1},
		RowSpec{Label: lang.L("Speech-to-Text A.I. Size"), ControlNames: []string{"STTModelSize", "STTPrecision"}, Cols: 2},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Text-Translation Type"), ControlNames: []string{"TxtType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Text-Translation"), ControlNames: []string{"TxtDevice"}, Cols: 1},
		RowSpec{Label: lang.L("Text-Translation A.I. Size"), ControlNames: []string{"TxtSize", "TxtPrecision"}, Cols: 2},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Integrated Text-to-Speech"), ControlNames: []string{"TTSType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Text-to-Speech"), ControlNames: []string{"TTSDevice"}, Cols: 1},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Integrated Image-to-Text"), ControlNames: []string{"OCRType"}, Cols: 1},
		RowSpec{Label: lang.L("A.I. Device for Image-to-Text"), ControlNames: []string{"OCRDevice", "OCRPrecision"}, Cols: 2},
	)
	return rows
}

func BuildFullProfileLayout() []RowSpec {
	rows := make([]RowSpec, 0, 64)
	rows = append(rows,
		RowSpec{Label: lang.L("Websocket IP + Port"), Hint: lang.L("IP + Port of the websocket server the backend will start and the UI will connect to."), ControlNames: []string{"WebsocketIP", "WebsocketPort", "RunBackend"}, Cols: 3},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Audio API"), ControlNames: []string{"AudioAPI"}, Cols: 1},
		RowSpec{Label: lang.L("Audio Input (mic)"), ControlNames: []string{"AudioInput"}, Cols: 1},
		RowSpec{CustomKey: "AudioInputProgress"},
		RowSpec{Label: lang.L("Audio Output (speaker)"), ControlNames: []string{"AudioOutput"}, Cols: 1},
		RowSpec{CustomKey: "AudioOutputProgress"},
	)
	rows = append(rows,
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("VAD (Voice activity detection)"), Hint: lang.L("Press ESC in Push to Talk field to clear the keybinding."), CustomKey: "VADGroup"},
		RowSpec{Label: lang.L("vad_confidence_threshold.Name"), Hint: lang.L("The confidence level required to detect speech."), CustomKey: "VADConfidence"},
	)
	rows = append(rows,
		RowSpec{Label: lang.L("energy.Name"), Hint: lang.L("The volume level at which the speech detection will trigger. (0 = Disabled, useful for Push2Talk)"), CustomKey: "EnergyRow"},
		RowSpec{Label: lang.L("Noise Filter"), Hint: lang.L("Requires a restart when switching the Noise Filter type. Disabling will stop applying it even without restart."), ControlNames: []string{"DenoiseAudio"}, Cols: 1},
		RowSpec{Label: lang.L("pause.Name"), Hint: lang.L("pause.Description"), CustomKey: "PauseRow"},
		RowSpec{Label: lang.L("phrase_time_limit.Name"), Hint: lang.L("phrase_time_limit.Description"), CustomKey: "PhraseRow"},
	)
	rows = append(rows,
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Speech-to-Text Type"), ControlNames: []string{"STTType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Speech-to-Text"), ControlNames: []string{"STTDevice"}, Cols: 1},
		RowSpec{Label: lang.L("Speech-to-Text A.I. Size"), ControlNames: []string{"STTModelSize", "STTPrecision"}, Cols: 2},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Text-Translation Type"), ControlNames: []string{"TxtType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Text-Translation"), ControlNames: []string{"TxtDevice"}, Cols: 1},
		RowSpec{Label: lang.L("Text-Translation A.I. Size"), ControlNames: []string{"TxtSize", "TxtPrecision"}, Cols: 2},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Integrated Text-to-Speech"), ControlNames: []string{"TTSType"}, Cols: 2},
		RowSpec{Label: lang.L("A.I. Device for Text-to-Speech"), ControlNames: []string{"TTSDevice"}, Cols: 1},
		RowSpec{Spacer: true},
		RowSpec{Label: lang.L("Integrated Image-to-Text"), ControlNames: []string{"OCRType"}, Cols: 1},
		RowSpec{Label: lang.L("A.I. Device for Image-to-Text"), ControlNames: []string{"OCRDevice", "OCRPrecision"}, Cols: 2},
	)
	return rows
}

type FullFormDeps struct {
	InputOptions         []CustomWidget.TextValueOption
	OutputOptions        []CustomWidget.TextValueOption
	AudioInputProgress   fyne.CanvasObject
	AudioOutputProgress  fyne.CanvasObject
	OnAudioAPIChanged    func(CustomWidget.TextValueOption)
	OnAudioInputChanged  func(CustomWidget.TextValueOption)
	OnAudioOutputChanged func(CustomWidget.TextValueOption)
	OnDetectEnergy       func(apiValue, deviceIndexValue, deviceText string) (float64, error)
	AfterDetectEnergy    func()
	CPUMemoryBar         *widget.ProgressBar
	GPUMemoryBar         *widget.ProgressBar
	TotalGPUMemory       func() int64
	HasNvidiaGPU         func() bool
}

func BuildAndRenderFullProfile(form *widget.Form, engine *FormEngine, deps FullFormDeps) *AllProfileControls {
	builder := NewProfileBuilder()

	_ = builder.BuildConnectionSection(engine)

	audioSection := builder.BuildAudioSection(engine, deps.InputOptions, deps.OutputOptions)
	if audioSection.ApiSelect != nil && deps.OnAudioAPIChanged != nil {
		audioSection.ApiSelect.OnChanged = deps.OnAudioAPIChanged
	}
	if audioSection.InputSelect != nil && deps.OnAudioInputChanged != nil {
		audioSection.InputSelect.OnChanged = deps.OnAudioInputChanged
	}
	if audioSection.OutputSelect != nil && deps.OnAudioOutputChanged != nil {
		audioSection.OutputSelect.OnChanged = deps.OnAudioOutputChanged
	}

	vadSection := builder.BuildVADSection(engine)

	energyState := widget.NewLabel("0.0")
	energySlider := widget.NewSlider(0, SettingsMappings.EnergySliderMax)
	engine.Controls.Energy = energySlider
	engine.Register("energy", energySlider)
	energyBtn := widget.NewButtonWithIcon(lang.L("Autodetect"), theme.VolumeUpIcon(), func() {
		if deps.OnDetectEnergy == nil {
			return
		}
		selAPI, selInput := "", ""
		if audioSection.ApiSelect != nil && audioSection.ApiSelect.GetSelected() != nil {
			selAPI = audioSection.ApiSelect.GetSelected().Value
		}
		if audioSection.InputSelect != nil && audioSection.InputSelect.GetSelected() != nil {
			selInput = audioSection.InputSelect.GetSelected().Value
		}
		selInputText := ""
		if audioSection.InputSelect != nil && audioSection.InputSelect.GetSelected() != nil {
			selInputText = audioSection.InputSelect.GetSelected().Text
		}
		busy := dialog.NewCustomWithoutButtons(lang.L("Detecting..."), widget.NewProgressBarInfinite(), fyne.CurrentApp().Driver().AllWindows()[1])
		busy.Show()
		go func() {
			val, err := deps.OnDetectEnergy(selAPI, selInput, selInputText)
			fyne.Do(func() {
				busy.Hide()
				if err != nil {
					dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[1])
					return
				}
				energySlider.Max = SettingsMappings.EnergySliderMax
				if val >= energySlider.Max {
					energySlider.Max = val + 200
				}
				energySlider.SetValue(val + 20)
				if deps.AfterDetectEnergy != nil {
					deps.AfterDetectEnergy()
				}
			})
		}()
	})
	energyRow := container.NewBorder(nil, nil, nil, container.NewHBox(energyState, energyBtn), energySlider)

	pauseState := widget.NewLabel("0.0")
	pauseSlider := widget.NewSlider(0, 5)
	pauseSlider.Step = 0.1
	engine.Controls.PauseSeconds = pauseSlider
	engine.Register("pause", pauseSlider)
	pauseRow := container.NewBorder(nil, nil, nil, pauseState, pauseSlider)

	phraseState := widget.NewLabel("0.0")
	phraseSlider := widget.NewSlider(0, 30)
	phraseSlider.Step = 0.1
	engine.Controls.PhraseTimeLimit = phraseSlider
	engine.Register("phrase_time_limit", phraseSlider)
	phraseRow := container.NewBorder(nil, nil, nil, phraseState, phraseSlider)

	denoiseSelect := CustomWidget.NewTextValueSelect("denoise_audio", []CustomWidget.TextValueOption{{Text: lang.L("Disabled"), Value: ""}, {Text: "Noise Reduce", Value: "noise_reduce"}, {Text: "DeepFilterNet", Value: "deepfilter"}}, func(_ CustomWidget.TextValueOption) {}, 0)
	engine.Controls.DenoiseAudio = denoiseSelect
	engine.Register("denoise_audio", denoiseSelect)

	stt := builder.BuildSTTSection(engine)
	if stt != nil {
		stt.TypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.ApplySTTTypeChange(s.Value)
			}
		}
		stt.DeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", Device: s.Value}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil {
				prec := ""
				if stt.PrecisionSelect.GetSelected() != nil {
					prec = stt.PrecisionSelect.GetSelected().Value
				}
				engine.Coord.EnsurePrecisionDeviceCompatibility(s.Value, prec)
				if !engine.Coord.InProgrammaticUpdate {
					engine.Coord.HandleMultiModalAllSync()
				}
			}
		}
		stt.PrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16":
				precisionType = Hardwareinfo.Float16
			case "int32":
				precisionType = Hardwareinfo.Int32
			case "int16":
				precisionType = Hardwareinfo.Int16
			case "int8_float16", "int8", "int8_bfloat16":
				precisionType = Hardwareinfo.Int8
			case "bfloat16":
				precisionType = Hardwareinfo.Float16
			case "8bit":
				precisionType = Hardwareinfo.Bit8
			case "4bit":
				precisionType = Hardwareinfo.Bit4
			}
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", Precision: precisionType}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil && stt.DeviceSelect.GetSelected() != nil {
				engine.Coord.EnsurePrecisionDeviceCompatibility(stt.DeviceSelect.GetSelected().Value, s.Value)
				if !engine.Coord.InProgrammaticUpdate {
					engine.Coord.HandleMultiModalAllSync()
				}
			}
		}
		stt.SizeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", AIModelSize: s.Value}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.HandleMultiModalAllSync()
			}
		}
	}

	txt := builder.BuildTXTSection(engine)
	if txt != nil {
		txt.DeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if s.Value == "cuda" && deps.HasNvidiaGPU != nil && !deps.HasNvidiaGPU() && (engine.Coord == nil || !engine.Coord.InProgrammaticUpdate) {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if txt.PrecisionSelect.GetSelected() != nil {
				if s.Value == "cpu" && (txt.PrecisionSelect.GetSelected().Value == "float16" || txt.PrecisionSelect.GetSelected().Value == "int8_float16" || txt.PrecisionSelect.GetSelected().Value == "bfloat16" || txt.PrecisionSelect.GetSelected().Value == "int8_bfloat16") {
					txt.PrecisionSelect.SetSelected("float32")
				}
				if s.Value == "cuda" && txt.PrecisionSelect.GetSelected().Value == "int16" {
					txt.PrecisionSelect.SetSelected("float16")
				}
			}
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "TxtTranslator", Device: s.Value}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil {
				prec := ""
				if txt.PrecisionSelect.GetSelected() != nil {
					prec = txt.PrecisionSelect.GetSelected().Value
				}
				engine.Coord.EnsurePrecisionDeviceCompatibility(s.Value, prec)
				if !engine.Coord.InProgrammaticUpdate {
					engine.Coord.HandleMultiModalAllSync()
				}
			}
		}
		txt.PrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16":
				precisionType = Hardwareinfo.Float16
			case "int32":
				precisionType = Hardwareinfo.Int32
			case "int16":
				precisionType = Hardwareinfo.Int16
			case "int8_float16", "int8", "int8_bfloat16":
				precisionType = Hardwareinfo.Int8
			case "bfloat16":
				precisionType = Hardwareinfo.Float16
			case "8bit":
				precisionType = Hardwareinfo.Bit8
			case "4bit":
				precisionType = Hardwareinfo.Bit4
			}
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "TxtTranslator", Precision: precisionType}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.HandleMultiModalAllSync()
			}
		}
		txt.SizeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "TxtTranslator", AIModelSize: s.Value}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.HandleMultiModalAllSync()
			}
		}
		txt.TypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.ApplyTXTTypeChange(s.Value)
			}
		}
	}

	tts := builder.BuildTTSSection(engine)
	if tts != nil {
		tts.DeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if s.Value == "cuda" && deps.HasNvidiaGPU != nil && !deps.HasNvidiaGPU() && (engine.Coord == nil || !engine.Coord.InProgrammaticUpdate) {
				dialog.ShowInformation(lang.L("No NVIDIA Card found"), lang.L("No NVIDIA Card found. You might need to use CPU instead for it to work."), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ttsType", Device: s.Value, Precision: Hardwareinfo.Float32}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
		}
		tts.TypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.ApplyTTSTypeChange(s.Value)
			}
		}
	}

	ocr := builder.BuildOCRSection(engine)
	if ocr != nil {
		ocr.TypeSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.ApplyOCRTypeChange(s.Value)
			}
		}
		ocr.DeviceSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", Device: s.Value}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
		}
		ocr.PrecisionSelect.OnChanged = func(s CustomWidget.TextValueOption) {
			precisionType := Hardwareinfo.Float32
			switch s.Value {
			case "float32":
				precisionType = Hardwareinfo.Float32
			case "float16", "bfloat16":
				precisionType = Hardwareinfo.Float16
			}
			total := int64(0)
			if deps.TotalGPUMemory != nil {
				total = deps.TotalGPUMemory()
			}
			AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", Precision: precisionType}
			AIModel.CalculateMemoryConsumption(deps.CPUMemoryBar, deps.GPUMemoryBar, total)
			if engine.Coord != nil && !engine.Coord.InProgrammaticUpdate {
				engine.Coord.HandleMultiModalAllSync()
			}
		}
	}

	zeroEnergyInfo := dialog.NewError(errors.New(lang.L("You did set Speech volume level to 0 and have no PushToTalk Button set.This would prevent the app from recording anything.")), fyne.CurrentApp().Driver().AllWindows()[1])
	energySlider.OnChanged = func(v float64) {
		if v >= energySlider.Max {
			energySlider.Max += 10
		}
		energyState.SetText(fmt.Sprintf("%.0f", v))
		if engine.Controls.PushToTalk.Text == "" && v == 0 {
			energySlider.SetValue(1)
			zeroEnergyInfo.Show()
		}
	}
	zeroPauseInfo := dialog.NewError(errors.New(lang.L("You did set Speech pause detection to 0 and have no PushToTalk Button set.This would prevent the app from stopping recording automatically.")), fyne.CurrentApp().Driver().AllWindows()[1])
	pauseSlider.OnChanged = func(v float64) {
		pauseState.SetText(fmt.Sprintf("%.1f", v))
		if engine.Controls.PushToTalk.Text == "" && v == 0 {
			pauseSlider.SetValue(0.5)
			zeroPauseInfo.Show()
		}
	}
	phraseSlider.OnChanged = func(v float64) { phraseState.SetText(fmt.Sprintf("%.1f", v)) }

	custom := map[string]fyne.CanvasObject{"AudioInputProgress": deps.AudioInputProgress, "AudioOutputProgress": deps.AudioOutputProgress, "VADGroup": vadSection.GroupRow, "VADConfidence": vadSection.ConfidenceRow, "EnergyRow": energyRow, "PauseRow": pauseRow, "PhraseRow": phraseRow}
	AppendProfileLayout(form, engine.Controls, BuildFullProfileLayout(), custom)

	engine.Controls.VadEnable.OnChanged = func(b bool) {
		if b {
			pauseSlider.Min = 0.0
			phraseSlider.Min = 0.0
			engine.Controls.VadConfidence.Show()
			engine.Controls.VadRealtime.Show()
			vadSection.PushToTalkBlock.Show()
		} else {
			engine.Controls.VadConfidence.Hide()
			engine.Controls.VadOnFullClip.Hide()
			engine.Controls.VadRealtime.Hide()
			vadSection.PushToTalkBlock.Hide()
			if audioSection.ApiSelect != nil && audioSection.ApiSelect.Selected != "MME" && (engine.Coord == nil || engine.Coord.IsLoadingSettings == nil || !*engine.Coord.IsLoadingSettings) {
				dialog.ShowInformation(lang.L("Information"), lang.L("Disabled VAD is only supported with MME Audio API. Please make sure MME is selected as audio API. (Enabling VAD is highly recommended)"), fyne.CurrentApp().Driver().AllWindows()[1])
			}
			if (pauseSlider.Value == 0 || phraseSlider.Value == 0) && (engine.Coord == nil || engine.Coord.IsLoadingSettings == nil || !*engine.Coord.IsLoadingSettings) {
				dialog.ShowInformation(lang.L("Information"), lang.L("You disabled VAD but have set the pause or phrase limit to 0. This is not supported. Setting Pause and Phrase limits to non-zero values."), fyne.CurrentApp().Driver().AllWindows()[1])
				if pauseSlider.Value == 0 {
					pauseSlider.SetValue(1.2)
				}
				if phraseSlider.Value == 0 {
					phraseSlider.SetValue(30)
				}
			}
			pauseSlider.Min = 0.1
			phraseSlider.Min = 0.1
		}
	}

	var pushToTalkChanged bool
	engine.Controls.PushToTalk.OnChanged = func(s string) {
		if s != "" && (engine.Coord == nil || engine.Coord.IsLoadingSettings == nil || !*engine.Coord.IsLoadingSettings) {
			pushToTalkChanged = true
		}
	}
	engine.Controls.PushToTalk.OnFocusChanged = func(focusGained bool) {
		if !focusGained && pushToTalkChanged && engine.Controls.PushToTalk.Text != "" {
			dialog.NewConfirm(lang.L("Change speech trigger settings?"), lang.L("You did set a PushToTalk Button. Do you want to set settings to trigger with only a Button press?"), func(b bool) {
				if b {
					engine.Controls.Energy.SetValue(0)
					pauseSlider.SetValue(0)
					phraseSlider.SetValue(0)
				}
			}, fyne.CurrentApp().Driver().AllWindows()[1]).Show()
			pushToTalkChanged = false
		}
	}

	return engine.Controls
}

func AppendProfileLayout(form *widget.Form, controls any, rows []RowSpec, custom map[string]fyne.CanvasObject) {
	for _, row := range rows {
		if row.Spacer {
			form.Append("", layout.NewSpacer())
			continue
		}
		if row.CustomKey != "" {
			if obj, ok := custom[row.CustomKey]; ok && obj != nil {
				appendWidgetToForm(form, row.Label, obj, row.Hint)
			}
			continue
		}
		widgets := make([]fyne.CanvasObject, 0, len(row.ControlNames))
		for _, name := range row.ControlNames {
			if w := resolveControlByName(controls, name); w != nil {
				widgets = append(widgets, w)
			}
		}
		var content fyne.CanvasObject
		if row.Cols > 1 {
			content = container.NewGridWithColumns(row.Cols, widgets...)
		} else {
			if len(widgets) > 0 {
				content = widgets[0]
			} else {
				content = widget.NewLabel("")
			}
		}
		appendWidgetToForm(form, row.Label, content, row.Hint)
	}
}

func resolveControlByName(ctrls any, field string) fyne.CanvasObject {
	if ctrls == nil || field == "" {
		return nil
	}
	rv := reflect.ValueOf(ctrls)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	f := rv.FieldByName(field)
	if !f.IsValid() || f.IsNil() {
		return nil
	}
	switch v := f.Interface().(type) {
	case *widget.Entry:
		return v
	case *widget.Check:
		return v
	case *widget.Slider:
		return v
	case *CustomWidget.TextValueSelect:
		return v
	case *CustomWidget.HotKeyEntry:
		return v
	default:
		if w, ok := f.Interface().(fyne.CanvasObject); ok {
			return w
		}
	}
	return nil
}

func appendWidgetToForm(form *widget.Form, text string, itemWidget fyne.CanvasObject, hintText string) {
	item := &widget.FormItem{Text: text, Widget: itemWidget, HintText: hintText}
	form.AppendItem(item)
}
