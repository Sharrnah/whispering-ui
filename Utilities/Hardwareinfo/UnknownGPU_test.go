package Hardwareinfo

import (
	"github.com/jaypipes/ghw/pkg/pci"
	"testing"
)

func TestUnknownGPUVendorDoesNotPanic(t *testing.T) {
	if IsNVIDIACard(&pci.Device{}) {
		t.Fatal("unknown PCI device identified as NVIDIA")
	}
}
