package ModelDownloader

import (
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/big"
	"net/url"
	"os"
	"path"
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

func uniqueURLs(urls []string) []string {
	uniqueURLs := make([]string, 0, len(urls))
	seen := make(map[string]struct{}, len(urls))
	for _, downloadURL := range urls {
		if downloadURL == "" {
			continue
		}
		if _, exists := seen[downloadURL]; exists {
			continue
		}
		seen[downloadURL] = struct{}{}
		uniqueURLs = append(uniqueURLs, downloadURL)
	}
	return uniqueURLs
}

func randomDownloadServerIndex(serverCount int) int {
	if serverCount <= 1 {
		return 0
	}
	randomIndex, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(serverCount)))
	if err == nil {
		return int(randomIndex.Int64())
	}
	return int(time.Now().UnixNano() % int64(serverCount))
}

func rotateDownloadURLs(urls []string, startIndex int) []string {
	if len(urls) <= 1 {
		return urls
	}
	startIndex %= len(urls)
	if startIndex < 0 {
		startIndex += len(urls)
	}
	rotated := make([]string, 0, len(urls))
	rotated = append(rotated, urls[startIndex:]...)
	rotated = append(rotated, urls[:startIndex]...)
	return rotated
}

func randomizedUniqueURLs(urls []string) []string {
	unique := uniqueURLs(urls)
	return rotateDownloadURLs(unique, randomDownloadServerIndex(len(unique)))
}

func nextDownloadServerIndex(currentIndex, serverCount int) int {
	if serverCount <= 1 {
		return 0
	}
	return (currentIndex + 1) % serverCount
}

func downloadURLFilename(downloadURL string) string {
	parsedURL, err := url.Parse(downloadURL)
	if err == nil {
		if filename := path.Base(parsedURL.Path); filename != "." && filename != "/" {
			return filename
		}
	}
	return downloadURL[strings.LastIndex(downloadURL, "/")+1:]
}

func downloadServerName(downloadURL string) string {
	parsedURL, err := url.Parse(downloadURL)
	if err == nil && parsedURL.Hostname() != "" {
		return strings.SplitN(parsedURL.Hostname(), ".", 2)[0]
	}
	return downloadURL
}

func DownloadFile(urls []string, targetDir string, checksum string, title string, extractFormat string) error {
	downloadURLs := randomizedUniqueURLs(urls)
	if len(downloadURLs) == 0 {
		return fmt.Errorf("no download URLs were provided")
	}

	// If the file is already being downloaded, skip and return.
	if isDownloading(targetDir) {
		return fmt.Errorf("File is already being downloaded: %s", targetDir)
	}
	addDownload(targetDir)
	// Ensure removal on exit.
	defer removeDownload(targetDir)

	// find active window
	window, _ := Utilities.GetCurrentMainWindow("Downloading " + title)

	// get file name from download url
	filename := downloadURLFilename(downloadURLs[0])

	// create download dialog
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)

	dialogTitlePart := filename
	if title != "" {
		dialogTitlePart = title + " [" + filename + "]"
	}
	downloadDialog := dialog.NewCustomWithoutButtons("Downloading "+dialogTitlePart, statusBarContainer, window)
	downloadingLabel := widget.NewLabel(lang.L("Downloading...") + " ")

	var downloadStateMutex sync.Mutex
	var activeDownloader *Updater.Download
	switchRequested := false
	changeServerButton := widget.NewButton(lang.L("Change server"), nil)
	changeServerButton.OnTapped = func() {
		downloadStateMutex.Lock()
		if activeDownloader == nil || switchRequested || len(downloadURLs) <= 1 {
			downloadStateMutex.Unlock()
			return
		}
		switchRequested = true
		downloaderToCancel := activeDownloader
		downloadStateMutex.Unlock()

		changeServerButton.Disable()
		downloadingLabel.SetText(lang.L("Downloading...") + " ")
		downloaderToCancel.Cancel()
	}
	hideDownloadButton := widget.NewButton(lang.L("Hide (Download will continue)"), func() {
		downloadDialog.Hide()
	})

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
	if extractFormat == "none" {
		needsExtract = false
		extractType = ""
	} else if extractFormat != "" {
		needsExtract = true
		extractType = extractFormat
	}

	//appExec, _ := os.Executable()

	downloadTargetDir := filepath.Dir(targetDir)
	os.MkdirAll(downloadTargetDir, 0755)
	//downloadTargetFile := filepath.Join(downloadTargetDir, filename)
	receiptWriter := Updater.Download{Filepath: targetDir}

	// guard for adding the infinite bar only once
	var infiniteOnce sync.Once

	fyne.Do(func() {
		statusBarContainer.Add(downloadingLabel)
		statusBarContainer.Refresh()
		if len(downloadURLs) <= 1 {
			changeServerButton.Disable()
		}
		downloadDialog.SetButtons([]fyne.CanvasObject{changeServerButton, hideDownloadButton})
		downloadDialog.Show()
	})

	var err error
	serverIndex := 0
	automaticFailures := 0
	for {
		downloadURL := downloadURLs[serverIndex]
		downloader := &Updater.Download{
			Url:                 downloadURL,
			Filepath:            targetDir,
			ConcurrentDownloads: 4,
			ChunkSize:           15 * 1024 * 1024, // 15 MB
		}
		serverName := downloadServerName(downloadURL)
		downloader.WriteCounter.OnProgress = func(progress, total uint64, speed float64) {
			if int64(total) == -1 {
				infiniteOnce.Do(func() {
					fyne.Do(func() {
						statusBarContainer.Remove(statusBar)
						statusBarContainer.Add(widget.NewProgressBarInfinite())
						statusBarContainer.Refresh()
					})
				})
			} else {
				fyne.Do(func() {
					statusBar.Max = float64(total)
					statusBar.SetValue(float64(progress))
				})
				resumeStatusText := ""
				if downloader.IsResuming() {
					resumeStatusText = " (" + lang.L("Resuming") + ")"
				}

				speedStr := ""
				if speed < 1024 {
					speedStr = fmt.Sprintf("%.2f B/s", speed)
				} else if speed < 1024*1024 {
					speedStr = fmt.Sprintf("%.2f KiB/s", speed/1024)
				} else {
					speedStr = fmt.Sprintf("%.2f MiB/s", speed/(1024*1024))
				}
				fyne.Do(func() {
					downloadingLabel.SetText(lang.L("Downloading from location", map[string]interface{}{
						"Location":  serverName,
						"TotalSize": humanize.Bytes(total),
						"Speed":     speedStr,
					}) + " " + resumeStatusText)
				})
			}
		}

		downloadStateMutex.Lock()
		activeDownloader = downloader
		switchRequested = false
		downloadStateMutex.Unlock()
		displayServerIndex := serverIndex
		fyne.Do(func() {
			changeServerButton.SetText(fmt.Sprintf("%s (%d/%d)", lang.L("Change server"), displayServerIndex+1, len(downloadURLs)))
			if len(downloadURLs) > 1 {
				changeServerButton.Enable()
			} else {
				changeServerButton.Disable()
			}
		})

		err = downloader.DownloadFile(3)

		downloadStateMutex.Lock()
		wasSwitchRequested := switchRequested
		activeDownloader = nil
		switchRequested = false
		downloadStateMutex.Unlock()

		if err == nil {
			break
		}
		if errors.Is(err, Updater.ErrDownloadCanceled) && wasSwitchRequested {
			automaticFailures = 0
			serverIndex = nextDownloadServerIndex(serverIndex, len(downloadURLs))
			continue
		}
		Logging.CaptureException(err)
		automaticFailures++
		if automaticFailures >= len(downloadURLs) {
			break
		}
		serverIndex = nextDownloadServerIndex(serverIndex, len(downloadURLs))
	}
	fyne.Do(func() {
		changeServerButton.Disable()
	})

	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(err, window)
		})
		return err
	}

	// check if the file has the correct hash
	if checksum != "" {
		fyne.Do(func() {
			statusBarContainer.Add(widget.NewLabel(lang.L("Checking checksum...")))
		})
		if err := Updater.CheckFileHash(targetDir, checksum); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			if removeErr := os.Remove(targetDir); removeErr != nil && !os.IsNotExist(removeErr) {
				err = fmt.Errorf("%w (could not remove invalid archive: %v)", err, removeErr)
			}
			fyne.Do(func() {
				dialog.ShowError(err, window)
				statusBarContainer.Add(widget.NewLabel("Checksum check failed. Please retry. If it still fails, please contact support."))
			})
			return err
		}
	}

	// wait a bit before trying to extract
	if needsExtract {
		time.Sleep(1 * time.Second)
		fyne.Do(func() {
			statusBarContainer.Add(widget.NewLabel(lang.L("Extracting...")))
			statusBarContainer.Refresh()
		})
		if extractType == "zip" {
			err = Updater.Unzip(targetDir, downloadTargetDir)
		} else if extractType == "tar.gz" {
			err = Updater.Untar(targetDir, downloadTargetDir)
		}
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, window)
			})
			return err
		}
		err = receiptWriter.CreateFinishedFile(".finished", 5, 3*time.Second)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, window)
			})
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
		err = receiptWriter.CreateFinishedFile(".finished", 5, 3*time.Second)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, window)
			})
			return err
		}
	}

	fyne.Do(func() {
		changeServerButton.Disable()
		statusBarContainer.Add(widget.NewLabel(lang.L("Finished.")))
		downloadDialog.SetButtons([]fyne.CanvasObject{
			widget.NewButton(lang.L("Close"), func() {
				downloadDialog.Hide()
			}),
		})
		downloadDialog.Refresh()
		if reOpenAfterHide {
			downloadDialog.Show()
		} else {
			downloadDialog.Hide()
		}
	})

	fyne.Do(func() {
		statusBarContainer.Refresh()
	})

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
