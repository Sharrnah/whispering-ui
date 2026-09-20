package Hardwareinfo

import "testing"

func TestParseCUDADevices(t *testing.T) {
	devices, err := parseCUDADevices("1, NVIDIA RTX B, 24564, 8.9\n0, NVIDIA RTX A, 32768, 12.0\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 2 {
		t.Fatalf("device count = %d, want 2", len(devices))
	}
	if devices[0].Index != 0 || devices[0].Name != "NVIDIA RTX A" || devices[0].MemoryTotalMiB != 32768 || devices[0].ComputeCapability != 12.0 {
		t.Fatalf("unexpected first device: %#v", devices[0])
	}
	if devices[1].Index != 1 {
		t.Fatalf("second device index = %d, want 1", devices[1].Index)
	}
}

func TestParseCUDADevicesRejectsMalformedRows(t *testing.T) {
	if _, err := parseCUDADevices("0, Missing fields\n"); err == nil {
		t.Fatal("malformed nvidia-smi row was accepted")
	}
}
