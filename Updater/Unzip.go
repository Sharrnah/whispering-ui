package Updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ExtractionProgress func(extractedBytes, totalBytes uint64)

const extractionProgressInterval = 100 * time.Millisecond

type extractionProgressReporter struct {
	callback     ExtractionProgress
	total        uint64
	extracted    uint64
	lastReported uint64
	lastUpdate   time.Time
}

func newExtractionProgressReporter(total uint64, callback ExtractionProgress) *extractionProgressReporter {
	reporter := &extractionProgressReporter{
		callback:   callback,
		total:      total,
		lastUpdate: time.Now(),
	}
	if callback != nil {
		callback(0, total)
	}
	return reporter
}

func (r *extractionProgressReporter) add(bytes uint64) {
	r.extracted += bytes
	if r.callback == nil {
		return
	}
	if (r.total > 0 && r.extracted >= r.total) || time.Since(r.lastUpdate) >= extractionProgressInterval {
		r.callback(r.extracted, r.total)
		r.lastReported = r.extracted
		r.lastUpdate = time.Now()
	}
}

func (r *extractionProgressReporter) finish() {
	if r.callback != nil && r.lastReported != r.extracted {
		r.callback(r.extracted, r.total)
		r.lastReported = r.extracted
	}
}

type extractionProgressWriter struct {
	writer   io.Writer
	reporter *extractionProgressReporter
}

func (w extractionProgressWriter) Write(data []byte) (int, error) {
	written, err := w.writer.Write(data)
	w.reporter.add(uint64(written))
	return written, err
}

func Unzip(src, dest string) error {
	return UnzipWithProgress(src, dest, nil)
}

// UnzipWithProgress extracts a ZIP archive and reports progress using the
// exact total of its uncompressed file sizes. The callback is throttled to
// avoid flooding a UI event queue while large files are copied.
func UnzipWithProgress(src, dest string, onProgress ExtractionProgress) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	var totalBytes uint64
	for _, archiveFile := range r.File {
		if !archiveFile.FileInfo().IsDir() {
			totalBytes += archiveFile.UncompressedSize64
		}
	}
	progressReporter := newExtractionProgressReporter(totalBytes, onProgress)
	defer progressReporter.finish()

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(extractionProgressWriter{writer: f, reporter: progressReporter}, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func Untar(src, dest string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			panic(err)
		}
	}()

	tr := tar.NewReader(gzr)

	// Ensure the destination directory exists
	os.MkdirAll(dest, 0755)

	// Iterate over files in .tar.gz
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(dest, hdr.Name)

		// Check for TarSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		// Check if it's a dir or a file
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(path, 0755)
		} else if hdr.Typeflag == tar.TypeReg {
			os.MkdirAll(filepath.Dir(path), 0755)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
	return nil
}
