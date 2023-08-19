package ModelDownloader

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whispering-tiger-ui/Updater"
)

const rootCacheFolder = ".cache"

func DownloadFile(urls []string, targetDir string, checksum string, title string, extractFormat string) error {
	// find active window
	window := fyne.CurrentApp().Driver().AllWindows()[0]
	if len(fyne.CurrentApp().Driver().AllWindows()) == 1 && fyne.CurrentApp().Driver().AllWindows()[0] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[0]
	} else if len(fyne.CurrentApp().Driver().AllWindows()) == 2 && fyne.CurrentApp().Driver().AllWindows()[1] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[1]
	} else {
		return fmt.Errorf("no active window found")
	}

	// select download url
	randomUrlIndex := rand.Int() % len(urls)
	downloadUrl := urls[randomUrlIndex]

	// get file name from download url
	filename := downloadUrl[strings.LastIndex(downloadUrl, "/")+1:]

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

	// create downloader
	downloader := Updater.Download{
		Url:                 downloadUrl,
		FallbackUrls:        urls,
		Filepath:            targetDir,
		ConcurrentDownloads: 4,
		ChunkSize:           15 * 1024 * 1024, // 15 MB
	}
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
		dialog.ShowError(err, window)
		return err
	}

	// check if the file has the correct hash
	statusBarContainer.Add(widget.NewLabel("Checking checksum..."))
	if err := Updater.CheckFileHash(targetDir, checksum); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		dialog.ShowError(err, window)
		statusBarContainer.Add(widget.NewLabel("Checksum check failed. Please delete temporary file and download again. If it still fails, please contact support."))
		return err
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
		err = os.Rename(targetDir, targetDir+".finished")
		if err != nil {
			dialog.ShowError(err, window)
			return err
		}
		//err = os.Remove(downloadTargetFile)
		//if err != nil {
		//	dialog.ShowError(err, window)
		//	return err
		//}
	} else {
		err = os.Rename(targetDir, targetDir+".finished")
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
	window := fyne.CurrentApp().Driver().AllWindows()[0]
	if len(fyne.CurrentApp().Driver().AllWindows()) == 1 && fyne.CurrentApp().Driver().AllWindows()[0] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[0]
	} else if len(fyne.CurrentApp().Driver().AllWindows()) == 2 && fyne.CurrentApp().Driver().AllWindows()[1] != nil {
		window = fyne.CurrentApp().Driver().AllWindows()[1]
	} else {
		return fmt.Errorf("no active window found")
	}

	err := DownloadFile(modelLinks.urls, modelCachePath, modelChecksum, modelName+" "+modelType, "")
	if err != nil {
		dialog.ShowError(err, window)
	}

	return err
}
