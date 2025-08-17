package ProfileForm

import (
	"fmt"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/AudioAPI"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProfileBuilder constructs the profile form sections and registers controls
type ProfileBuilder struct{}

func NewProfileBuilder() *ProfileBuilder { return &ProfileBuilder{} }

// --- small helpers to reduce repetition across section builders ---

// newEntry creates an Entry with default text and registers it under key
func (b *ProfileBuilder) newEntry(engine *FormEngine, key, defaultText string) *widget.Entry {
	e := widget.NewEntry()
	if defaultText != "" {
		e.SetText(defaultText)
	}
	engine.Register(key, e)
	return e
}

// newCheck creates a Check with label and registers it under key
func (b *ProfileBuilder) newCheck(engine *FormEngine, key, label string) *widget.Check {
	c := widget.NewCheck(label, func(bool) {})
	engine.Register(key, c)
	return c
}

// newSelect creates a TextValueSelect with options and registers it under key
func (b *ProfileBuilder) newSelect(engine *FormEngine, key string, opts []CustomWidget.TextValueOption) *CustomWidget.TextValueSelect {
	s := CustomWidget.NewTextValueSelect(key, opts, func(_ CustomWidget.TextValueOption) {}, 0)
	engine.Register(key, s)
	return s
}

// sliderWithLabel creates a slider, binds a right-aligned label to its value and registers the slider under key
func (b *ProfileBuilder) sliderWithLabel(engine *FormEngine, key string, min, max, step float64) (*widget.Slider, *widget.Label, fyne.CanvasObject) {
	lbl := widget.NewLabel("0.00")
	s := widget.NewSlider(min, max)
	s.Step = step
	s.OnChanged = func(v float64) { lbl.SetText(fmt.Sprintf("%.2f", v)) }
	engine.Register(key, s)
	row := container.NewBorder(nil, nil, nil, lbl, s)
	return s, lbl, row
}

// BuildResult collects custom UI rows constructed by the builder, where needed for layout
type BuildResult struct {
	VADGroupRow        fyne.CanvasObject
	VADConfidenceRow   fyne.CanvasObject
	VADPushToTalkBlock fyne.CanvasObject
}

// BuildAll creates and registers all controls and custom rows in one pass
func (b *ProfileBuilder) BuildAll(engine *FormEngine, inputOptions, outputOptions []CustomWidget.TextValueOption) *BuildResult {
	if engine == nil || engine.Controls == nil {
		return &BuildResult{}
	}
	// Connection
	wsIP := b.newEntry(engine, "websocket_ip", "127.0.0.1")
	wsPort := b.newEntry(engine, "websocket_port", "5000")
	runBackend := b.newCheck(engine, "run_backend", lang.L("Run Backend"))
	engine.Controls.WebsocketIP, engine.Controls.WebsocketPort, engine.Controls.RunBackend = wsIP, wsPort, runBackend

	// Audio selects
	audioOptions := make([]CustomWidget.TextValueOption, 0, len(AudioAPI.AudioBackends))
	for _, backend := range AudioAPI.AudioBackends {
		audioOptions = append(audioOptions, CustomWidget.TextValueOption{Text: backend.Name, Value: backend.Name})
	}
	audioApiSelect := b.newSelect(engine, "audio_api", audioOptions)
	audioInputSelect := b.newSelect(engine, "device_index", inputOptions)
	audioOutputSelect := b.newSelect(engine, "device_out_index", outputOptions)
	engine.Controls.AudioAPI, engine.Controls.AudioInput, engine.Controls.AudioOutput = audioApiSelect, audioInputSelect, audioOutputSelect

	// VAD
	vadEnable := b.newCheck(engine, "vad_enabled", lang.L("Enable"))
	vadOnFullClip := b.newCheck(engine, "vad_on_full_clip", "+ Check on Full Clip")
	vadOnFullClip.Hide()
	vadRealtime := b.newCheck(engine, "realtime", lang.L("Realtime"))
	pushToTalk := CustomWidget.NewHotKeyEntry()
	pushToTalk.PlaceHolder = lang.L("Keypress")
	pushToTalkBlock := container.NewBorder(nil, nil, container.NewHBox(widget.NewLabel(lang.L("Push to Talk")), widget.NewIcon(theme.ComputerIcon())), nil, pushToTalk)
	vadGroup := container.NewGridWithColumns(3, vadEnable, vadOnFullClip, vadRealtime, pushToTalkBlock)
	confSlider, confLabel, confRow := b.sliderWithLabel(engine, "vad_confidence_threshold", 0, 1, 0.01)
	_ = confLabel // label managed by sliderWithLabel
	engine.Controls.VadEnable, engine.Controls.VadOnFullClip, engine.Controls.VadRealtime = vadEnable, vadOnFullClip, vadRealtime
	engine.Controls.PushToTalk, engine.Controls.VadConfidence = pushToTalk, confSlider
	engine.Register("push_to_talk_key", pushToTalk)

	// STT
	sttDevice := b.newSelect(engine, "ai_device", DefaultDeviceOptions())
	sttPrecision := b.newSelect(engine, "Precision", GenericWhisperPrecisionOptions())
	sttType := b.newSelect(engine, "stt_type", STTTypeOptions())
	sttOpts, _, _ := STTModelOptions("faster_whisper")
	sttModel := b.newSelect(engine, "model", sttOpts)
	engine.Controls.STTDevice, engine.Controls.STTPrecision, engine.Controls.STTType, engine.Controls.STTModelSize = sttDevice, sttPrecision, sttType, sttModel
	engine.Register("whisper_precision", sttPrecision)

	// TXT
	txtType := b.newSelect(engine, "txt_translator", TXTTypeOptions())
	txtDevice := b.newSelect(engine, "txt_translator_device", DefaultDeviceOptions())
	sOpts, _, _ := TXTSizeOptions("NLLB200_CT2")
	txtSize := b.newSelect(engine, "txt_translator_size", sOpts)
	txtPrecision := b.newSelect(engine, "txt_translator_precision", GenericTextPrecisionOptions())
	engine.Controls.TxtType, engine.Controls.TxtDevice, engine.Controls.TxtSize, engine.Controls.TxtPrecision = txtType, txtDevice, txtSize, txtPrecision

	// TTS
	ttsType := b.newSelect(engine, "tts_type", TTSTypeOptions())
	ttsDevice := b.newSelect(engine, "tts_ai_device", DefaultDeviceOptions())
	engine.Controls.TTSType, engine.Controls.TTSDevice = ttsType, ttsDevice

	// OCR
	ocrType := b.newSelect(engine, "ocr_type", OcrTypeOptions())
	ocrDevice := b.newSelect(engine, "ocr_ai_device", DefaultDeviceOptions())
	ocrPrecision := b.newSelect(engine, "ocr_precision", GenericOcrPrecisionOptions())
	engine.Controls.OCRType, engine.Controls.OCRDevice, engine.Controls.OCRPrecision = ocrType, ocrDevice, ocrPrecision

	return &BuildResult{VADGroupRow: vadGroup, VADConfidenceRow: confRow, VADPushToTalkBlock: pushToTalkBlock}
}
