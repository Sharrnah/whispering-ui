package Pages

import (
	"fmt"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/AudioAPI"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProfileUI encapsulates the built UI and references for external wiring
type ProfileUI struct {
	Form         *widget.Form
	SubmitButton *widget.Button
	HelpContent  *fyne.Container
	ListContent  *fyne.Container
	MainContent  *container.Split
	Controls     *AllProfileControls
	Coord        *Coordinator
	Engine       *FormEngine
}

// ProfileBuilder is responsible for building the profile form using the existing
// control struct and schema helpers. It stays close to current layout but makes
// the construction self-contained.
type ProfileBuilder struct {
	ComputeCapability float32
}

func NewProfileBuilder() *ProfileBuilder { return &ProfileBuilder{} }

// BuildMinimal creates a barebones layout and attaches controls; remaining content
// continues to live in Profiles.go for now. This is a stepping stone to move more
// UI here progressively without breaking behavior.
func (b *ProfileBuilder) BuildMinimal() *ProfileUI {
	ui := &ProfileUI{Controls: &AllProfileControls{}}
	ui.Engine = NewFormEngine(ui.Controls, nil)

	ui.Form = widget.NewForm()
	// Minimal header row placeholder (others still in Profiles.go)
	wsIP := widget.NewEntry()
	wsIP.SetText("127.0.0.1")
	wsPort := widget.NewEntry()
	wsPort.SetText("5000")
	runBackend := widget.NewCheck("Run Backend", func(bool) {})
	ui.Controls.WebsocketIP, ui.Controls.WebsocketPort, ui.Controls.RunBackend = wsIP, wsPort, runBackend
	appendWidgetToForm(ui.Form, "Websocket IP + Port", container.NewGridWithColumns(3, wsIP, wsPort, runBackend), "")
	ui.Form.Append("", layout.NewSpacer())

	// Placeholder for list content and help content; actual content is still assembled in Profiles.go
	ui.SubmitButton = widget.NewButtonWithIcon("Save and Load Profile", theme.ConfirmIcon(), func() {})
	ui.HelpContent = container.NewVBox(widget.NewLabel(""))
	ui.ListContent = container.NewBorder(nil, ui.SubmitButton, nil, nil, container.NewVScroll(ui.Form))
	ui.ListContent.Hide()
	ui.MainContent = container.NewHSplit(container.NewStack(ui.HelpContent, ui.ListContent), widget.NewLabel(""))
	return ui
}

// BuildConnectionSection builds the Websocket IP/Port + Run Backend row and registers controls.
// Returns a single CanvasObject suitable to be used as a Form row widget (Grid with 3 columns).
func (b *ProfileBuilder) BuildConnectionSection(engine *FormEngine) fyne.CanvasObject {
	if engine == nil || engine.Controls == nil {
		return container.NewHBox()
	}
	wsIP := widget.NewEntry()
	wsIP.SetText("127.0.0.1")
	wsPort := widget.NewEntry()
	wsPort.SetText("5000")
	runBackend := widget.NewCheck(lang.L("Run Backend"), func(bool) {})

	engine.Controls.WebsocketIP = wsIP
	engine.Controls.WebsocketPort = wsPort
	engine.Controls.RunBackend = runBackend
	engine.Register("websocket_ip", wsIP)
	engine.Register("websocket_port", wsPort)
	engine.Register("run_backend", runBackend)

	return container.NewGridWithColumns(3, wsIP, wsPort, runBackend)
}

// AudioSection encapsulates the Audio UI widgets created by the builder
type AudioSection struct {
	ApiSelect    *CustomWidget.TextValueSelect
	InputSelect  *CustomWidget.TextValueSelect
	OutputSelect *CustomWidget.TextValueSelect
}

// BuildAudioSection creates Audio API / Input / Output selects and registers them
// inputOptions/outputOptions should be pre-fetched device option lists
func (b *ProfileBuilder) BuildAudioSection(engine *FormEngine, inputOptions, outputOptions []CustomWidget.TextValueOption) *AudioSection {
	if engine == nil || engine.Controls == nil {
		return &AudioSection{}
	}
	// Build Audio API options from known backends
	audioOptions := make([]CustomWidget.TextValueOption, 0, len(AudioAPI.AudioBackends))
	for _, backend := range AudioAPI.AudioBackends {
		audioOptions = append(audioOptions, CustomWidget.TextValueOption{Text: backend.Name, Value: backend.Name})
	}

	audioApiSelect := CustomWidget.NewTextValueSelect("audio_api", audioOptions, func(_ CustomWidget.TextValueOption) {}, 0)
	audioInputSelect := CustomWidget.NewTextValueSelect("device_index", inputOptions, func(_ CustomWidget.TextValueOption) {}, 0)
	audioOutputSelect := CustomWidget.NewTextValueSelect("device_out_index", outputOptions, func(_ CustomWidget.TextValueOption) {}, 0)

	// Register and wire to controls registry
	engine.Controls.AudioAPI = audioApiSelect
	engine.Controls.AudioInput = audioInputSelect
	engine.Controls.AudioOutput = audioOutputSelect
	engine.Register("audio_api", audioApiSelect)
	engine.Register("device_index", audioInputSelect)
	engine.Register("device_out_index", audioOutputSelect)

	return &AudioSection{ApiSelect: audioApiSelect, InputSelect: audioInputSelect, OutputSelect: audioOutputSelect}
}

// VADSection encapsulates the VAD-related UI widgets created by the builder
type VADSection struct {
	EnableCheck      *widget.Check
	OnFullClipCheck  *widget.Check
	RealtimeCheck    *widget.Check
	PushToTalk       *CustomWidget.HotKeyEntry
	PushToTalkBlock  fyne.CanvasObject
	ConfidenceLabel  *widget.Label
	ConfidenceSlider *widget.Slider
	ConfidenceRow    fyne.CanvasObject
	GroupRow         fyne.CanvasObject
}

// BuildVADSection creates the VAD enable/visibility group and confidence slider row, and registers them
func (b *ProfileBuilder) BuildVADSection(engine *FormEngine) *VADSection {
	if engine == nil || engine.Controls == nil {
		return &VADSection{}
	}
	vadEnable := widget.NewCheck(lang.L("Enable"), func(bool) {})
	vadOnFullClip := widget.NewCheck("+ Check on Full Clip", func(bool) {})
	vadOnFullClip.Hide() // remains hidden by default
	vadRealtime := widget.NewCheck(lang.L("Realtime"), func(bool) {})

	pushToTalk := CustomWidget.NewHotKeyEntry()
	pushToTalk.PlaceHolder = lang.L("Keypress")
	pushToTalkBlock := container.NewBorder(nil, nil, container.NewHBox(widget.NewLabel(lang.L("Push to Talk")), widget.NewIcon(theme.ComputerIcon())), nil, pushToTalk)

	groupRow := container.NewGridWithColumns(3, vadEnable, vadOnFullClip, vadRealtime, pushToTalkBlock)

	// Confidence slider with label
	confLabel := widget.NewLabel("0.00")
	confSlider := widget.NewSlider(0, 1)
	confSlider.Step = 0.01
	confSlider.OnChanged = func(v float64) { confLabel.SetText(fmt.Sprintf("%.2f", v)) }
	confRow := container.NewBorder(nil, nil, nil, confLabel, confSlider)

	// Register
	engine.Controls.VadEnable = vadEnable
	engine.Controls.VadOnFullClip = vadOnFullClip
	engine.Controls.VadRealtime = vadRealtime
	engine.Controls.PushToTalk = pushToTalk
	engine.Controls.VadConfidence = confSlider
	engine.Register("vad_enabled", vadEnable)
	engine.Register("vad_on_full_clip", vadOnFullClip)
	engine.Register("realtime", vadRealtime)
	engine.Register("push_to_talk_key", pushToTalk)
	engine.Register("vad_confidence_threshold", confSlider)

	return &VADSection{
		EnableCheck:      vadEnable,
		OnFullClipCheck:  vadOnFullClip,
		RealtimeCheck:    vadRealtime,
		PushToTalk:       pushToTalk,
		PushToTalkBlock:  pushToTalkBlock,
		ConfidenceLabel:  confLabel,
		ConfidenceSlider: confSlider,
		ConfidenceRow:    confRow,
		GroupRow:         groupRow,
	}
}

// STTSection encapsulates STT widgets
type STTSection struct {
	TypeSelect      *CustomWidget.TextValueSelect
	DeviceSelect    *CustomWidget.TextValueSelect
	PrecisionSelect *CustomWidget.TextValueSelect
	SizeSelect      *CustomWidget.TextValueSelect
}

// BuildSTTSection creates STT selects and registers them
func (b *ProfileBuilder) BuildSTTSection(engine *FormEngine) *STTSection {
	if engine == nil || engine.Controls == nil {
		return &STTSection{}
	}
	sttDevice := CustomWidget.NewTextValueSelect("ai_device", []CustomWidget.TextValueOption{
		{Text: "CUDA", Value: "cuda"},
		{Text: "CPU", Value: "cpu"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	sttPrecision := CustomWidget.NewTextValueSelect("Precision", []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "int16 " + lang.L("precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
		{Text: "8bit " + lang.L("precision"), Value: "8bit"},
		{Text: "4bit " + lang.L("precision"), Value: "4bit"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	sttType := CustomWidget.NewTextValueSelect("stt_type", []CustomWidget.TextValueOption{
		{Text: "Faster Whisper", Value: "faster_whisper"},
		{Text: "Original Whisper", Value: "original_whisper"},
		{Text: "Transformer Whisper", Value: "transformer_whisper"},
		{Text: "Seamless M4T", Value: "seamless_m4t"},
		{Text: "MMS", Value: "mms"},
		{Text: "Speech T5 (English only)", Value: "speech_t5"},
		{Text: "Wav2Vec Bert 2.0", Value: "wav2vec_bert"},
		{Text: "NeMo Canary", Value: "nemo_canary"},
		{Text: "Phi-4", Value: "phi4"},
		{Text: "Voxtral", Value: "voxtral"},
		{Text: lang.L("Disabled"), Value: ""},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	// initial model list (will be adapted by coordinator based on type)
	fasterWhisperModelList := []CustomWidget.TextValueOption{
		{Text: "Tiny", Value: "tiny"},
		{Text: "Tiny (English only)", Value: "tiny.en"},
		{Text: "Base", Value: "base"},
		{Text: "Base (English only)", Value: "base.en"},
		{Text: "Small", Value: "small"},
		{Text: "Small (English only)", Value: "small.en"},
		{Text: "Medium", Value: "medium"},
		{Text: "Medium (English only)", Value: "medium.en"},
		{Text: "Large V1", Value: "large-v1"},
		{Text: "Large V2", Value: "large-v2"},
		{Text: "Large V3", Value: "large-v3"},
		{Text: "Large V3 Turbo", Value: "large-v3-turbo"},
		{Text: "Medium Distilled (English)", Value: "medium-distilled.en"},
		{Text: "Large V2 Distilled (English)", Value: "large-distilled-v2.en"},
		{Text: "Large V3 Distilled (English)", Value: "large-distilled-v3.en"},
		{Text: "Large V3.5 Distilled (English)", Value: "large-distilled-v3.5.en"},
		{Text: "Crisper", Value: "crisper"},
		{Text: "Small (European finetune)", Value: "small.eu"},
		{Text: "Medium (European finetune)", Value: "medium.eu"},
		{Text: "Small (German finetune)", Value: "small.de"},
		{Text: "Medium (German finetune)", Value: "medium.de"},
		{Text: "Large V2 (German finetune)", Value: "large-v2.de2"},
		{Text: "Large V3 Distilled (German finetune)", Value: "large-distilled-v3.de"},
		{Text: "Small (German-Swiss finetune)", Value: "small.de-swiss"},
		{Text: "Medium (Mix-Japanese-v2 finetune)", Value: "medium.mix-jpv2"},
		{Text: "Large V2 (Mix-Japanese finetune)", Value: "large-v2.mix-jp"},
		{Text: "Small (Japanese finetune)", Value: "small.jp"},
		{Text: "Medium (Japanese finetune)", Value: "medium.jp"},
		{Text: "Large V2 (Japanese finetune)", Value: "large-v2.jp"},
		{Text: "Medium (Korean finetune)", Value: "medium.ko"},
		{Text: "Large V2 (Korean finetune)", Value: "large-v2.ko"},
		{Text: "Small (Chinese finetune)", Value: "small.zh"},
		{Text: "Medium (Chinese finetune)", Value: "medium.zh"},
		{Text: "Large V2 (Chinese finetune)", Value: "large-v2.zh"},
		{Text: "Custom (Place in '.cache/whisper/custom-ct2' directory)", Value: "custom"},
	}
	sttModel := CustomWidget.NewTextValueSelect("model", fasterWhisperModelList, func(_ CustomWidget.TextValueOption) {}, 0)

	engine.Controls.STTDevice = sttDevice
	engine.Controls.STTPrecision = sttPrecision
	engine.Controls.STTType = sttType
	engine.Controls.STTModelSize = sttModel
	engine.Register("ai_device", sttDevice)
	engine.Register("whisper_precision", sttPrecision)
	engine.Register("stt_type", sttType)
	engine.Register("model", sttModel)
	return &STTSection{TypeSelect: sttType, DeviceSelect: sttDevice, PrecisionSelect: sttPrecision, SizeSelect: sttModel}
}

// TXTSection encapsulates Text-Translator widgets
type TXTSection struct {
	TypeSelect      *CustomWidget.TextValueSelect
	DeviceSelect    *CustomWidget.TextValueSelect
	PrecisionSelect *CustomWidget.TextValueSelect
	SizeSelect      *CustomWidget.TextValueSelect
}

// BuildTXTSection creates TXT selects and registers them
func (b *ProfileBuilder) BuildTXTSection(engine *FormEngine) *TXTSection {
	if engine == nil || engine.Controls == nil {
		return &TXTSection{}
	}
	txtType := CustomWidget.NewTextValueSelect("txt_translator", []CustomWidget.TextValueOption{
		{Text: "Faster NLLB200 (200 languages)", Value: "NLLB200_CT2"},
		{Text: "Original NLLB200 (200 languages)", Value: "NLLB200"},
		{Text: "M2M100 (100 languages)", Value: "M2M100"},
		{Text: "Seamless M4T (101 languages)", Value: "seamless_m4t"},
		{Text: "Phi-4 (23 languages)", Value: "phi4"},
		{Text: "Voxtral (13 languages)", Value: "voxtral"},
		{Text: lang.L("Disabled"), Value: ""},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	txtDevice := CustomWidget.NewTextValueSelect("txt_translator_device", []CustomWidget.TextValueOption{
		{Text: "CUDA", Value: "cuda"},
		{Text: "CPU", Value: "cpu"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	txtSize := CustomWidget.NewTextValueSelect("txt_translator_size", []CustomWidget.TextValueOption{
		{Text: "Small", Value: "small"},
		{Text: "Medium", Value: "medium"},
		{Text: "Large", Value: "large"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	txtPrecision := CustomWidget.NewTextValueSelect("txt_translator_precision", []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "int16 " + lang.L("precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)

	engine.Controls.TxtType = txtType
	engine.Controls.TxtDevice = txtDevice
	engine.Controls.TxtSize = txtSize
	engine.Controls.TxtPrecision = txtPrecision
	engine.Register("txt_translator", txtType)
	engine.Register("txt_translator_device", txtDevice)
	engine.Register("txt_translator_size", txtSize)
	engine.Register("txt_translator_precision", txtPrecision)
	return &TXTSection{TypeSelect: txtType, DeviceSelect: txtDevice, PrecisionSelect: txtPrecision, SizeSelect: txtSize}
}

// TTSSection encapsulates TTS widgets
type TTSSection struct {
	TypeSelect   *CustomWidget.TextValueSelect
	DeviceSelect *CustomWidget.TextValueSelect
}

// BuildTTSSection creates TTS selects and registers them
func (b *ProfileBuilder) BuildTTSSection(engine *FormEngine) *TTSSection {
	if engine == nil || engine.Controls == nil {
		return &TTSSection{}
	}
	ttsType := CustomWidget.NewTextValueSelect("tts_type", []CustomWidget.TextValueOption{
		{Text: "Silero", Value: "silero"},
		{Text: "F5/E2", Value: "f5_e2"},
		{Text: "Zonos", Value: "zonos"},
		{Text: "Kokoro", Value: "kokoro"},
		{Text: "Orpheus", Value: "orpheus"},
		{Text: lang.L("Disabled"), Value: ""},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	ttsDevice := CustomWidget.NewTextValueSelect("tts_ai_device", []CustomWidget.TextValueOption{
		{Text: "CUDA", Value: "cuda"},
		{Text: "CPU", Value: "cpu"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)

	engine.Controls.TTSType = ttsType
	engine.Controls.TTSDevice = ttsDevice
	engine.Register("tts_type", ttsType)
	engine.Register("tts_ai_device", ttsDevice)
	return &TTSSection{TypeSelect: ttsType, DeviceSelect: ttsDevice}
}

// OCRSection encapsulates OCR widgets
type OCRSection struct {
	TypeSelect      *CustomWidget.TextValueSelect
	DeviceSelect    *CustomWidget.TextValueSelect
	PrecisionSelect *CustomWidget.TextValueSelect
}

// BuildOCRSection creates OCR selects and registers them
func (b *ProfileBuilder) BuildOCRSection(engine *FormEngine) *OCRSection {
	if engine == nil || engine.Controls == nil {
		return &OCRSection{}
	}
	ocrType := CustomWidget.NewTextValueSelect("ocr_type", []CustomWidget.TextValueOption{
		{Text: "Easy OCR", Value: "easyocr"},
		{Text: "GOT OCR 2.0", Value: "got_ocr_20"},
		{Text: "Phi-4", Value: "phi4"},
		{Text: lang.L("Disabled"), Value: ""},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	ocrDevice := CustomWidget.NewTextValueSelect("ocr_ai_device", []CustomWidget.TextValueOption{
		{Text: "CPU", Value: "cpu"},
		{Text: "CUDA", Value: "cuda"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)
	ocrPrecision := CustomWidget.NewTextValueSelect("ocr_precision", []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "bfloat16 " + lang.L("precision"), Value: "bfloat16"},
	}, func(_ CustomWidget.TextValueOption) {}, 0)

	engine.Controls.OCRType = ocrType
	engine.Controls.OCRDevice = ocrDevice
	engine.Controls.OCRPrecision = ocrPrecision
	engine.Register("ocr_type", ocrType)
	engine.Register("ocr_ai_device", ocrDevice)
	engine.Register("ocr_precision", ocrPrecision)
	return &OCRSection{TypeSelect: ocrType, DeviceSelect: ocrDevice, PrecisionSelect: ocrPrecision}
}
