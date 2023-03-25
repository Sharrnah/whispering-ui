package Updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type OnProgress func(bytesWritten, contentLength uint64)

type WriteCounter struct {
	Total         uint64
	ContentLength uint64
	InitialBytes  uint64
	OnProgress    OnProgress
}

type Download struct {
	Url           string
	FallbackUrls  []string
	Filepath      string
	WriteCounter  WriteCounter
	ResumeSupport bool
	maxRetries    int
	urlIndex      int
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.OnProgress(wc.Total+wc.InitialBytes, wc.ContentLength)
	return n, nil
}

func (d *Download) DownloadFile(retries int) error {
	d.maxRetries = retries
	return d.downloadFileWithRetry(retries)
}

func (d *Download) getCurrentUrl() string {
	currentUrl := d.Url
	if d.urlIndex > 0 && d.urlIndex <= len(d.FallbackUrls) {
		currentUrl = d.FallbackUrls[d.urlIndex-1]
	}
	return currentUrl
}

func (d *Download) retryAction(retries int, err error) error {
	currentUrl := d.getCurrentUrl()

	if retries > 0 {
		fmt.Printf("Error downloading %s: %s. Retrying in 1 seconds...\n", d.Url, err.Error())
		time.Sleep(2 * time.Second)
		return d.downloadFileWithRetry(retries - 1)
	} else {
		if d.urlIndex < len(d.FallbackUrls) {
			fmt.Printf("All retries for URL %s have failed. Trying the next fallback URL...\n", currentUrl)
			d.urlIndex++
			return d.downloadFileWithRetry(d.maxRetries)
		} else {
			fmt.Printf("All retries for URL %s and all fallback URLs have failed.\n", currentUrl)
			return err
		}
	}
}

func (d *Download) downloadFileWithRetry(retries int) error {
	var out *os.File
	var err error

	currentUrl := d.getCurrentUrl()

	// Check if the file already exists and get its size
	var startBytes int64 = 0
	if _, err := os.Stat(d.Filepath + ".tmp"); err == nil {
		startBytes = d.getFileSize(d.Filepath + ".tmp")
	}

	// Create the file, but give it a tmp file extension, this means we won't overwrite a
	// file until it's downloaded, but we'll remove the tmp extension once downloaded.
	out, err = os.OpenFile(d.Filepath+".tmp", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Get the data
	var resp *http.Response
	if startBytes > 0 {
		req, err := http.NewRequest("GET", currentUrl, nil)
		if err != nil {
			out.Close()
			return d.retryAction(retries, err)
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startBytes))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			out.Close()
			return d.retryAction(retries, err)
		}
		if resp.StatusCode != http.StatusPartialContent {
			startBytes = 0
			resp.Body.Close()
			resp = nil

			// Close the file before truncating
			out.Close()

			// Truncate the file
			err = os.Truncate(d.Filepath+".tmp", 0)
			if err != nil {
				return err
			}

			// Reopen the file
			out, err = os.OpenFile(d.Filepath+".tmp", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
		} else {
			d.ResumeSupport = true
		}
	}
	if resp == nil {
		req, err := http.NewRequest("GET", currentUrl, nil)
		if err != nil {
			out.Close()
			return d.retryAction(retries, err)
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			out.Close()
			return d.retryAction(retries, err)
		}
	}

	d.WriteCounter.ContentLength = uint64(resp.ContentLength) + uint64(startBytes)
	d.WriteCounter.InitialBytes = uint64(startBytes)

	// Seek to the right position
	if startBytes > 0 {
		_, err = out.Seek(startBytes, 0)
		if err != nil {
			out.Close()
			return err
		}
	}

	// Create our progress reporter and pass it to be used alongside our writer
	if _, err = io.Copy(out, io.TeeReader(resp.Body, &d.WriteCounter)); err != nil {
		out.Close()
		resp.Body.Close()
		return d.retryAction(retries, err)
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Print("\n")

	// Close the file without defer so it can happen before Rename()
	out.Close()
	resp.Body.Close()

	if err = os.Rename(d.Filepath+".tmp", d.Filepath); err != nil {
		return err
	}
	return nil
}

func (d *Download) getFileSize(filepath string) int64 {
	file, err := os.Open(filepath)
	if err != nil {
		return 0
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
