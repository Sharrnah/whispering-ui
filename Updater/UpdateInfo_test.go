package Updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUpdateInfoParsesPackages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/yaml")
		_, _ = writer.Write([]byte("packages:\n  ai_platform:\n    version: 1.2.3\n"))
	}))
	defer server.Close()

	packages := UpdatePackages{}
	if err := packages.GetUpdateInfo(server.URL); err != nil {
		t.Fatalf("get update info: %v", err)
	}
	if got := packages.Packages["ai_platform"].Version; got != "1.2.3" {
		t.Fatalf("unexpected parsed version: %q", got)
	}
}

func TestGetUpdateInfoHonorsHTTPTimeout(t *testing.T) {
	originalClient := updateInfoHTTPClient
	updateInfoHTTPClient = &http.Client{Timeout: 40 * time.Millisecond}
	t.Cleanup(func() { updateInfoHTTPClient = originalClient })

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()

	started := time.Now()
	err := (&UpdatePackages{}).GetUpdateInfo(server.URL)
	if err == nil {
		t.Fatal("expected the blocked update request to time out")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("update request timeout took too long: %s", elapsed)
	}
}
