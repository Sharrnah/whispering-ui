package ProfileForm

import (
	"runtime"
	"strings"
	"testing"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/Hardwareinfo"
)

func audioCppOptionValues(options []TVO) map[string]bool {
	values := make(map[string]bool, len(options))
	for _, option := range options {
		values[option.Value] = true
	}
	return values
}

func TestAudioCppProfileOptions(t *testing.T) {
	sttFound := false
	for _, option := range STTTypeOptions() {
		if option.Value == "audio_cpp" {
			sttFound = true
			break
		}
	}
	if !sttFound {
		t.Fatal("audio.cpp is missing from STT type options")
	}

	ttsFound := false
	for _, option := range TTSTypeOptions() {
		if option.Value == "audio_cpp" {
			ttsFound = true
			break
		}
	}
	if !ttsFound {
		t.Fatal("audio.cpp is missing from TTS type options")
	}

	models, defaultIndex, enabled := STTModelOptions("audio_cpp")
	wantSTTModels := []string{
		"Qwen3-ASR-0.6B-GGUF",
		"Qwen3-ASR-1.7B-GGUF",
		"Nemotron-3.5-ASR-Streaming-0.6B-GGUF",
		"VibeVoice-ASR-GGUF",
		"Voxtral-Mini-4B-Realtime-2602-GGUF",
		"Audio8-ASR-0.1B-GGUF",
		"Kroko-ASR-English-64L-GGUF",
		"Canary-180M-Flash-GGUF",
		"Cohere-Transcribe-GGUF",
		"Moonshine-Streaming-Tiny-GGUF",
		"Moonshine-Streaming-Small-GGUF",
		"Moonshine-Streaming-Medium-GGUF",
		"Niagara-19M-Batch-English-GGUF",
		"Niagara-38M-Batch-English-GGUF",
		"MOSS-Transcribe-Diarize-GGUF",
		"VibeVoice-ASR-Streaming-7B-GGUF",
	}
	if !enabled || defaultIndex != 0 || len(models) != len(wantSTTModels) {
		t.Fatalf("unexpected audio.cpp models: enabled=%v default=%d models=%#v", enabled, defaultIndex, models)
	}
	for index, want := range wantSTTModels {
		if models[index].Value != want || models[index].Text == want {
			t.Fatalf("audio.cpp STT option %d = %#v, want value %q with a helpful description", index, models[index], want)
		}
	}

	sttPrecisions, enabled := STTPrecisionOptions("audio_cpp")
	if !enabled {
		t.Fatal("audio.cpp STT precision selection should be enabled")
	}
	values := audioCppOptionValues(sttPrecisions)
	if !values["q8_0"] || !values["f16"] || len(values) != 2 {
		t.Fatalf("unexpected audio.cpp STT precisions: %#v", values)
	}
	voxtralPrecisions, enabled := STTPrecisionOptionsForModel("audio_cpp", "Voxtral-Mini-4B-Realtime-2602-GGUF")
	values = audioCppOptionValues(voxtralPrecisions)
	if !enabled || !values["q4_k"] || !values["q8_0"] || !values["bf16"] || len(values) != 3 {
		t.Fatalf("unexpected Voxtral package precisions: enabled=%v values=%#v", enabled, values)
	}
	krokoPrecisions, enabled := STTPrecisionOptionsForModel("audio_cpp", "Kroko-ASR-English-64L-GGUF")
	if enabled || len(krokoPrecisions) != 1 || krokoPrecisions[0].Value != "q8_0" {
		t.Fatalf("Kroko must expose its sole Q8_0 package read-only: enabled=%v options=%#v", enabled, krokoPrecisions)
	}

	ttsPrecisions, enabled := TTSPrecisionOptions("audio_cpp")
	if !enabled {
		t.Fatal("audio.cpp TTS precision selection should be enabled")
	}
	values = audioCppOptionValues(ttsPrecisions)
	if !values["orig"] || !values["f16"] || len(values) != 2 {
		t.Fatalf("unexpected audio.cpp TTS precisions: %#v", values)
	}

	ttsModels, defaultIndex, enabled := TTSModelOptions("audio_cpp")
	wantTTSModels := []string{
		"Supertonic-3-GGUF",
		"Confucius4-TTS-GGUF",
		"DotTTS-SOAR-GGUF",
		"DotTTS-MeanFlow-GGUF",
		"DotTTS-Edit-GGUF",
		"IndexTTS2-GGUF",
		"IndexTTS2.5-GGUF",
		"MagpieTTS-Multilingual-357M-GGUF",
		"OmniVoice-GGUF",
		"VoxCPM1-0.5B-GGUF",
		"VoxCPM2-GGUF",
		"Breeze-TTS-2-GGUF",
		"Chatterbox-Turbo-GGUF",
		"CosyVoice3-GGUF",
		"Kokoro-82M-GGUF",
		"Audio8-TTS-Preview-0.6B-GGUF",
		"sanoTTS-heart-nano-GGUF",
		"sanoTTS-heart-GGUF",
		"sanoTTS-amy-GGUF",
		"sanoTTS-hfc-GGUF",
		"sanoTTS-kristin-GGUF",
		"sanoTTS-vi-GGUF",
		"sanoTTS-id-GGUF",
		"sanoTTS-cs-GGUF",
		"sanoTTS-de-GGUF",
		"sanoTTS-es-GGUF",
		"sanoTTS-fr-GGUF",
		"sanoTTS-it-GGUF",
		"sanoTTS-pt-GGUF",
		"sanoTTS-ro-GGUF",
		"sanoTTS-ru-GGUF",
		"sanoTTS-tr-GGUF",
		"sanoTTS-ne-GGUF",
		"sanoTTS-hi-GGUF",
	}
	if !enabled || defaultIndex != 0 || len(ttsModels) != len(wantTTSModels) {
		t.Fatalf("unexpected audio.cpp TTS models: enabled=%v default=%d models=%#v", enabled, defaultIndex, ttsModels)
	}
	for index, want := range wantTTSModels {
		if ttsModels[index].Value != want || ttsModels[index].Text == want {
			t.Fatalf("audio.cpp TTS option %d = %#v, want value %q with a helpful description", index, ttsModels[index], want)
		}
	}
	if group, ok := TTSModelProfileGroup("audio_cpp", ttsModels[0].Value); !ok || group != "Preset voices" {
		t.Fatalf("audio.cpp TTS model group = %q/%v, want Preset voices/true", group, ok)
	}
	confuciusPrecisions, enabled := TTSPrecisionOptionsForModel("audio_cpp", "Confucius4-TTS-GGUF")
	if enabled || len(confuciusPrecisions) != 1 || confuciusPrecisions[0].Value != "orig" {
		t.Fatalf("Confucius4 must expose its sole original package read-only: enabled=%v options=%#v", enabled, confuciusPrecisions)
	}
	indexPrecisions, enabled := TTSPrecisionOptionsForModel("audio_cpp", "IndexTTS2.5-GGUF")
	values = audioCppOptionValues(indexPrecisions)
	if !enabled || !values["q8_0"] || !values["f16"] || !values["orig"] || len(values) != 3 {
		t.Fatalf("unexpected IndexTTS2.5 package precisions: enabled=%v values=%#v", enabled, values)
	}
}

func TestAudioCppDeviceOptionsAreBackendSpecific(t *testing.T) {
	values := audioCppOptionValues(AudioCppDeviceOptions("audio_cpp"))
	if !values["cpu"] {
		t.Fatal("audio.cpp must support CPU")
	}
	if runtime.GOOS == "darwin" {
		if !values["metal"] || values["cuda"] || values["vulkan"] || values["hip"] {
			t.Fatalf("unexpected macOS audio.cpp devices: %#v", values)
		}
	} else if runtime.GOOS == "linux" {
		if !values["vulkan"] || values["cuda"] || values["hip"] || values["metal"] {
			t.Fatalf("unexpected managed Linux audio.cpp devices: %#v", values)
		}
	} else if !values["cuda"] || !values["vulkan"] || values["hip"] {
		t.Fatalf("unexpected visible audio.cpp GPU backends: %#v", values)
	}
	for _, option := range AudioCppDeviceOptions("audio_cpp") {
		if option.Value == "vulkan" && !strings.Contains(option.Text, "AMD") {
			t.Fatalf("Vulkan option must identify AMD support: %q", option.Text)
		}
	}

	ordinary := audioCppOptionValues(STTDeviceOptions("qwen3_asr"))
	if ordinary["vulkan"] || ordinary["hip"] || ordinary["metal"] {
		t.Fatalf("native audio.cpp devices leaked into a PyTorch backend: %#v", ordinary)
	}
}

func TestAudioCppVulkanUsesTheGPUIndexSelector(t *testing.T) {
	device := CustomWidget.NewTextValueSelect("device", AudioCppDeviceOptions("audio_cpp"), nil, 0)
	device.SetSelected("vulkan")
	gpu := CustomWidget.NewTextValueSelect("gpu", DefaultGPUOptions(), nil, 1)
	coordinator := &Coordinator{
		GPUOptions: []TVO{{Text: "CUDA adapter name", Value: "0"}},
		NativeGPUOptions: AudioCppGPUOptions([]Hardwareinfo.AudioCppDeviceInfo{
			{Backend: "vulkan", Index: 0, Name: "NVIDIA GeForce RTX 5090", Kind: "GPU"},
			{Backend: "vulkan", Index: 1, Name: "Intel Integrated Graphics", Kind: "IGPU"},
		}, nil),
	}

	coordinator.updateGPUSelectorState(device, gpu)

	if gpu.Disabled() {
		t.Fatal("GPU selector should be enabled for Vulkan")
	}
	if len(gpu.Options) != 2 || gpu.Options[0].Value != "0" || gpu.Options[1].Value != "1" {
		t.Fatalf("Vulkan should use reported adapter indices, got %#v", gpu.Options)
	}
	if gpu.Options[0].Text != "GPU 0 - NVIDIA GeForce RTX 5090 [GPU]" || gpu.Options[1].Text != "GPU 1 - Intel Integrated Graphics [IGPU]" {
		t.Fatalf("Vulkan adapter names are missing: %#v", gpu.Options)
	}
	if selectedValue(gpu) != "1" {
		t.Fatalf("Vulkan adapter index = %q, want 1", selectedValue(gpu))
	}
}
