package UpdateUtility

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
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

func packageDownloadURLs(packageInfo Updater.UpdateInfo, locale string) []string {
	preferredLocation := ""
	if Updater.IsUSLocale(locale) && len(packageInfo.LocationUrls["US"]) > 0 {
		preferredLocation = "US"
	} else if Updater.IsEULocale(locale) && len(packageInfo.LocationUrls["EU"]) > 0 {
		preferredLocation = "EU"
	}

	locationOrder := make([]string, 0, len(packageInfo.LocationUrls))
	appendLocation := func(location string) {
		if location == "" {
			return
		}
		for _, existing := range locationOrder {
			if existing == location {
				return
			}
		}
		if len(packageInfo.LocationUrls[location]) > 0 {
			locationOrder = append(locationOrder, location)
		}
	}
	appendLocation(preferredLocation)
	appendLocation("DEFAULT")
	appendLocation("EU")
	appendLocation("US")

	remainingLocations := make([]string, 0, len(packageInfo.LocationUrls))
	for location := range packageInfo.LocationUrls {
		remainingLocations = append(remainingLocations, location)
	}
	sort.Strings(remainingLocations)
	for _, location := range remainingLocations {
		appendLocation(location)
	}

	urls := make([]string, 0)
	seen := make(map[string]struct{})
	for _, location := range locationOrder {
		for _, downloadURL := range packageInfo.LocationUrls[location] {
			if downloadURL == "" {
				continue
			}
			if _, exists := seen[downloadURL]; exists {
				continue
			}
			seen[downloadURL] = struct{}{}
			urls = append(urls, downloadURL)
		}
	}
	return urls
}

func downloadServerName(downloadURL string) string {
	parsedURL, err := url.Parse(downloadURL)
	if err == nil && parsedURL.Hostname() != "" {
		return strings.SplitN(parsedURL.Hostname(), ".", 2)[0]
	}
	return downloadURL
}

func versionDownload(updater Updater.UpdatePackages, packageName, filename string, window fyne.Window, startBackend bool, progressTitle string, noDismiss bool, cleanUpFunc func()) error {
	packageInfo, exists := updater.Packages[packageName]
	if !exists {
		return fmt.Errorf("update package %q is missing", packageName)
	}
	downloadURLs := packageDownloadURLs(packageInfo, Updater.GetLanguage())
	if len(downloadURLs) == 0 {
		return fmt.Errorf("update package %q: %w", packageName, Updater.ErrNoDownloadURLs)
	}

	downloadSession := Updater.NewDownloadSession(downloadURLs, filename, 4, 15*1024*1024)
	var reOpenAfterHide atomic.Bool

	var statusBar *widget.ProgressBar
	var statusBarContainer *fyne.Container
	var infiniteProgress *widget.ProgressBarInfinite
	determinateProgressVisible := true
	infiniteProgressVisible := false
	var downloadDialog *dialog.CustomDialog
	var downloadingLabel *widget.Label
	var changeServerButton *widget.Button
	var pauseDownloadButton *widget.Button

	showIndeterminateProgress := func() {
		if determinateProgressVisible {
			statusBarContainer.Remove(statusBar)
			determinateProgressVisible = false
		}
		if infiniteProgress == nil {
			infiniteProgress = widget.NewProgressBarInfinite()
		}
		if !infiniteProgressVisible {
			statusBarContainer.Add(infiniteProgress)
			infiniteProgressVisible = true
		}
		infiniteProgress.Show()
		infiniteProgress.Start()
		statusBarContainer.Refresh()
	}
	stopIndeterminateProgress := func() {
		if infiniteProgress == nil || !infiniteProgressVisible {
			return
		}
		infiniteProgress.Stop()
		infiniteProgress.Hide()
		statusBarContainer.Remove(infiniteProgress)
		infiniteProgressVisible = false
		statusBarContainer.Refresh()
	}
	showDeterminateProgress := func(progress, total uint64) {
		stopIndeterminateProgress()
		if !determinateProgressVisible {
			statusBarContainer.Add(statusBar)
			determinateProgressVisible = true
		}
		statusBar.Show()
		statusBar.Max = float64(total)
		statusBar.SetValue(float64(progress))
		statusBarContainer.Refresh()
	}

	fyne.DoAndWait(func() {
		statusBar = widget.NewProgressBar()
		statusBarContainer = container.NewVBox(statusBar)
		downloadingLabel = widget.NewLabel(lang.L("Downloading...") + " ")
		statusBarContainer.Add(downloadingLabel)
		downloadDialog = dialog.NewCustomWithoutButtons(progressTitle, statusBarContainer, window)

		changeServerButton = widget.NewButton(lang.L("Change server"), func() {
			if downloadSession.SwitchServer() {
				changeServerButton.Disable()
				pauseDownloadButton.Disable()
				downloadingLabel.SetText(lang.L("Downloading...") + " ")
			}
		})
		pauseDownloadButton = widget.NewButton(lang.L("Pause"), func() {
			if downloadSession.IsPaused() {
				if downloadSession.Resume() {
					pauseDownloadButton.Disable()
					downloadingLabel.SetText(lang.L("Downloading...") + " ")
				}
				return
			}
			if downloadSession.Pause() {
				pauseDownloadButton.Disable()
				changeServerButton.Disable()
				downloadingLabel.SetText(lang.L("Pausing...") + " ")
			}
		})
		changeServerButton.Disable()
		pauseDownloadButton.Disable()

		buttons := []fyne.CanvasObject{changeServerButton, pauseDownloadButton}
		if !noDismiss {
			buttons = append(buttons, widget.NewButton(lang.L("Hide (Download will continue)"), func() {
				downloadDialog.Hide()
			}))
		}
		downloadDialog.SetButtons(buttons)
		downloadDialog.SetOnClosed(func() {
			reOpenAfterHide.Store(true)
		})
		downloadDialog.Show()
	})

	showDownloadError := func(downloadErr error, allowRetry bool) {
		fyne.Do(func() {
			stopIndeterminateProgress()
			buttons := []fyne.CanvasObject{
				widget.NewButton(lang.L("Close"), func() { downloadDialog.Hide() }),
			}
			if allowRetry {
				buttons = append(buttons, widget.NewButton(lang.L("Retry Download"), func() {
					downloadDialog.Hide()
					go func() {
						if retryErr := versionDownload(updater, packageName, filename, window, startBackend, progressTitle, false, cleanUpFunc); retryErr != nil {
							Logging.CaptureException(retryErr)
						}
					}()
				}))
			}
			downloadDialog.SetButtons(buttons)
			dialog.ShowError(downloadErr, window)
		})
	}

	downloadSession.OnAttempt = func(_ string, serverIndex, serverCount int) {
		fyne.Do(func() {
			changeServerButton.SetText(fmt.Sprintf("%s (%d/%d)", lang.L("Change server"), serverIndex+1, serverCount))
			if serverCount > 1 {
				changeServerButton.Enable()
			} else {
				changeServerButton.Disable()
			}
			pauseDownloadButton.SetText(lang.L("Pause"))
			pauseDownloadButton.Enable()
		})
	}
	downloadSession.OnProgress = func(downloadURL string, progress, total uint64, speed float64, resuming bool) {
		if int64(total) == -1 {
			fyne.Do(showIndeterminateProgress)
			return
		}
		resumeStatusText := ""
		if resuming {
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
		serverName := downloadServerName(downloadURL)
		fyne.Do(func() {
			showDeterminateProgress(progress, total)
			downloadingLabel.SetText(lang.L("Downloading from location", map[string]interface{}{
				"Location":  serverName,
				"TotalSize": humanize.Bytes(total),
				"Speed":     speedStr,
			}) + " " + resumeStatusText)
		})
	}
	downloadSession.OnPaused = func() {
		fyne.Do(func() {
			downloadingLabel.SetText(lang.L("Download paused.") + " ")
			pauseDownloadButton.SetText(lang.L("Resume"))
			pauseDownloadButton.Enable()
			changeServerButton.Disable()
		})
	}
	downloadSession.OnResumed = func() {
		fyne.Do(func() {
			downloadingLabel.SetText(lang.L("Downloading...") + " ")
			pauseDownloadButton.SetText(lang.L("Pause"))
			pauseDownloadButton.Enable()
		})
	}

	err := downloadSession.Run(3)
	fyne.Do(func() {
		changeServerButton.Disable()
		pauseDownloadButton.Disable()
	})
	if err != nil {
		Logging.CaptureException(err)
		showDownloadError(err, true)
		return err
	}

	fyne.DoAndWait(func() {
		statusBarContainer.Add(widget.NewLabel(lang.L("Checking checksum...")))
		statusBarContainer.Refresh()
	})
	if err = Updater.CheckFileHash(filename, packageInfo.SHA256); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		if removeErr := os.Remove(filename); removeErr != nil && !os.IsNotExist(removeErr) {
			err = fmt.Errorf("%w (could not remove invalid archive: %v)", err, removeErr)
		}
		Logging.CaptureException(err)
		fyne.Do(func() {
			checksumCheckFailLabel := widget.NewLabel(lang.L("Checksum check failed. Please retry. If it still fails, please contact support."))
			checksumCheckFailLabel.Wrapping = fyne.TextWrapWord
			statusBarContainer.Add(checksumCheckFailLabel)
		})
		showDownloadError(err, true)
		return err
	}

	appExec, executableErr := os.Executable()
	if executableErr != nil {
		showDownloadError(executableErr, false)
		return executableErr
	}

	if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
		fyne.DoAndWait(func() {
			statusBarContainer.Add(widget.NewLabel(lang.L("Stopping Backend...")))
			statusBarContainer.Refresh()
		})
		RuntimeBackend.BackendsList[0].Stop()
		time.Sleep(1 * time.Second)
	}

	time.Sleep(2 * time.Second)
	if cleanUpFunc != nil {
		fyne.DoAndWait(func() {
			statusBarContainer.Add(widget.NewLabel(lang.L("Removing old version...")))
			statusBarContainer.Refresh()
		})
		cleanUpFunc()
	}

	fyne.DoAndWait(func() {
		statusBar.TextFormatter = func() string {
			if statusBar.Max <= 0 {
				return ""
			}
			percentage := statusBar.Value / statusBar.Max * 100
			return fmt.Sprintf("%.0f%% (%s / %s)", percentage, humanize.Bytes(uint64(statusBar.Value)), humanize.Bytes(uint64(statusBar.Max)))
		}
		showIndeterminateProgress()
		statusBarContainer.Add(widget.NewLabel(lang.L("Extracting...")))
		statusBarContainer.Refresh()
	})
	var firstExtractionProgress atomic.Bool
	if err = Updater.UnzipWithProgress(filename, filepath.Dir(appExec), func(extracted, total uint64) {
		updateProgress := func() {
			if total == 0 {
				showIndeterminateProgress()
				return
			}
			showDeterminateProgress(extracted, total)
		}
		if firstExtractionProgress.CompareAndSwap(false, true) {
			fyne.DoAndWait(updateProgress)
		} else {
			fyne.Do(updateProgress)
		}
	}); err != nil {
		Logging.CaptureException(err)
		showDownloadError(err, false)
		return err
	}
	if err = os.Remove(filename); err != nil {
		Logging.CaptureException(err)
		showDownloadError(err, false)
		return err
	}

	fyne.Do(func() {
		stopIndeterminateProgress()
		statusBarContainer.Add(widget.NewLabel(lang.L("Finished.")))
		downloadDialog.SetButtons([]fyne.CanvasObject{
			widget.NewButton(lang.L("Close"), func() { downloadDialog.Hide() }),
		})
		downloadDialog.Refresh()
		if reOpenAfterHide.Load() {
			downloadDialog.Show()
		}
		statusBarContainer.Refresh()
	})

	if startBackend && len(RuntimeBackend.BackendsList) > 0 && !RuntimeBackend.BackendsList[0].IsRunning() {
		fyne.Do(func() {
			statusBarContainer.Add(widget.NewLabel(lang.L("Restarting Backend") + "..."))
			statusBarContainer.Refresh()
		})
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

func reportUpdateCheckError(window fyne.Window, checkErr error) {
	Logging.CaptureException(checkErr)
	fyne.Do(func() {
		dialog.ShowError(checkErr, window)
	})
}

// VersionCheck performs network work synchronously and must be called from a
// worker goroutine. All Fyne updates are dispatched to the UI thread.
func VersionCheck(window fyne.Window, startBackend bool) (bool, error) {
	updater := Updater.UpdatePackages{}
	if err := updater.GetUpdateInfo(updateInfoUrl); err != nil {
		checkErr := fmt.Errorf("%s: %w", lang.L("Could not retrieve update information. Check your internet connection or firewall, then try again."), err)
		reportUpdateCheckError(window, checkErr)
		return false, checkErr
	}

	platformInfo, platformExists := updater.Packages["ai_platform"]
	appInfo, appExists := updater.Packages["app"]
	if !platformExists || platformInfo.Version == "" || !appExists || appInfo.Version == "" {
		checkErr := fmt.Errorf("%s", lang.L("The update information is incomplete. Please try again later."))
		reportUpdateCheckError(window, checkErr)
		return false, checkErr
	}

	updateAvailable := false
	platformFileWithoutVersion := !Utilities.FileExists(currentPlatformFile) && (Utilities.FileExists("audioWhisper/audioWhisper.exe") || Utilities.FileExists("audioWhisper.py"))
	platformRequiresUpdate := GetCurrentPlatformVersion() != platformInfo.Version
	platformMissing := !Utilities.FileExists("audioWhisper/audioWhisper.exe") && !Utilities.FileExists("audioWhisper.py")

	platformUpdateTitle := lang.L("Platform Update available")
	platformUpdateText := lang.L("There is a new Update of the Platform available. Update to new version now?", map[string]interface{}{"Version": platformInfo.Version})
	progressTitle := lang.L("Downloading Platform Update. (Please wait until this is finished!)")
	if platformMissing {
		platformRequiresUpdate = true
		platformUpdateTitle = lang.L("Platform not found")
		platformUpdateText = lang.L("No required Platform file found. Download version now?", map[string]interface{}{"Version": platformInfo.Version})
		progressTitle = lang.L("first-time Setup - Downloading Platform.\n(Please wait until this is finished!)")
	}

	if platformRequiresUpdate || platformFileWithoutVersion {
		updateAvailable = true
		fyne.Do(func() {
			dialog.ShowConfirm(platformUpdateTitle, platformUpdateText, func(confirmed bool) {
				if !confirmed {
					if platformFileWithoutVersion {
						platformInfo.WriteYaml(currentPlatformFile)
					}
					return
				}

				go func() {
					cleanUpFunc := func() {
						cleanupPaths := []string{"audioWhisper", "toolchain", "ffmpeg"}
						cleanupErrors := []error{}
						executablePath, executableErr := os.Executable()
						if executableErr != nil {
							cleanupErrors = append(cleanupErrors, executableErr)
						} else {
							for _, relPath := range cleanupPaths {
								oldVersionDir := filepath.Join(filepath.Dir(executablePath), relPath)
								if removeErr := os.RemoveAll(oldVersionDir); removeErr != nil {
									cleanupErrors = append(cleanupErrors, removeErr)
								}
							}
						}
						if len(cleanupErrors) > 0 {
							cleanupErr := fmt.Errorf("errors during cleanup: %v", cleanupErrors)
							Logging.CaptureException(cleanupErr)
							fyne.Do(func() { dialog.ShowError(cleanupErr, window) })
						}
					}

					if downloadErr := versionDownload(updater, "ai_platform", "audioWhisper_platform.zip", window, startBackend, progressTitle, true, cleanUpFunc); downloadErr == nil {
						platformInfo.WriteYaml(currentPlatformFile)
					}
				}()
			}, window)
		})
	}

	currentAppVersion := Utilities.AppVersion + "." + Utilities.AppBuild
	if appInfo.Version != currentAppVersion {
		updateAvailable = true
		fyne.Do(func() {
			dialog.ShowConfirm(lang.L("App Update available"), lang.L("There is a new Update of the App available. Open GitHub Release page now?"), func(confirmed bool) {
				if confirmed {
					uiReleaseURL, _ := url.Parse("https://github.com/Sharrnah/whispering-ui/releases/latest")
					_ = fyne.CurrentApp().OpenURL(uiReleaseURL)
				}
			}, window)
		})
	}

	return updateAvailable, nil
}
