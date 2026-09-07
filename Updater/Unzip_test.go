package Updater

import (
	"archive/zip"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUnzipPreservesExecutableAndTraversableParents(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix executable permissions")
	}
	root := t.TempDir()
	archive := filepath.Join(root, "backend.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	dataHeader := &zip.FileHeader{Name: "audioWhisper/_internal/data.txt"}
	dataHeader.SetMode(0644)
	data, err := writer.CreateHeader(dataHeader)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = data.Write([]byte("runtime data"))
	header := &zip.FileHeader{Name: "audioWhisper/audioWhisper"}
	header.SetMode(0755)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("#!/bin/sh\nprintf 'backend-ready'\n"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(dest, "audioWhisper"), 0755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dest, "audioWhisper", "audioWhisper")
	if err := os.WriteFile(binary, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Unzip(archive, dest); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dest, "audioWhisper", "_internal"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 != 0111 {
		t.Fatalf("non-traversable parent: %v", info.Mode())
	}
	output, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatal(err)
	}
	if string(output) != "backend-ready" {
		t.Fatalf("unexpected backend output: %q", output)
	}
}

func writeTestZIP(t *testing.T, archivePath string, files map[string][]byte) {
	t.Helper()
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	zipWriter := zip.NewWriter(archiveFile)
	for name, contents := range files {
		entry, err := zipWriter.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUnzipWithProgressReportsUncompressedBytes(t *testing.T) {
	temporaryDir := t.TempDir()
	archivePath := filepath.Join(temporaryDir, "package.zip")
	destination := filepath.Join(temporaryDir, "extracted")
	firstContents := bytes.Repeat([]byte("a"), 3*1024*1024)
	secondContents := bytes.Repeat([]byte("b"), 2*1024*1024)
	writeTestZIP(t, archivePath, map[string][]byte{
		"folder/first.bin": firstContents,
		"second.bin":       secondContents,
	})

	type progressEvent struct {
		extracted uint64
		total     uint64
	}
	events := []progressEvent{}
	err := UnzipWithProgress(archivePath, destination, func(extracted, total uint64) {
		events = append(events, progressEvent{extracted: extracted, total: total})
	})
	if err != nil {
		t.Fatalf("extract ZIP: %v", err)
	}

	expectedTotal := uint64(len(firstContents) + len(secondContents))
	if len(events) < 2 {
		t.Fatalf("expected initial and final progress events, got %v", events)
	}
	if events[0] != (progressEvent{extracted: 0, total: expectedTotal}) {
		t.Fatalf("unexpected initial progress: %+v", events[0])
	}
	if events[len(events)-1] != (progressEvent{extracted: expectedTotal, total: expectedTotal}) {
		t.Fatalf("unexpected final progress: %+v", events[len(events)-1])
	}
	for index, event := range events {
		if event.total != expectedTotal {
			t.Fatalf("event %d changed its total: %+v", index, event)
		}
		if index > 0 && event.extracted < events[index-1].extracted {
			t.Fatalf("progress moved backwards at event %d: %v", index, events)
		}
	}

	extractedFirst, err := os.ReadFile(filepath.Join(destination, "folder", "first.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(extractedFirst, firstContents) {
		t.Fatal("first extracted file has unexpected contents")
	}
}

func TestUnzipWrapperStillExtractsWithoutProgress(t *testing.T) {
	temporaryDir := t.TempDir()
	archivePath := filepath.Join(temporaryDir, "package.zip")
	destination := filepath.Join(temporaryDir, "extracted")
	writeTestZIP(t, archivePath, map[string][]byte{"file.txt": []byte("content")})

	if err := Unzip(archivePath, destination); err != nil {
		t.Fatalf("extract ZIP: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(destination, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "content" {
		t.Fatalf("unexpected extracted contents: %q", contents)
	}
}

func TestUnzipWithProgressRejectsTraversal(t *testing.T) {
	temporaryDir := t.TempDir()
	archivePath := filepath.Join(temporaryDir, "package.zip")
	destination := filepath.Join(temporaryDir, "extracted")
	writeTestZIP(t, archivePath, map[string][]byte{"../outside.txt": []byte("blocked")})

	err := UnzipWithProgress(archivePath, destination, func(uint64, uint64) {})
	if err == nil || !strings.Contains(err.Error(), "illegal file path") {
		t.Fatalf("expected traversal error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(temporaryDir, "outside.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("archive wrote outside its destination: %v", statErr)
	}
}
