package UpdateUtilitiy

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Updater"
	"whispering-tiger-ui/Utilities"
)

var updateInfoUrl = "https://s3.libs.space:9000/projects/whispering/latest.yaml"

func versionDownload(updater Updater.UpdatePackages, packageName, filename string, window fyne.Window, startBackend bool) error {
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)
	dialog.ShowCustom("Update in progress...", "Hide (Download will continue)", statusBarContainer, window)
	downloadingLabel := widget.NewLabel("Downloading... ")

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
	lang := Updater.GetLanguage()
	print("lang: " + lang + "\n")
	if hasUSServer && Updater.IsUSLocale(lang) {
		locationString = "US"
		// Download from US server
		randomUrlIndex = rand.Int() % len(updater.Packages[packageName].LocationUrls["US"])
		downloadUrl = updater.Packages[packageName].LocationUrls["US"][randomUrlIndex]
	} else if hasEUServer && Updater.IsEULocale(lang) {
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
	downloader.WriteCounter.OnProgress = func(progress, total uint64) {
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

			downloadingLabel.SetText("Downloading from " + locationString + "... " + humanize.Bytes(total) + resumeStatusText)
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
	statusBarContainer.Add(widget.NewLabel("Checking checksum..."))
	if err := Updater.CheckFileHash(filename, updater.Packages[packageName].SHA256); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		dialog.ShowError(err, window)
		statusBarContainer.Add(widget.NewLabel("Checksum check failed. Please delete temporary file and download again. If it still fails, please contact support."))
		return err
	}

	// close running backend process
	if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
		statusBarContainer.Add(widget.NewLabel("Stopping Backend..."))
		RuntimeBackend.BackendsList[0].Stop()
		time.Sleep(2 * time.Second)
	}

	statusBarContainer.Add(widget.NewLabel("Extracting..."))
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
		statusBarContainer.Add(widget.NewLabel("Finished."))
	}

	statusBarContainer.Refresh()

	// start backend
	if startBackend && !RuntimeBackend.BackendsList[0].IsRunning() {
		statusBarContainer.Add(widget.NewLabel("Restarting Backend..."))
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

	platformUpdateTitle := "Platform Update available"
	platformUpdateText := "There is a new Update of the Platform available. Update to " + updater.Packages["ai_platform"].Version + " now?"

	if !Utilities.FileExists("audioWhisper/audioWhisper.exe") && !Utilities.FileExists("audioWhisper.py") {
		platformRequiresUpdate = true
		platformUpdateTitle = "Platform not found"
		platformUpdateText = "No Platform file found. download version " + updater.Packages["ai_platform"].Version + " now?"
	}

	if platformRequiresUpdate || platformFileWithoutVersion {
		updateAvailable = true
		dialog.ShowConfirm(platformUpdateTitle, platformUpdateText, func(b bool) {
			if b {
				go func() {
					err = versionDownload(updater, "ai_platform", "audioWhisper_platform.zip", window, startBackend)
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
	currentAppVersion := fyne.CurrentApp().Metadata().Version + "." + strconv.Itoa(fyne.CurrentApp().Metadata().Build)
	if updater.Packages["app"].Version != currentAppVersion {
		updateAvailable = true
		dialog.ShowConfirm("App Update available", "There is a new Update of the App available. Open GitHub Release page now?", func(b bool) {
			if b {
				uiReleaseUrl, _ := url.Parse("https://github.com/Sharrnah/whispering-ui/releases/latest")
				fyne.CurrentApp().OpenURL(uiReleaseUrl)
			}
		}, window)
	}

	return updateAvailable
}
