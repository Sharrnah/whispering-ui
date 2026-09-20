package Updater

import "runtime"

// Set at build time with -X whispering-tiger-ui/Updater.LinuxBackendFlavor=cu128.
// A missing Linux entry must never fall back to the legacy Windows archive.
var LinuxBackendFlavor = "cu128"

// Private preview bundles are tested before their package is published.
// Set with -X for preview builds only; normal releases retain update checks.
var LinuxPreview = "false"

func UpdatesEnabled() bool {
	return updatesEnabled(runtime.GOOS, LinuxPreview)
}

func PlatformPackageName() string {
	return platformPackageName(runtime.GOOS, runtime.GOARCH, LinuxBackendFlavor)
}

func updatesEnabled(goos, preview string) bool {
	return goos != "linux" || preview != "true"
}

func platformPackageName(goos, goarch, flavor string) string {
	if goos == "windows" {
		return "ai_platform"
	}
	return "ai_platform_" + goos + "_" + goarch + "_" + flavor
}
