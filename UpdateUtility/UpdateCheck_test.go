package UpdateUtility

import (
	"reflect"
	"testing"
	"whispering-tiger-ui/Updater"
)

func TestPackageDownloadURLsPrefersLocaleAndDeduplicates(t *testing.T) {
	packageInfo := Updater.UpdateInfo{
		LocationUrls: map[string][]string{
			"EU":      {"https://eu.example/platform.zip", "https://shared.example/platform.zip"},
			"US":      {"https://us.example/platform.zip", "https://shared.example/platform.zip"},
			"DEFAULT": {"https://default.example/platform.zip"},
			"BACKUP":  {"https://backup.example/platform.zip"},
		},
	}

	europeanURLs := packageDownloadURLs(packageInfo, "de_DE")
	expectedEuropeanURLs := []string{
		"https://eu.example/platform.zip",
		"https://shared.example/platform.zip",
		"https://default.example/platform.zip",
		"https://us.example/platform.zip",
		"https://backup.example/platform.zip",
	}
	if !reflect.DeepEqual(europeanURLs, expectedEuropeanURLs) {
		t.Fatalf("unexpected European mirror order: %v", europeanURLs)
	}

	usURLs := packageDownloadURLs(packageInfo, "en_US")
	expectedUSURLs := []string{
		"https://us.example/platform.zip",
		"https://shared.example/platform.zip",
		"https://default.example/platform.zip",
		"https://eu.example/platform.zip",
		"https://backup.example/platform.zip",
	}
	if !reflect.DeepEqual(usURLs, expectedUSURLs) {
		t.Fatalf("unexpected US mirror order: %v", usURLs)
	}
}

func TestPackageDownloadURLsHandlesEmptyManifest(t *testing.T) {
	if urls := packageDownloadURLs(Updater.UpdateInfo{}, "de_DE"); len(urls) != 0 {
		t.Fatalf("expected no mirrors, got %v", urls)
	}
}

func TestUpdateDownloadServerName(t *testing.T) {
	if got := downloadServerName("https://eu2.example.test/platform.zip"); got != "eu2" {
		t.Fatalf("unexpected server name: %q", got)
	}
}
