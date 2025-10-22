package Advanced

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"image/color"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/UpdateUtility"
	"whispering-tiger-ui/Utilities"
)

var FreshInstalledPlugins []string

func CreatePluginListWindow(closeFunction func(), backendRunning bool) {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Pages\\Advanced\\PluginList->CreatePluginListWindow")
	})

	md, err := UpdateUtility.DownloadFile(UpdateUtility.PluginListUrl)
	if err != nil {
		Logging.CaptureException(err)
		print(err)
		return
	}

	pluginListWindow := fyne.CurrentApp().NewWindow(lang.L("Plugin List"))

	windowSize := fyne.NewSize(1150, 700)
	if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
		newWindowSize := fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
		if newWindowSize.Height > 1 && newWindowSize.Width > 1 {
			windowSize.Height = newWindowSize.Height - 50
			windowSize.Width = newWindowSize.Width - 50
		}
	}
	pluginListWindow.Resize(windowSize)

	pluginListWindow.CenterOnScreen()

	CloseFunctionCall := func() {
		if closeFunction != nil {
			closeFunction()
		}
		if pluginListWindow != nil && pluginListWindow.Content().Visible() {
			pluginListWindow.Close()
		}
		if len(FreshInstalledPlugins) > 0 && backendRunning {
			dialog.NewConfirm(lang.L("New Plugins Installed"), lang.L("Would you like to restart Whispering Tiger now? (Required for new Plugins to load.)"), func(response bool) {
				// close running backend process
				if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
					infinityProcessDialog := dialog.NewCustom(lang.L("Restarting Backend"), lang.L("OK"), container.NewVBox(widget.NewLabel(lang.L("Restarting Backend")+"..."), widget.NewProgressBarInfinite()), fyne.CurrentApp().Driver().AllWindows()[0])
					infinityProcessDialog.Show()
					RuntimeBackend.BackendsList[0].Stop()
					time.Sleep(2 * time.Second)
					RuntimeBackend.BackendsList[0].Start()
					infinityProcessDialog.Hide()

					FreshInstalledPlugins = []string{}
				}
			}, fyne.CurrentApp().Driver().AllWindows()[0]).Show()
		}
	}
	pluginListWindow.SetCloseIntercept(CloseFunctionCall)

	// Create a new Fyne container for the table
	tableContainer := container.NewVBox()

	tableContainer.Add(widget.NewLabel(lang.L("Plugin List")))

	loadingBar := widget.NewProgressBarInfinite()

	table := UpdateUtility.ExtractTable(md)

	tableData := UpdateUtility.ParseTableIntoStruct(table, UpdateUtility.PluginRelativeUrlPrefix)

	localPluginFilesData := UpdateUtility.ParseLocalPluginFiles()

	checkAllButton := widget.NewButton(lang.L("Check all Plugins for Updates"), func() {
		loadingBar.Show()
		for _, row := range tableData {
			UpdateUtility.PluginsUpdateWidgetsRefresh(&row, localPluginFilesData)
		}
		loadingBar.Hide()
	})
	checkAllButton.Importance = widget.HighImportance
	// hide button as we already update on window open
	checkAllButton.Hide()

	grid := container.New(layout.NewFormLayout())

	// Set the content of the window to the table container
	scrollContainer := container.NewVScroll(grid)
	verticalContent := container.NewBorder(container.NewVBox(checkAllButton, loadingBar), nil, nil, nil, scrollContainer)
	pluginListWindow.SetContent(verticalContent)

	// Show and run the application
	pluginListWindow.Show()

	loadingBar.Start()

	// iterate over the table data and create a new widget for each row
	for _, row := range tableData {

		title := row.Title

		titleLabel := canvas.NewText(title, color.RGBA{255, 255, 255, 255})
		titleLabel.TextSize = theme.TextSubHeadingSize()
		titleLabel.TextStyle.Bold = true

		remoteVersionLabel := widget.NewLabel(lang.L("Newest V") + ": ")
		currentVersionLabel := canvas.NewText("  "+lang.L("Current V")+": ", color.RGBA{255, 255, 255, 255})
		currentVersionLabel.Move(fyne.NewPos(theme.Padding(), 0))

		author := row.Author
		authorLabel := widget.NewLabel(lang.L("Author") + ":\n" + author)

		row.Widgets.RemoteVersion = remoteVersionLabel
		row.Widgets.CurrentVersion = currentVersionLabel

		titleLink := row.TitleLink

		titleButton := widget.NewButtonWithIcon(lang.L("Update")+" / "+lang.L("Install"), theme.DownloadIcon(), nil)
		titleButton.OnTapped = func() {
			version, class, _, fileContent := UpdateUtility.FetchAndAnalyzePluginUrl(titleLink)

			pluginFileName := Utilities.CamelToSnake(class) + ".py"
			pluginFileDir := UpdateUtility.PluginDir + pluginFileName

			localPluginFile := UpdateUtility.FindLocalPluginFileByClass(localPluginFilesData, class)
			if localPluginFile.Class != "" {
				pluginFileDir = localPluginFile.FilePath
			}

			// write the file to disk
			err := os.WriteFile(pluginFileDir, fileContent, 0644)
			if err != nil {
				window, _ := Utilities.GetCurrentMainWindow("Error writing file")
				dialog.ShowError(err, window)
				log.Fatalf("Error writing file: %v", err)
			}

			//currentVersionLabel.SetText("Current V: " + version)
			currentVersionLabel.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			currentVersionLabel.Text = "  " + lang.L("Current V") + ": " + version
			currentVersionLabel.Refresh()

			titleButton.Importance = widget.LowImportance
			titleButton.SetText(lang.L("Installed"))

			// show success installed dialog

			dialog.ShowInformation(lang.L("Plugin Installed"), lang.L("Plugin has been installed. The Plugin is disabled by default.", map[string]interface{}{"Plugin": class})+"\n",
				pluginListWindow)

			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "plugin_install",
				Name:  "plugin",
				Value: map[string]string{"name": class, "file": pluginFileName},
			}
			sendMessage.SendMessage()

			// add to FreshInstalledPlugins list
			FreshInstalledPlugins = append(FreshInstalledPlugins, class)
		}

		row.Widgets.UpdateButton = titleButton

		descriptionText := strings.ReplaceAll(row.Description, "<sub>", "\n<sub>")
		descriptionText = strings.ReplaceAll(descriptionText, "</sub>", "</sub>")
		descriptionText = strings.ReplaceAll(descriptionText, "</br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br/>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<ul>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "</ul>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<li>", " - ")
		descriptionText = strings.ReplaceAll(descriptionText, "</li>", "\n")
		descriptionText = Utilities.ParseEmojisMarkdown(descriptionText)

		// remove html tags from description using regex
		re := regexp.MustCompile("<.*?>")
		descriptionText = re.ReplaceAllString(descriptionText, "")

		descriptionLabel := widget.NewRichTextFromMarkdown(descriptionText)
		descriptionLabel.Wrapping = fyne.TextWrapWord

		grid.Add(container.NewVBox(row.Widgets.UpdateButton, row.Widgets.RemoteVersion, row.Widgets.CurrentVersion))

		openPageButton := widget.NewButton(lang.L("Open Webpage"), func() {
			err := fyne.CurrentApp().OpenURL(parseURL(titleLink))
			if err != nil {
				Logging.CaptureException(err)
				dialog.ShowError(err, pluginListWindow)
			}
		})

		rightColumn := container.NewBorder(authorLabel, nil, nil, nil, openPageButton)

		previewLink := row.PreviewLink

		previewImageContainer := container.NewStack()
		previewBorder := CustomWidget.NewHoverableBorder(container.NewBorder(container.NewPadded(titleLabel), nil, nil, previewImageContainer, descriptionLabel),
			nil, nil)
		go func() {
			if previewLink != "" {
				previewFileUri, err := storage.ParseURI(previewLink)
				if err == nil {
					// is preview an image?
					if previewFileUri.Extension() == ".png" || previewFileUri.Extension() == ".jpg" || previewFileUri.Extension() == ".jpeg" {
						fyne.Do(func() {
							var previewImageStatic *canvas.Image = nil
							previewImageStatic = canvas.NewImageFromURI(previewFileUri)
							if previewImageContainer != nil && previewImageStatic != nil {
								previewImageStatic.ScaleMode = canvas.ImageScaleFastest
								previewImageStatic.FillMode = canvas.ImageFillContain
								previewImageStatic.SetMinSize(fyne.NewSize(220, 120))
								previewImageContainer.Add(previewImageStatic)
							}
						})
					}
					if previewFileUri.Extension() == ".gif" {
						fyne.Do(func() {
							previewImageAni, err := CustomWidget.NewAnimatedGif(previewFileUri)
							if err == nil {
								if previewImageContainer != nil {
									previewImageAni.SetMinSize(fyne.NewSize(230, 130))
									previewImageContainer.Add(previewImageAni)
									//previewImageAni.Start()
									previewImageAni.Stop()
								}
							}
						})
					}
				}
			}
		}()
		previewBorder.SetOnMouseEnter(func() {
			if len(previewBorder.GetContainer().Objects) == 0 {
				return
			}
			previewImageBorderContainer := previewBorder.GetContainer().Objects[2].(*fyne.Container)
			if len(previewImageBorderContainer.Objects) == 0 {
				return
			}
			previewImageObject := previewImageBorderContainer.Objects[0]
			// check if the previewImageObject is a AnimatedGif
			if previewImageObject != nil {
				if gif, ok := previewImageObject.(*CustomWidget.AnimatedGif); ok {
					gif.Start()
				}
			}
		})
		previewBorder.SetOnMouseLeave(func() {
			if len(previewBorder.GetContainer().Objects) == 0 {
				return
			}
			previewImageBorderContainer := previewBorder.GetContainer().Objects[2].(*fyne.Container)
			if len(previewImageBorderContainer.Objects) == 0 {
				return
			}
			previewImageObject := previewImageBorderContainer.Objects[0]
			// check if the previewImageObject is a AnimatedGif
			if previewImageObject != nil {
				if gif, ok := previewImageObject.(*CustomWidget.AnimatedGif); ok {
					gif.Stop()
				}
			}
		})

		descriptionBorder := container.NewBorder(nil, nil, nil, rightColumn, previewBorder)

		grid.Add(descriptionBorder)

		grid.Add(widget.NewSeparator())
		grid.Add(widget.NewSeparator())
	}

	loadingBar.Stop()
	loadingBar.Hide()

	// run the check all button once at window showing
	checkAllButton.OnTapped()
}
