package ModelDownloader

import (
	"fmt"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/Updater"
	"whispering-tiger-ui/Utilities"
)

const rootCacheFolder = ".cache"

// Global variables to track active downloads.
var (
	activeDownloads      = []string{}
	activeDownloadsMutex sync.Mutex
)

func isDownloading(target string) bool {
	activeDownloadsMutex.Lock()
	defer activeDownloadsMutex.Unlock()
	for _, v := range activeDownloads {
		if v == target {
			return true
		}
	}
	return false
}

func addDownload(target string) {
	activeDownloadsMutex.Lock()
	defer activeDownloadsMutex.Unlock()
	activeDownloads = append(activeDownloads, target)
}

func removeDownload(target string) {
	activeDownloadsMutex.Lock()
	defer activeDownloadsMutex.Unlock()
	for i, v := range activeDownloads {
		if v == target {
			activeDownloads = append(activeDownloads[:i], activeDownloads[i+1:]...)
			break
		}
	}
}

func DownloadFile(urls []string, targetDir string, checksum string, title string, extractFormat string) error {
	// If the file is already being downloaded, skip and return.
	if isDownloading(targetDir) {
		return fmt.Errorf("File is already being downloaded: %s", targetDir)
	}
	addDownload(targetDir)
	// Ensure removal on exit.
	defer removeDownload(targetDir)

	// find active window
	window, _ := Utilities.GetCurrentMainWindow("Downloading " + title)

	// select download url
	randomUrlIndex := rand.Int() % len(urls)
	downloadUrl := urls[randomUrlIndex]

	// get file name from download url
	filename := downloadUrl[strings.LastIndex(downloadUrl, "/")+1:]

	// create downloader
	downloader := Updater.Download{
		Url:                 downloadUrl,
		FallbackUrls:        urls,
		Filepath:            targetDir,
		ConcurrentDownloads: 4,
		ChunkSize:           15 * 1024 * 1024, // 15 MB
	}

	// create download dialog
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)

	dialogTitlePart := filename
	if title != "" {
		dialogTitlePart = title + " [" + filename + "]"
	}
	downloadDialog := dialog.NewCustom("Downloading "+dialogTitlePart, "Hide (Download will continue)", statusBarContainer, window)
	downloadDialog.Show()
	downloadingLabel := widget.NewLabel("Downloading... ")

	reOpenAfterHide := false
	downloadDialog.SetOnClosed(func() {
		reOpenAfterHide = true
	})

	// is filename a zip file?
	needsExtract := false
	extractType := ""
	if strings.HasSuffix(filename, ".zip") {
		needsExtract = true
		extractType = "zip"
	} else if strings.HasSuffix(filename, ".tar.gz") {
		needsExtract = true
		extractType = "tar.gz"
	}
	if extractFormat != "" {
		needsExtract = true
		extractType = extractFormat
	} else if extractFormat == "none" {
		needsExtract = false
		extractType = ""
	}

	// get subdomain from download url
	subdomain := downloadUrl[strings.Index(downloadUrl, "://")+3 : strings.Index(downloadUrl, ".")]

	//appExec, _ := os.Executable()

	downloadTargetDir := filepath.Dir(targetDir)
	os.MkdirAll(downloadTargetDir, 0755)
	//downloadTargetFile := filepath.Join(downloadTargetDir, filename)

	downloader.WriteCounter.OnProgress = func(progress, total uint64, speed float64) {
		if int64(total) == -1 {
			statusBarContainer.Remove(statusBar)
			statusBarContainer.Add(widget.NewProgressBarInfinite())
			statusBarContainer.Refresh()
		} else {
			statusBar.Max = float64(total)
			statusBar.SetValue(float64(progress))

			resumeStatusText := ""
			if downloader.IsResuming() {
				resumeStatusText = " (Resume)"
			}

			speedStr := ""
			if speed < 1024 {
				speedStr = fmt.Sprintf("%.2f B/s", speed)
			} else if speed < 1024*1024 {
				speedStr = fmt.Sprintf("%.2f KiB/s", speed/1024)
			} else {
				speedStr = fmt.Sprintf("%.2f MiB/s", speed/(1024*1024))
			}

			downloadingLabel.SetText("Downloading from " + subdomain + "... " + humanize.Bytes(total) + " (" + speedStr + ") " + resumeStatusText)
		}
	}

	statusBarContainer.Add(downloadingLabel)
	statusBarContainer.Refresh()
	err := downloader.DownloadFile(3)
	if err != nil {
		Logging.CaptureException(err)
		dialog.ShowError(err, window)
		return err
	}

	// check if the file has the correct hash
	if checksum != "" {
		statusBarContainer.Add(widget.NewLabel("Checking checksum..."))
		if err := Updater.CheckFileHash(targetDir, checksum); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			dialog.ShowError(err, window)
			statusBarContainer.Add(widget.NewLabel("Checksum check failed. Please delete temporary file and download again. If it still fails, please contact support."))
			return err
		}
	}

	// wait a bit before trying to extract
	if needsExtract {
		time.Sleep(1 * time.Second)

		statusBarContainer.Add(widget.NewLabel("Extracting..."))
		statusBarContainer.Refresh()
		if extractType == "zip" {
			err = Updater.Unzip(targetDir, downloadTargetDir)
		} else if extractType == "tar.gz" {
			err = Updater.Untar(targetDir, downloadTargetDir)
		}
		if err != nil {
			dialog.ShowError(err, window)
			return err
		}
		err = downloader.CreateFinishedFile(".finished", 5, 3*time.Second)
		if err != nil {
			dialog.ShowError(err, window)
			return err
		}

		//err = os.Rename(targetDir, targetDir+".finished")
		//if err != nil {
		//	dialog.ShowError(err, window)
		//	return err
		//}

		//err = os.Remove(downloadTargetFile)
		//if err != nil {
		//	dialog.ShowError(err, window)
		//	return err
		//}
	} else {
		//err = os.Rename(targetDir, targetDir+".finished")
		err = downloader.CreateFinishedFile(".finished", 5, 3*time.Second)
		if err != nil {
			dialog.ShowError(err, window)
			return err
		}
	}

	if err == nil {
		statusBarContainer.Add(widget.NewLabel("Finished."))
		downloadDialog.SetDismissText("Close")
		downloadDialog.Refresh()
		if reOpenAfterHide {
			downloadDialog.Show()
		} else {
			downloadDialog.Hide()
		}
	}

	statusBarContainer.Refresh()

	return nil
}

func (c *modelNameLinksMap) DownloadModel(modelName string, modelType string) error {
	// get model links from map
	modelLinksMap := (*c)[modelName].modelLink
	modelCachePath := (*c)[modelName].cachePath

	modelLinks := modelLinksMap[modelType]
	modelChecksum := modelLinks.checksum

	// find active window
	window, _ := Utilities.GetCurrentMainWindow("Downloading " + modelName + " " + modelType)

	err := DownloadFile(modelLinks.urls, modelCachePath, modelChecksum, modelName+" "+modelType, "")
	if err != nil {
		Logging.CaptureException(err)
		dialog.ShowError(err, window)
	}

	return err
}
