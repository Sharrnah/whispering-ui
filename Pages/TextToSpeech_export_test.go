package Pages

import (
	"errors"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

type ttsExportWriter struct {
	uri      fyne.URI
	closed   bool
	closeErr error
}

type ttsFileDialogSpy struct {
	calls []string
	size  fyne.Size
}

func (d *ttsFileDialogSpy) Show() {
	d.calls = append(d.calls, "show")
}

func (d *ttsFileDialogSpy) Resize(size fyne.Size) {
	d.calls = append(d.calls, "resize")
	d.size = size
}

func (w *ttsExportWriter) URI() fyne.URI               { return w.uri }
func (w *ttsExportWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *ttsExportWriter) Close() error {
	w.closed = true
	return w.closeErr
}

func TestTTSExportClosesDialogWriterBeforeBackendHandoff(t *testing.T) {
	expectedPath := filepath.Clean(filepath.Join(t.TempDir(), "index-tts.wav"))
	writer := &ttsExportWriter{uri: storage.NewFileURI(expectedPath)}
	callbackCalled := false

	path, err := handoffTTSExportWriter(writer, func(path string) {
		callbackCalled = true
		if !writer.closed {
			t.Fatal("backend callback ran before the Fyne writer was closed")
		}
		if filepath.Clean(path) != expectedPath {
			t.Fatalf("backend path = %q, want %q", path, expectedPath)
		}
	})

	if err != nil {
		t.Fatal(err)
	}
	if !callbackCalled {
		t.Fatal("backend callback was not called")
	}
	if filepath.Clean(path) != expectedPath {
		t.Fatalf("returned path = %q, want %q", path, expectedPath)
	}
}

func TestTTSExportDoesNotCallBackendWhenCloseFails(t *testing.T) {
	writer := &ttsExportWriter{
		uri:      storage.NewFileURI(filepath.Join(t.TempDir(), "index-tts.wav")),
		closeErr: errors.New("close failed"),
	}
	callbackCalled := false

	_, err := handoffTTSExportWriter(writer, func(string) { callbackCalled = true })

	if err == nil {
		t.Fatal("expected close error")
	}
	if callbackCalled {
		t.Fatal("backend callback ran after the writer failed to close")
	}
}

func TestTTSExportShowsFileDialogBeforeResize(t *testing.T) {
	dialog := &ttsFileDialogSpy{}
	wantSize := fyne.NewSize(800, 500)

	showSizedTTSFileDialog(dialog, wantSize)

	if len(dialog.calls) != 2 || dialog.calls[0] != "show" || dialog.calls[1] != "resize" {
		t.Fatalf("dialog calls = %v, want [show resize]", dialog.calls)
	}
	if dialog.size != wantSize {
		t.Fatalf("dialog size = %v, want %v", dialog.size, wantSize)
	}
}
