package Updater

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDownloadSessionDeduplicatesURLs(t *testing.T) {
	session := NewDownloadSession([]string{"eu", "", "us", "eu"}, "target", 1, 1)
	if len(session.URLs) != 2 || session.URLs[0] != "eu" || session.URLs[1] != "us" {
		t.Fatalf("unexpected URL list: %v", session.URLs)
	}
}

func TestDownloadSessionPauseAndResume(t *testing.T) {
	targetPath := filepath.Join(t.TempDir(), "platform.zip")
	secondChunkStarted := make(chan struct{})
	paused := make(chan struct{})
	var secondChunkOnce sync.Once
	var secondChunkRequests int
	var requestMu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Length", "64")
			writer.WriteHeader(http.StatusOK)
			return
		}

		switch request.Header.Get("Range") {
		case "bytes=0-31":
			writer.WriteHeader(http.StatusPartialContent)
			_, _ = writer.Write(bytes.Repeat([]byte{'a'}, 32))
		case "bytes=32-63":
			requestMu.Lock()
			secondChunkRequests++
			attempt := secondChunkRequests
			requestMu.Unlock()
			if attempt == 1 {
				secondChunkOnce.Do(func() { close(secondChunkStarted) })
				<-request.Context().Done()
				return
			}
			writer.WriteHeader(http.StatusPartialContent)
			_, _ = writer.Write(bytes.Repeat([]byte{'b'}, 32))
		default:
			t.Errorf("unexpected range: %q", request.Header.Get("Range"))
			writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		}
	}))
	defer server.Close()

	session := NewDownloadSession([]string{server.URL + "/platform.zip"}, targetPath, 1, 32)
	session.OnPaused = func() { close(paused) }
	result := make(chan error, 1)
	go func() { result <- session.Run(0) }()

	select {
	case <-secondChunkStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("download did not reach the second chunk")
	}
	if !session.Pause() {
		t.Fatal("active download could not be paused")
	}
	select {
	case <-paused:
	case <-time.After(3 * time.Second):
		t.Fatal("download did not enter the paused state")
	}

	if info, err := os.Stat(targetPath); err != nil {
		t.Fatalf("partial file was not preserved: %v", err)
	} else if info.Size() != 32 {
		t.Fatalf("expected a 32-byte prefix while paused, got %d", info.Size())
	}
	if !session.Resume() {
		t.Fatal("paused download could not be resumed")
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("resumed download failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("resumed download did not finish")
	}

	contents, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(bytes.Repeat([]byte{'a'}, 32), bytes.Repeat([]byte{'b'}, 32)...)
	if !bytes.Equal(contents, expected) {
		t.Fatal("resumed file does not contain the expected data")
	}
}

func TestDownloadSessionSwitchesServerAndResumes(t *testing.T) {
	targetPath := filepath.Join(t.TempDir(), "model.zip")
	blockedChunk := make(chan struct{})
	firstAttempt := sync.Once{}

	firstServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Length", "64")
			writer.WriteHeader(http.StatusOK)
			return
		}
		if request.Header.Get("Range") == "bytes=0-31" {
			writer.WriteHeader(http.StatusPartialContent)
			_, _ = writer.Write(bytes.Repeat([]byte{'a'}, 32))
			return
		}
		firstAttempt.Do(func() { close(blockedChunk) })
		<-request.Context().Done()
	}))
	defer firstServer.Close()

	secondServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Length", "64")
			writer.WriteHeader(http.StatusOK)
			return
		}
		if request.Header.Get("Range") != "bytes=32-63" {
			t.Errorf("second server did not resume the partial file: %q", request.Header.Get("Range"))
			writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		writer.WriteHeader(http.StatusPartialContent)
		_, _ = writer.Write(bytes.Repeat([]byte{'b'}, 32))
	}))
	defer secondServer.Close()

	session := NewDownloadSession(
		[]string{firstServer.URL + "/model.zip", secondServer.URL + "/model.zip"},
		targetPath,
		1,
		32,
	)
	result := make(chan error, 1)
	go func() { result <- session.Run(0) }()

	select {
	case <-blockedChunk:
	case <-time.After(3 * time.Second):
		t.Fatal("first server did not reach its blocked chunk")
	}
	if !session.SwitchServer() {
		t.Fatal("active download could not switch servers")
	}

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("download after server switch failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("download did not finish after changing servers")
	}
}
