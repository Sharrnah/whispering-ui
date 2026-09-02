package ProfileForm

import (
	"testing"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Settings"
)

func optionValues(options []TVO) map[string]bool {
	values := make(map[string]bool, len(options))
	for _, option := range options {
		values[option.Value] = true
	}
	return values
}

func TestTTSPrecisionOptionsMatchBackendCapabilities(t *testing.T) {
	tests := []struct {
		model   string
		values  []string
		enabled bool
	}{
		{model: "silero", values: []string{"float32"}, enabled: false},
		{model: "orpheus", values: []string{"8bit"}, enabled: false},
		{model: "chatterbox", values: []string{"float32", "float16"}, enabled: true},
		{model: "index_tts", values: []string{"bfloat16", "float32"}, enabled: true},
		{model: "qwen3_tts", values: []string{"auto", "bfloat16", "float16", "float32"}, enabled: true},
		{model: "audio8_tts", values: []string{"auto", "bfloat16", "float32"}, enabled: true},
	}
	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			options, enabled := TTSPrecisionOptions(test.model)
			if enabled != test.enabled {
				t.Fatalf("enabled = %v, want %v", enabled, test.enabled)
			}
			values := optionValues(options)
			if len(values) != len(test.values) {
				t.Fatalf("values = %#v, want %#v", values, test.values)
			}
			for _, value := range test.values {
				if !values[value] {
					t.Fatalf("missing precision %q in %#v", value, values)
				}
			}
		})
	}
}

func TestGPUIndexAndTTSPrecisionRoundTrip(t *testing.T) {
	controls := &AllProfileControls{}
	engine := NewFormEngine(controls, nil)
	controls.STTGPU = CustomWidget.NewTextValueSelect("ai_device_index", DefaultGPUOptions(), nil, 0)
	controls.TxtGPU = CustomWidget.NewTextValueSelect("txt_translator_device_index", DefaultGPUOptions(), nil, 0)
	controls.TTSGPU = CustomWidget.NewTextValueSelect("tts_ai_device_index", DefaultGPUOptions(), nil, 0)
	controls.OCRGPU = CustomWidget.NewTextValueSelect("ocr_ai_device_index", DefaultGPUOptions(), nil, 0)
	controls.TTSPrecision = CustomWidget.NewTextValueSelect("tts_precision", GenericTTSPrecisionOptions(), nil, 0)
	engine.Register("ai_device_index", controls.STTGPU)
	engine.Register("txt_translator_device_index", controls.TxtGPU)
	engine.Register("tts_ai_device_index", controls.TTSGPU)
	engine.Register("ocr_ai_device_index", controls.OCRGPU)
	engine.Register("tts_precision", controls.TTSPrecision)

	conf := Settings.Conf{
		Ai_device_index: 2, Txt_translator_device_index: 3,
		Tts_ai_device_index: 4, Ocr_ai_device_index: 5,
		Tts_precision: "bfloat16",
	}
	engine.LoadFromSettings(&conf)
	if selectedValue(controls.STTGPU) != "2" || selectedValue(controls.TxtGPU) != "3" || selectedValue(controls.TTSGPU) != "4" || selectedValue(controls.OCRGPU) != "5" {
		t.Fatalf("GPU selections were not loaded: STT=%q TXT=%q TTS=%q OCR=%q", selectedValue(controls.STTGPU), selectedValue(controls.TxtGPU), selectedValue(controls.TTSGPU), selectedValue(controls.OCRGPU))
	}
	if selectedValue(controls.TTSPrecision) != "bfloat16" {
		t.Fatalf("TTS precision = %q, want bfloat16", selectedValue(controls.TTSPrecision))
	}

	controls.STTGPU.SetSelected("6")
	controls.TTSPrecision.SetSelected("float32")
	engine.SaveToSettings(&conf)
	if conf.Ai_device_index != 6 || conf.Tts_precision != "float32" {
		t.Fatalf("saved GPU/precision = %d/%q, want 6/float32", conf.Ai_device_index, conf.Tts_precision)
	}
}

func TestGPUSelectorOnlyEnablesForCUDA(t *testing.T) {
	device := CustomWidget.NewTextValueSelect("device", DefaultDeviceOptions(), nil, 0)
	gpu := CustomWidget.NewTextValueSelect("gpu", DefaultGPUOptions(), nil, 0)
	coordinator := &Coordinator{}
	coordinator.updateGPUSelectorState(device, gpu)
	if !gpu.Disabled() {
		t.Fatal("GPU selector should be disabled for CPU")
	}
	device.SetSelected("cuda")
	coordinator.updateGPUSelectorState(device, gpu)
	if gpu.Disabled() {
		t.Fatal("GPU selector should be enabled for CUDA")
	}
}

func TestMultiModalSyncCopiesGPUIndex(t *testing.T) {
	devices := DefaultDeviceOptions()
	gpus := DefaultGPUOptions()
	controls := &AllProfileControls{
		STTDevice:    CustomWidget.NewTextValueSelect("stt_device", devices, nil, 1),
		STTGPU:       CustomWidget.NewTextValueSelect("stt_gpu", gpus, nil, 3),
		TxtDevice:    CustomWidget.NewTextValueSelect("txt_device", devices, nil, 0),
		TxtGPU:       CustomWidget.NewTextValueSelect("txt_gpu", gpus, nil, 0),
		TxtSize:      CustomWidget.NewTextValueSelect("txt_size", []TVO{{Text: "Small", Value: "small"}}, nil, 0),
		TxtPrecision: CustomWidget.NewTextValueSelect("txt_precision", []TVO{{Text: "float32", Value: "float32"}}, nil, 0),
	}
	coordinator := &Coordinator{Controls: controls}
	coordinator.mirrorFromTo(groupSTT, groupTXT)
	if selectedValue(controls.TxtDevice) != "cuda" || selectedValue(controls.TxtGPU) != "3" {
		t.Fatalf("mirrored device = %q GPU %q, want cuda GPU 3", selectedValue(controls.TxtDevice), selectedValue(controls.TxtGPU))
	}
}
