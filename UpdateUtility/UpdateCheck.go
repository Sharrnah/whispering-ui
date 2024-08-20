package UpdateUtility

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"time"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Updater"
	"whispering-tiger-ui/Utilities"
)

var updateInfoUrl = "https://s3.libs.space:9000/projects/whispering/latest.yaml"

func versionDownload(updater Updater.UpdatePackages, packageName, filename string, window fyne.Window, startBackend bool, progressTitle string, cleanUpFunc func()) error {
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)
	downloadDialog := dialog.NewCustom(progressTitle, lang.L("Hide (Download will continue)"), statusBarContainer, window)
	downloadDialog.Show()
	downloadingLabel := widget.NewLabel(lang.L("Downloading...") + " ")

	reOpenAfterHide := false
	downloadDialog.SetOnClosed(func() {
		reOpenAfterHide = true
	})

	hasEUServer := false
	hasUSServer := false
	var mergedUrls []string

	// go through all url locations in the yaml slice and check if it exists
	for langKey, locations := range updater.Packages[packageName].LocationUrls {
		if langKey == "EU" && len(locations) > 0 {
			hasEUServer = true
		}
		if langKey == "US" && len(locations) > 0 {
			hasUSServer = true
		}
		if len(locations) > 0 {
			mergedUrls = append(mergedUrls, locations...)
		}
	}

	randomUrlIndex := rand.Int() % len(mergedUrls)
	downloadUrl := mergedUrls[randomUrlIndex]

	locationString := "DEFAULT"
	// try to download from the closest server by checking the user's location
	pcLang := Updater.GetLanguage()
	print("lang: " + pcLang + "\n")
	if hasUSServer && Updater.IsUSLocale(pcLang) {
		locationString = "US"
		// Download from US server
		randomUrlIndex = rand.Int() % len(updater.Packages[packageName].LocationUrls["US"])
		downloadUrl = updater.Packages[packageName].LocationUrls["US"][randomUrlIndex]
	} else if hasEUServer && Updater.IsEULocale(pcLang) {
		locationString = "EU"
		// Download from EU server
		randomUrlIndex = rand.Int() % len(updater.Packages[packageName].LocationUrls["EU"])
		downloadUrl = updater.Packages[packageName].LocationUrls["EU"][randomUrlIndex]
	}

	downloader := Updater.Download{
		Url:                 downloadUrl,
		FallbackUrls:        mergedUrls,
		Filepath:            filename,
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

			downloadingLabel.SetText(lang.L("Downloading from location", map[string]interface{}{"Location": locationString, "TotalSize": humanize.Bytes(total), "Speed": speedStr}) + " " + resumeStatusText)
		}
	}

	statusBarContainer.Add(downloadingLabel)
	statusBarContainer.Refresh()
	err := downloader.DownloadFile(3)
	if err != nil {
		dialog.ShowError(err, window)
		return err
	}
	appExec, _ := os.Executable()

	// check if the file has the correct hash
	statusBarContainer.Add(widget.NewLabel(lang.L("Checking checksum...")))
	if err := Updater.CheckFileHash(filename, updater.Packages[packageName].SHA256); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		dialog.ShowError(err, window)
		checksumCheckFailLabel := widget.NewLabel(lang.L("Checksum check failed. Please delete temporary file and download again. If it still fails, please contact support."))
		checksumCheckFailLabel.Wrapping = fyne.TextWrapWord
		statusBarContainer.Add(checksumCheckFailLabel)
		return err
	}

	// close running backend process
	if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
		statusBarContainer.Add(widget.NewLabel(lang.L("Stopping Backend...")))
		RuntimeBackend.BackendsList[0].Stop()
		time.Sleep(1 * time.Second)
	}

	// wait a bit before trying to extract
	time.Sleep(2 * time.Second)

	// clean up before extracting
	if cleanUpFunc != nil {
		statusBarContainer.Add(widget.NewLabel(lang.L("Removing old version...")))
		cleanUpFunc()
	}

	// extract
	statusBarContainer.Add(widget.NewLabel(lang.L("Extracting...")))
	statusBarContainer.Refresh()
	err = Updater.Unzip(filename, filepath.Dir(appExec))
	if err != nil {
		dialog.ShowError(err, window)
		return err
	}
	err = os.Remove(filename)
	if err != nil {
		dialog.ShowError(err, window)
		return err
	}

	if err == nil {
		statusBarContainer.Add(widget.NewLabel(lang.L("Finished.")))
		downloadDialog.SetDismissText(lang.L("Close"))
		downloadDialog.Refresh()
		if reOpenAfterHide {
			downloadDialog.Show()
		}
	}

	statusBarContainer.Refresh()

	// start backend
	if startBackend && !RuntimeBackend.BackendsList[0].IsRunning() {
		statusBarContainer.Add(widget.NewLabel(lang.L("Restarting Backend") + "..."))
		RuntimeBackend.BackendsList[0].Start()
	}

	return nil
}

func VersionCheck(window fyne.Window, startBackend bool) bool {
	updateAvailable := false

	updater := Updater.UpdatePackages{}
	err := updater.GetUpdateInfo(updateInfoUrl)
	if err != nil {
		return false
	}

	// check platform version
	platformFileWithoutVersion := !Utilities.FileExists(".current_platform.yaml") && (Utilities.FileExists("audioWhisper/audioWhisper.exe") || Utilities.FileExists("audioWhisper.py"))
	platformRequiresUpdate := false
	if Utilities.FileExists(".current_platform.yaml") {
		currentPlatformVersion := Updater.UpdateInfo{}
		data, err := os.ReadFile(".current_platform.yaml")
		if err == nil {
			_ = currentPlatformVersion.ReadYaml(data)
			if currentPlatformVersion.Version != updater.Packages["ai_platform"].Version {
				platformRequiresUpdate = true
			}
		}
	}

	platformUpdateTitle := lang.L("Platform Update available")
	platformUpdateText := lang.L("There is a new Update of the Platform available. Update to new version now?", map[string]interface{}{"Version": updater.Packages["ai_platform"].Version})
	progressTitle := lang.L("Downloading Platform Update. (Please wait until this is finished!)")

	if !Utilities.FileExists("audioWhisper/audioWhisper.exe") && !Utilities.FileExists("audioWhisper.py") {
		platformRequiresUpdate = true
		platformUpdateTitle = lang.L("Platform not found")
		platformUpdateText = lang.L("No required Platform file found. Download version now?", map[string]interface{}{"Version": updater.Packages["ai_platform"].Version})
		progressTitle = lang.L("first-time Setup - Downloading Platform.\n(Please wait until this is finished!)")
	}

	if platformRequiresUpdate || platformFileWithoutVersion {
		updateAvailable = true
		dialog.ShowConfirm(platformUpdateTitle, platformUpdateText, func(b bool) {
			if b {
				go func() {
					cleanUpFunc := func() {
						appExec, _ := os.Executable()
						oldVersionDir := filepath.Join(filepath.Dir(appExec), "audioWhisper")
						err := os.RemoveAll(oldVersionDir)
						if err != nil {
							dialog.ShowError(err, window)
						}
					}
					err = versionDownload(updater, "ai_platform", "audioWhisper_platform.zip", window, startBackend, progressTitle, cleanUpFunc)
					if err == nil {
						packageInfo := updater.Packages["ai_platform"]
						packageInfo.WriteYaml(".current_platform.yaml")
					}
				}()
			} else {
				if platformFileWithoutVersion {
					packageInfo := updater.Packages["ai_platform"]
					packageInfo.WriteYaml(".current_platform.yaml")
				}
			}
		}, window)
	}

	// check app version
	currentAppVersion := Utilities.AppVersion + "." + Utilities.AppBuild
	if updater.Packages["app"].Version != currentAppVersion {
		updateAvailable = true
		dialog.ShowConfirm(lang.L("App Update available"), lang.L("There is a new Update of the App available. Open GitHub Release page now?"), func(b bool) {
			if b {
				uiReleaseUrl, _ := url.Parse("https://github.com/Sharrnah/whispering-ui/releases/latest")
				fyne.CurrentApp().OpenURL(uiReleaseUrl)
			}
		}, window)
	}

	return updateAvailable
}
