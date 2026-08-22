package ModelDownloader

import (
	"fyne.io/fyne/v2/lang"
	"testing"
	"whispering-tiger-ui/Resources"
)

func TestRandomizedUniqueURLsDeduplicatesServers(t *testing.T) {
	urls := []string{
		"https://eu.example.test/model.zip",
		"https://us.example.test/model.zip",
		"https://eu.example.test/model.zip",
		"",
		"https://backup.example.test/model.zip",
	}

	randomized := randomizedUniqueURLs(urls)
	if len(randomized) != 3 {
		t.Fatalf("expected three unique non-empty URLs, got %v", randomized)
	}

	seen := make(map[string]bool, len(randomized))
	for _, downloadURL := range randomized {
		if seen[downloadURL] {
			t.Fatalf("URL was scheduled more than once: %s", downloadURL)
		}
		seen[downloadURL] = true
	}
}

func TestRotateDownloadURLsSelectsStartingServer(t *testing.T) {
	urls := []string{"eu", "us", "backup"}
	rotated := rotateDownloadURLs(urls, 2)
	expected := []string{"backup", "eu", "us"}
	for index := range expected {
		if rotated[index] != expected[index] {
			t.Fatalf("unexpected rotated server order: %v", rotated)
		}
	}
}

func TestNextDownloadServerIndexWrapsAround(t *testing.T) {
	if got := nextDownloadServerIndex(0, 3); got != 1 {
		t.Fatalf("expected the next server after 0 to be 1, got %d", got)
	}
	if got := nextDownloadServerIndex(2, 3); got != 0 {
		t.Fatalf("expected the final server to wrap to 0, got %d", got)
	}
	if got := nextDownloadServerIndex(0, 1); got != 0 {
		t.Fatalf("a one-server list must remain at 0, got %d", got)
	}
}

func TestDownloadURLDisplayHelpers(t *testing.T) {
	downloadURL := "https://eu2.example.test/models/model.zip?version=2"
	if got := downloadURLFilename(downloadURL); got != "model.zip" {
		t.Fatalf("unexpected filename: %q", got)
	}
	if got := downloadServerName(downloadURL); got != "eu2" {
		t.Fatalf("unexpected server name: %q", got)
	}
}

func TestChangeServerLabelIsLocalized(t *testing.T) {
	if err := lang.AddTranslationsFS(Resources.Translations, "translations"); err != nil {
		t.Fatalf("load translations: %v", err)
	}
	defer lang.SetPreferredLocale("en")

	lang.SetPreferredLocale("en")
	if got := lang.L("Change server"); got != "Change server" {
		t.Fatalf("unexpected English label: %q", got)
	}

	lang.SetPreferredLocale("de")
	if got := lang.L("Change server"); got != "Server wechseln" {
		t.Fatalf("unexpected German label: %q", got)
	}
}
