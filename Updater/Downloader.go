package Updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type OnProgress func(bytesWritten, contentLength uint64)

type WriteCounter struct {
	Total         uint64
	ContentLength uint64
	OnProgress    OnProgress
}

type Download struct {
	Url          string
	Filepath     string
	WriteCounter WriteCounter
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.OnProgress(wc.Total, wc.ContentLength)
	return n, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func (d *Download) DownloadFile() error {

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err := os.Create(d.Filepath + ".tmp")
	if err != nil {
		return err
	}

	// Get the data
	resp, err := http.Get(d.Url)
	if err != nil {
		out.Close()
		return err
	}
	defer resp.Body.Close()

	d.WriteCounter.ContentLength = uint64(resp.ContentLength)

	// Create our progress reporter and pass it to be used alongside our writer
	if _, err = io.Copy(out, io.TeeReader(resp.Body, &d.WriteCounter)); err != nil {
		out.Close()
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(d.Filepath+".tmp", d.Filepath); err != nil {
		return err
	}
	return nil
}
