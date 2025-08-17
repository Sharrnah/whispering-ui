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

// common option lists (wrapped in funcs so they are built at call-time)
func deviceOptionsCUDAFirst() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{{Text: "CUDA", Value: "cuda"}, {Text: "CPU", Value: "cpu"}}
}

func deviceOptionsCPUFirst() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{{Text: "CPU", Value: "cpu"}, {Text: "CUDA", Value: "cuda"}}
}

func precisionWhisperOptions() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "int16 " + lang.L("precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
		{Text: "8bit " + lang.L("precision"), Value: "8bit"},
		{Text: "4bit " + lang.L("precision"), Value: "4bit"},
	}
}

func precisionTextOptions() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "int16 " + lang.L("precision"), Value: "int16"},
		{Text: "int8_float16 " + lang.L("precision"), Value: "int8_float16"},
		{Text: "int8 " + lang.L("precision"), Value: "int8"},
		{Text: "bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "bfloat16"},
		{Text: "int8_bfloat16 " + lang.L("precision") + " (Compute >=8.0)", Value: "int8_bfloat16"},
	}
}

func precisionOCROptions() []CustomWidget.TextValueOption {
	return []CustomWidget.TextValueOption{
		{Text: "float32 " + lang.L("precision"), Value: "float32"},
		{Text: "float16 " + lang.L("precision"), Value: "float16"},
		{Text: "bfloat16 " + lang.L("precision"), Value: "bfloat16"},
	}
}

// BuildConnectionSection builds the Websocket IP/Port + Run Backend row and registers controls
func (b *ProfileBuilder) BuildConnectionSection(engine *FormEngine) fyne.CanvasObject {
	if engine == nil || engine.Controls == nil {
		return container.NewHBox()
	}
	wsIP := b.newEntry(engine, "websocket_ip", "127.0.0.1")
	wsPort := b.newEntry(engine, "websocket_port", "5000")
	runBackend := b.newCheck(engine, "run_backend", lang.L("Run Backend"))
	engine.Controls.WebsocketIP, engine.Controls.WebsocketPort, engine.Controls.RunBackend = wsIP, wsPort, runBackend
	return container.NewGridWithColumns(3, wsIP, wsPort, runBackend)
}

// AudioSection encapsulates the Audio UI widgets created by the builder
type AudioSection struct {
	ApiSelect    *CustomWidget.TextValueSelect
	InputSelect  *CustomWidget.TextValueSelect
	OutputSelect *CustomWidget.TextValueSelect
}

// BuildAudioSection creates Audio API / Input / Output selects and registers them
func (b *ProfileBuilder) BuildAudioSection(engine *FormEngine, inputOptions, outputOptions []CustomWidget.TextValueOption) *AudioSection {
	if engine == nil || engine.Controls == nil {
		return &AudioSection{}
	}
	audioOptions := make([]CustomWidget.TextValueOption, 0, len(AudioAPI.AudioBackends))
	for _, backend := range AudioAPI.AudioBackends {
		audioOptions = append(audioOptions, CustomWidget.TextValueOption{Text: backend.Name, Value: backend.Name})
	}
	audioApiSelect := b.newSelect(engine, "audio_api", audioOptions)
	audioInputSelect := b.newSelect(engine, "device_index", inputOptions)
	audioOutputSelect := b.newSelect(engine, "device_out_index", outputOptions)
	engine.Controls.AudioAPI, engine.Controls.AudioInput, engine.Controls.AudioOutput = audioApiSelect, audioInputSelect, audioOutputSelect
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
	vadEnable := b.newCheck(engine, "vad_enabled", lang.L("Enable"))
	vadOnFullClip := b.newCheck(engine, "vad_on_full_clip", "+ Check on Full Clip")
	vadOnFullClip.Hide()
	vadRealtime := b.newCheck(engine, "realtime", lang.L("Realtime"))
	pushToTalk := CustomWidget.NewHotKeyEntry()
	pushToTalk.PlaceHolder = lang.L("Keypress")
	pushToTalkBlock := container.NewBorder(nil, nil, container.NewHBox(widget.NewLabel(lang.L("Push to Talk")), widget.NewIcon(theme.ComputerIcon())), nil, pushToTalk)
	groupRow := container.NewGridWithColumns(3, vadEnable, vadOnFullClip, vadRealtime, pushToTalkBlock)
	confSlider, confLabel, confRow := b.sliderWithLabel(engine, "vad_confidence_threshold", 0, 1, 0.01)
	engine.Controls.VadEnable, engine.Controls.VadOnFullClip, engine.Controls.VadRealtime = vadEnable, vadOnFullClip, vadRealtime
	engine.Controls.PushToTalk, engine.Controls.VadConfidence = pushToTalk, confSlider
	engine.Register("push_to_talk_key", pushToTalk)
	return &VADSection{EnableCheck: vadEnable, OnFullClipCheck: vadOnFullClip, RealtimeCheck: vadRealtime, PushToTalk: pushToTalk, PushToTalkBlock: pushToTalkBlock, ConfidenceLabel: confLabel, ConfidenceSlider: confSlider, ConfidenceRow: confRow, GroupRow: groupRow}
}

// STTSection encapsulates STT widgets
type STTSection struct {
	TypeSelect      *CustomWidget.TextValueSelect
	DeviceSelect    *CustomWidget.TextValueSelect
	PrecisionSelect *CustomWidget.TextValueSelect
	SizeSelect      *CustomWidget.TextValueSelect
}

func (b *ProfileBuilder) BuildSTTSection(engine *FormEngine) *STTSection {
	if engine == nil || engine.Controls == nil {
		return &STTSection{}
	}
	sttDevice := b.newSelect(engine, "ai_device", deviceOptionsCUDAFirst())
	sttPrecision := b.newSelect(engine, "Precision", precisionWhisperOptions())
	sttType := b.newSelect(engine, "stt_type", []CustomWidget.TextValueOption{{Text: "Faster Whisper", Value: "faster_whisper"}, {Text: "Original Whisper", Value: "original_whisper"}, {Text: "Transformer Whisper", Value: "transformer_whisper"}, {Text: "Seamless M4T", Value: "seamless_m4t"}, {Text: "MMS", Value: "mms"}, {Text: "Speech T5 (English only)", Value: "speech_t5"}, {Text: "Wav2Vec Bert 2.0", Value: "wav2vec_bert"}, {Text: "NeMo Canary", Value: "nemo_canary"}, {Text: "Phi-4", Value: "phi4"}, {Text: "Voxtral", Value: "voxtral"}, {Text: lang.L("Disabled"), Value: ""}})
	fasterWhisperModelList := []CustomWidget.TextValueOption{{Text: "Tiny", Value: "tiny"}, {Text: "Tiny (English only)", Value: "tiny.en"}, {Text: "Base", Value: "base"}, {Text: "Base (English only)", Value: "base.en"}, {Text: "Small", Value: "small"}, {Text: "Small (English only)", Value: "small.en"}, {Text: "Medium", Value: "medium"}, {Text: "Medium (English only)", Value: "medium.en"}, {Text: "Large V1", Value: "large-v1"}, {Text: "Large V2", Value: "large-v2"}, {Text: "Large V3", Value: "large-v3"}, {Text: "Large V3 Turbo", Value: "large-v3-turbo"}, {Text: "Medium Distilled (English)", Value: "medium-distilled.en"}, {Text: "Large V2 Distilled (English)", Value: "large-distilled-v2.en"}, {Text: "Large V3 Distilled (English)", Value: "large-distilled-v3.en"}, {Text: "Large V3.5 Distilled (English)", Value: "large-distilled-v3.5.en"}, {Text: "Crisper", Value: "crisper"}, {Text: "Small (European finetune)", Value: "small.eu"}, {Text: "Medium (European finetune)", Value: "medium.eu"}, {Text: "Small (German finetune)", Value: "small.de"}, {Text: "Medium (German finetune)", Value: "medium.de"}, {Text: "Large V2 (German finetune)", Value: "large-v2.de2"}, {Text: "Large V3 Distilled (German finetune)", Value: "large-distilled-v3.de"}, {Text: "Small (German-Swiss finetune)", Value: "small.de-swiss"}, {Text: "Medium (Mix-Japanese-v2 finetune)", Value: "medium.mix-jpv2"}, {Text: "Large V2 (Mix-Japanese finetune)", Value: "large-v2.mix-jp"}, {Text: "Small (Japanese finetune)", Value: "small.jp"}, {Text: "Medium (Japanese finetune)", Value: "medium.jp"}, {Text: "Large V2 (Japanese finetune)", Value: "large-v2.jp"}, {Text: "Medium (Korean finetune)", Value: "medium.ko"}, {Text: "Large V2 (Korean finetune)", Value: "large-v2.ko"}, {Text: "Small (Chinese finetune)", Value: "small.zh"}, {Text: "Medium (Chinese finetune)", Value: "medium.zh"}, {Text: "Large V2 (Chinese finetune)", Value: "large-v2.zh"}, {Text: "Custom (Place in '.cache/whisper/custom-ct2' directory)", Value: "custom"}}
	sttModel := b.newSelect(engine, "model", fasterWhisperModelList)
	engine.Controls.STTDevice, engine.Controls.STTPrecision, engine.Controls.STTType, engine.Controls.STTModelSize = sttDevice, sttPrecision, sttType, sttModel
	engine.Register("whisper_precision", sttPrecision)
	return &STTSection{TypeSelect: sttType, DeviceSelect: sttDevice, PrecisionSelect: sttPrecision, SizeSelect: sttModel}
}

// TXTSection encapsulates Text-Translator widgets
type TXTSection struct {
	TypeSelect      *CustomWidget.TextValueSelect
	DeviceSelect    *CustomWidget.TextValueSelect
	PrecisionSelect *CustomWidget.TextValueSelect
	SizeSelect      *CustomWidget.TextValueSelect
}

func (b *ProfileBuilder) BuildTXTSection(engine *FormEngine) *TXTSection {
	if engine == nil || engine.Controls == nil {
		return &TXTSection{}
	}
	txtType := b.newSelect(engine, "txt_translator", []CustomWidget.TextValueOption{{Text: "Faster NLLB200 (200 languages)", Value: "NLLB200_CT2"}, {Text: "Original NLLB200 (200 languages)", Value: "NLLB200"}, {Text: "M2M100 (100 languages)", Value: "M2M100"}, {Text: "Seamless M4T (101 languages)", Value: "seamless_m4t"}, {Text: "Phi-4 (23 languages)", Value: "phi4"}, {Text: "Voxtral (13 languages)", Value: "voxtral"}, {Text: lang.L("Disabled"), Value: ""}})
	txtDevice := b.newSelect(engine, "txt_translator_device", deviceOptionsCUDAFirst())
	txtSize := b.newSelect(engine, "txt_translator_size", []CustomWidget.TextValueOption{{Text: "Small", Value: "small"}, {Text: "Medium", Value: "medium"}, {Text: "Large", Value: "large"}})
	txtPrecision := b.newSelect(engine, "txt_translator_precision", precisionTextOptions())
	engine.Controls.TxtType, engine.Controls.TxtDevice, engine.Controls.TxtSize, engine.Controls.TxtPrecision = txtType, txtDevice, txtSize, txtPrecision
	return &TXTSection{TypeSelect: txtType, DeviceSelect: txtDevice, PrecisionSelect: txtPrecision, SizeSelect: txtSize}
}

// TTSSection encapsulates TTS widgets
type TTSSection struct{ TypeSelect, DeviceSelect *CustomWidget.TextValueSelect }

func (b *ProfileBuilder) BuildTTSSection(engine *FormEngine) *TTSSection {
	if engine == nil || engine.Controls == nil {
		return &TTSSection{}
	}
	ttsType := b.newSelect(engine, "tts_type", []CustomWidget.TextValueOption{{Text: "Silero", Value: "silero"}, {Text: "F5/E2", Value: "f5_e2"}, {Text: "Zonos", Value: "zonos"}, {Text: "Kokoro", Value: "kokoro"}, {Text: "Orpheus", Value: "orpheus"}, {Text: lang.L("Disabled"), Value: ""}})
	ttsDevice := b.newSelect(engine, "tts_ai_device", deviceOptionsCUDAFirst())
	engine.Controls.TTSType, engine.Controls.TTSDevice = ttsType, ttsDevice
	return &TTSSection{TypeSelect: ttsType, DeviceSelect: ttsDevice}
}

// OCRSection encapsulates OCR widgets
type OCRSection struct{ TypeSelect, DeviceSelect, PrecisionSelect *CustomWidget.TextValueSelect }

func (b *ProfileBuilder) BuildOCRSection(engine *FormEngine) *OCRSection {
	if engine == nil || engine.Controls == nil {
		return &OCRSection{}
	}
	ocrType := b.newSelect(engine, "ocr_type", []CustomWidget.TextValueOption{{Text: "Easy OCR", Value: "easyocr"}, {Text: "GOT OCR 2.0", Value: "got_ocr_20"}, {Text: "Phi-4", Value: "phi4"}, {Text: lang.L("Disabled"), Value: ""}})
	ocrDevice := b.newSelect(engine, "ocr_ai_device", deviceOptionsCPUFirst())
	ocrPrecision := b.newSelect(engine, "ocr_precision", precisionOCROptions())
	engine.Controls.OCRType, engine.Controls.OCRDevice, engine.Controls.OCRPrecision = ocrType, ocrDevice, ocrPrecision
	return &OCRSection{TypeSelect: ocrType, DeviceSelect: ocrDevice, PrecisionSelect: ocrPrecision}
}
