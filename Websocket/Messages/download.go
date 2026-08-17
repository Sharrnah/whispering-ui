package Messages

import (
	"fmt"
	"os"
	"whispering-tiger-ui/ModelDownloader"
)

type Download struct {
	Urls          []string `json:"urls"`
	ExtractDir    string   `json:"extract_dir"`
	Checksum      string   `json:"checksum"`
	Title         string   `json:"title"`
	ExtractFormat string   `json:"extract_format"`
}

type DownloadMessage struct {
	Type     string   `json:"type"`
	Download Download `json:"data"`
}

func (res DownloadMessage) StartDownload() error {
	dl := res.Download
	finishedPath := dl.ExtractDir + ".finished"
	failedPath := dl.ExtractDir + ".failed"
	failedTemporaryPath := failedPath + ".tmp"
	for _, receiptPath := range []string{finishedPath, failedPath, failedTemporaryPath} {
		if err := os.Remove(receiptPath); err != nil && !os.IsNotExist(err) {
			cleanupErr := fmt.Errorf("remove stale download receipt %q: %w", receiptPath, err)
			return publishDownloadFailure(failedPath, cleanupErr)
		}
	}

	downloadErr := ModelDownloader.DownloadFile(
		dl.Urls,
		dl.ExtractDir,
		dl.Checksum,
		dl.Title,
		dl.ExtractFormat,
	)
	if downloadErr == nil {
		return nil
	}

	// Python waits for one of two atomic receipts. Without a failure receipt it
	// cannot distinguish a slow download from a UI-side error and waits forever.
	return publishDownloadFailure(failedPath, downloadErr)
}

func publishDownloadFailure(failedPath string, downloadErr error) error {
	failedTemporaryPath := failedPath + ".tmp"
	if writeErr := os.WriteFile(failedTemporaryPath, []byte(downloadErr.Error()), 0o644); writeErr != nil {
		return fmt.Errorf("%w (could not write failure receipt: %v)", downloadErr, writeErr)
	}
	if renameErr := os.Rename(failedTemporaryPath, failedPath); renameErr != nil {
		_ = os.Remove(failedTemporaryPath)
		return fmt.Errorf("%w (could not publish failure receipt: %v)", downloadErr, renameErr)
	}
	return downloadErr
}
