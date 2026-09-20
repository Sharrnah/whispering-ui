package Updater

import (
	"runtime"
	"testing"
)

func TestWindowsUpdateContract(t *testing.T) {
	for _, flavor := range []string{"cpu", "cu128"} {
		if got := platformPackageName("windows", "amd64", flavor); got != "ai_platform" {
			t.Fatalf("Windows package changed to %q", got)
		}
	}
	for _, preview := range []string{"true", "false"} {
		if !updatesEnabled("windows", preview) {
			t.Fatal("Linux preview flag disabled Windows updates")
		}
	}
}

func TestLinuxPreviewUpdateControl(t *testing.T) {
	if updatesEnabled("linux", "true") || !updatesEnabled("linux", "false") {
		t.Fatal("Linux update checks do not follow preview mode")
	}
}

func TestPlatformPackageNeverUsesWindowsArchiveOnLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux package selection")
	}
	previous := LinuxBackendFlavor
	t.Cleanup(func() { LinuxBackendFlavor = previous })
	for _, flavor := range []string{"cpu", "cu128"} {
		LinuxBackendFlavor = flavor
		want := "ai_platform_linux_" + runtime.GOARCH + "_" + flavor
		if got := PlatformPackageName(); got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	}
}
