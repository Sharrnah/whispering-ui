package Updater

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCancelDownloadAndResumeFromAnotherServer(t *testing.T) {
	targetPath := filepath.Join(t.TempDir(), "model.zip")
	secondRequestStarted := make(chan struct{})

	firstServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				if info, err := os.Stat(targetPath); err == nil && info.Size() == 32 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			close(secondRequestStarted)
			<-request.Context().Done()
		default:
			t.Errorf("unexpected range sent to first server: %q", request.Header.Get("Range"))
			writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		}
	}))
	defer firstServer.Close()

	firstDownload := &Download{
		Url:                 firstServer.URL + "/model.zip",
		Filepath:            targetPath,
		ConcurrentDownloads: 1,
		ChunkSize:           32,
	}
	firstDownload.WriteCounter.OnProgress = func(uint64, uint64, float64) {}
	firstResult := make(chan error, 1)
	go func() {
		firstResult <- firstDownload.DownloadFile(0)
	}()

	select {
	case <-secondRequestStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("first server did not start the second chunk")
	}
	firstDownload.Cancel()

	select {
	case err := <-firstResult:
		if !errors.Is(err, ErrDownloadCanceled) {
			t.Fatalf("expected cancellation error, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("canceled download did not stop")
	}

	if info, err := os.Stat(targetPath); err != nil {
		t.Fatalf("partial file was not preserved: %v", err)
	} else if info.Size() != 32 {
		t.Fatalf("expected a 32-byte resumable prefix, got %d bytes", info.Size())
	}

	secondServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodHead {
			writer.Header().Set("Content-Length", "64")
			writer.WriteHeader(http.StatusOK)
			return
		}
		if request.Header.Get("Range") != "bytes=32-63" {
			t.Errorf("second server did not resume at the partial-file boundary: %q", request.Header.Get("Range"))
			writer.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		writer.WriteHeader(http.StatusPartialContent)
		_, _ = writer.Write(bytes.Repeat([]byte{'b'}, 32))
	}))
	defer secondServer.Close()

	secondDownload := &Download{
		Url:                 secondServer.URL + "/model.zip",
		Filepath:            targetPath,
		ConcurrentDownloads: 1,
		ChunkSize:           32,
	}
	secondDownload.WriteCounter.OnProgress = func(uint64, uint64, float64) {}
	if err := secondDownload.DownloadFile(0); err != nil {
		t.Fatalf("resume from second server failed: %v", err)
	}

	contents, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := append(bytes.Repeat([]byte{'a'}, 32), bytes.Repeat([]byte{'b'}, 32)...)
	if !bytes.Equal(contents, expected) {
		t.Fatal("resumed file did not contain the bytes from both servers in order")
	}
}

func TestCancelBeforeDownloadStarts(t *testing.T) {
	download := &Download{
		Url:      "http://127.0.0.1:1/model.zip",
		Filepath: filepath.Join(t.TempDir(), "model.zip"),
	}
	download.WriteCounter.OnProgress = func(uint64, uint64, float64) {}
	download.Cancel()

	if err := download.DownloadFile(0); !errors.Is(err, ErrDownloadCanceled) {
		t.Fatalf("expected pre-start cancellation, got %v", err)
	}
}

func TestResumeSpeedExcludesBytesAlreadyOnDisk(t *testing.T) {
	const existingBytes = uint64(900)
	const newlyDownloadedBytes = uint64(100)

	download := &Download{}
	download.WriteCounter.Total = existingBytes
	download.WriteCounter.startTime = time.Now().Add(-2 * time.Second)
	download.WriteCounter.speedMA = NewMovingAverage(movingAverageWindow)

	var reportedTotal uint64
	var reportedSpeed float64
	download.WriteCounter.OnProgress = func(total, _ uint64, speed float64) {
		reportedTotal = total
		reportedSpeed = speed
	}
	download.addBytes(newlyDownloadedBytes)

	if reportedTotal != existingBytes+newlyDownloadedBytes {
		t.Fatalf("progress must include the resumed prefix: got %d bytes", reportedTotal)
	}
	if reportedSpeed < 45 || reportedSpeed > 55 {
		t.Fatalf("speed must include only newly transferred bytes: got %.2f B/s", reportedSpeed)
	}
}
