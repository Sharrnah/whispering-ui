package Hardwareinfo

import (
	"os"
	"testing"
)

func TestParseAudioCppDevices(t *testing.T) {
	output := `ggml_vulkan: Found 2 Vulkan devices:
available_devices=3
Vulkan:0 "NVIDIA GeForce RTX 5090" [GPU]
Vulkan:1 "Intel(R) RaptorLake-S Mobile Graphics Controller" [IGPU]
CPU:0 "13th Gen Intel(R) Core(TM) i9-13900K" [CPU]
select with: --backend vulkan --device 0`
	devices := ParseAudioCppDevices(output)
	if len(devices) != 3 {
		t.Fatalf("parsed %d devices, want 3: %#v", len(devices), devices)
	}
	if devices[0].Backend != "vulkan" || devices[0].Index != 0 || devices[0].Name != "NVIDIA GeForce RTX 5090" || devices[0].Kind != "GPU" {
		t.Fatalf("unexpected first device: %#v", devices[0])
	}
	if devices[1].Backend != "vulkan" || devices[1].Index != 1 || devices[1].Kind != "IGPU" {
		t.Fatalf("unexpected second device: %#v", devices[1])
	}
	if devices[2].Backend != "cpu" || devices[2].Index != 0 {
		t.Fatalf("unexpected CPU device: %#v", devices[2])
	}
}

func TestParseAudioCppDevicesNormalizesROCm(t *testing.T) {
	devices := ParseAudioCppDevices(`ROCm:2 "AMD Radeon RX 7900 XTX" [GPU]`)
	if len(devices) != 1 || devices[0].Backend != "hip" || devices[0].Index != 2 {
		t.Fatalf("unexpected devices: %#v", devices)
	}
}

func TestGetAudioCppDevicesFromConfiguredServer(t *testing.T) {
	server := os.Getenv("AUDIOCPP_TEST_SERVER")
	if server == "" {
		t.Skip("set AUDIOCPP_TEST_SERVER for an installed-runtime integration test")
	}
	t.Setenv("WHISPERING_TIGER_AUDIOCPP_SERVER", server)
	t.Setenv("AUDIOCPP_SERVER_PATH", "")
	devices, err := GetAudioCppDevices()
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) == 0 {
		t.Fatal("configured audio.cpp server returned no devices")
	}
}
