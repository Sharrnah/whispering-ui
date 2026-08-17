package Messages

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishDownloadFailureWritesAtomicReceipt(t *testing.T) {
	failedPath := filepath.Join(t.TempDir(), "model.zip.failed")
	downloadErr := errors.New("checksum mismatch")

	returnedErr := publishDownloadFailure(failedPath, downloadErr)
	if !errors.Is(returnedErr, downloadErr) {
		t.Fatalf("expected original download error, got %v", returnedErr)
	}

	receipt, err := os.ReadFile(failedPath)
	if err != nil {
		t.Fatalf("read failure receipt: %v", err)
	}
	if string(receipt) != downloadErr.Error() {
		t.Fatalf("unexpected failure receipt: %q", receipt)
	}
	if _, err := os.Stat(failedPath + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temporary receipt remains after publish: %v", err)
	}
}

func TestStartDownloadPublishesFailureReceipt(t *testing.T) {
	extractPath := filepath.Join(t.TempDir(), "model.zip")
	message := DownloadMessage{
		Download: Download{
			ExtractDir: extractPath,
			Title:      "test model",
		},
	}

	err := message.StartDownload()
	if err == nil {
		t.Fatal("expected download without URLs to fail")
	}
	receipt, readErr := os.ReadFile(extractPath + ".failed")
	if readErr != nil {
		t.Fatalf("read failure receipt: %v", readErr)
	}
	if string(receipt) != err.Error() {
		t.Fatalf("failure receipt %q does not match error %q", receipt, err)
	}
}
