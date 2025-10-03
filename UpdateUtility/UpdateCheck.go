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
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Updater"
	"whispering-tiger-ui/Utilities"
)

var updateInfoUrl = "https://s3.libs.space:9000/projects/whispering/latest.yaml"

var appExec, _ = os.Executable()
var appPath = filepath.Dir(appExec)
var currentPlatformFile = filepath.Join(appPath, ".current_platform.yaml")

func versionDownload(updater Updater.UpdatePackages, packageName, filename string, window fyne.Window, startBackend bool, progressTitle string, noDismiss bool, cleanUpFunc func()) error {
	statusBar := widget.NewProgressBar()
	statusBarContainer := container.NewVBox(statusBar)
	downloadDialog := dialog.NewCustomWithoutButtons(progressTitle, statusBarContainer, window)
	if !noDismiss {
		downloadDialog.SetButtons([]fyne.CanvasObject{
			&widget.Button{
				Text: lang.L("Hide (Download will continue)"),
				OnTapped: func() {
					downloadDialog.Hide()
				},
			},
		})
	}
	//downloadDialog := dialog.NewCustom(progressTitle, lang.L("Hide (Download will continue)"), statusBarContainer, window)
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
		Logging.CaptureException(err)
		downloadDialog.SetButtons([]fyne.CanvasObject{
			&widget.Button{
				Text: lang.L("Close"),
				OnTapped: func() {
					downloadDialog.Hide()
				},
			},
			// retry button
			&widget.Button{
				Text: lang.L("Retry Download"),
				OnTapped: func() {
					downloadDialog.Hide()
					err = versionDownload(updater, packageName, filename, window, startBackend, progressTitle, false, cleanUpFunc)
					if err != nil {
						Logging.CaptureException(err)
						dialog.ShowError(err, window)
					}
				},
			},
		})
		dialog.ShowError(err, window)
		return err
	}
	appExec, _ := os.Executable()

	// check if the file has the correct hash
	statusBarContainer.Add(widget.NewLabel(lang.L("Checking checksum...")))
	if err := Updater.CheckFileHash(filename, updater.Packages[packageName].SHA256); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		Logging.CaptureException(err)
		downloadDialog.SetButtons([]fyne.CanvasObject{
			&widget.Button{
				Text: lang.L("Close"),
				OnTapped: func() {
					downloadDialog.Hide()
				},
			},
			// retry button
			&widget.Button{
				Text: lang.L("Retry Download"),
				OnTapped: func() {
					downloadDialog.Hide()
					err = versionDownload(updater, packageName, filename, window, startBackend, progressTitle, false, cleanUpFunc)
					if err != nil {
						Logging.CaptureException(err)
						dialog.ShowError(err, window)
					}
				},
			},
		})
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
		Logging.CaptureException(err)
		downloadDialog.SetButtons([]fyne.CanvasObject{
			&widget.Button{
				Text: lang.L("Close"),
				OnTapped: func() {
					downloadDialog.Hide()
				},
			},
		})
		dialog.ShowError(err, window)
		return err
	}
	err = os.Remove(filename)
	if err != nil {
		Logging.CaptureException(err)
		downloadDialog.SetButtons([]fyne.CanvasObject{
			&widget.Button{
				Text: lang.L("Close"),
				OnTapped: func() {
					downloadDialog.Hide()
				},
			},
		})
		dialog.ShowError(err, window)
		return err
	}

	statusBarContainer.Add(widget.NewLabel(lang.L("Finished.")))
	downloadDialog.SetButtons([]fyne.CanvasObject{
		&widget.Button{
			Text: lang.L("Close"),
			OnTapped: func() {
				downloadDialog.Hide()
			},
		},
	})
	downloadDialog.Refresh()
	if reOpenAfterHide {
		downloadDialog.Show()
	}

	statusBarContainer.Refresh()

	// start backend
	if startBackend && !RuntimeBackend.BackendsList[0].IsRunning() {
		statusBarContainer.Add(widget.NewLabel(lang.L("Restarting Backend") + "..."))
		RuntimeBackend.BackendsList[0].Start()
	}

	return nil
}

func GetCurrentPlatformVersion() string {
	if Utilities.FileExists(currentPlatformFile) {
		currentPlatformVersion := Updater.UpdateInfo{}
		data, err := os.ReadFile(currentPlatformFile)
		if err == nil {
			_ = currentPlatformVersion.ReadYaml(data)
			return currentPlatformVersion.Version
		}
	}
	return ""
}

func VersionCheck(window fyne.Window, startBackend bool) bool {
	updateAvailable := false

	updater := Updater.UpdatePackages{}
	err := updater.GetUpdateInfo(updateInfoUrl)
	if err != nil {
		return false
	}

	// check platform version
	platformFileWithoutVersion := !Utilities.FileExists(currentPlatformFile) && (Utilities.FileExists("audioWhisper/audioWhisper.exe") || Utilities.FileExists("audioWhisper.py"))
	platformRequiresUpdate := false

	if GetCurrentPlatformVersion() != updater.Packages["ai_platform"].Version {
		platformRequiresUpdate = true
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
						errors := []error{}
						cleanupPaths := []string{
							"audioWhisper",
							"toolchain",
							"ffmpeg",
						}
						appExec, _ := os.Executable()
						for _, relPath := range cleanupPaths {
							oldVersionDir := filepath.Join(filepath.Dir(appExec), relPath)
							err := os.RemoveAll(oldVersionDir)
							if err != nil {
								errors = append(errors, err)
							}
						}
						if len(errors) > 0 {
							err := fmt.Errorf("Errors during cleanup: %v", errors)
							Logging.CaptureException(err)
							dialog.ShowError(err, window)
						}
					}
					err = versionDownload(updater, "ai_platform", "audioWhisper_platform.zip", window, startBackend, progressTitle, true, cleanUpFunc)
					if err == nil {
						packageInfo := updater.Packages["ai_platform"]
						packageInfo.WriteYaml(currentPlatformFile)
					}
				}()
			} else {
				if platformFileWithoutVersion {
					packageInfo := updater.Packages["ai_platform"]
					packageInfo.WriteYaml(currentPlatformFile)
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
