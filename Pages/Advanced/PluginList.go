package Advanced

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/UpdateUtility"
	"whispering-tiger-ui/Utilities"
)

var FreshInstalledPlugins []string

func CreatePluginListWindow(closeFunction func(), backendRunning bool) {
	defer Utilities.PanicLogger()

	md, err := UpdateUtility.DownloadFile(UpdateUtility.PluginListUrl)
	if err != nil {
		print(err)
		return
	}

	pluginListWindow := fyne.CurrentApp().NewWindow("Plugin List")

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
			dialog.NewConfirm("New Plugins Installed", "Would you like to restart Whispering Tiger now?\n(Required for new Plugins to load.)", func(response bool) {
				// close running backend process
				if len(RuntimeBackend.BackendsList) > 0 && RuntimeBackend.BackendsList[0].IsRunning() {
					infinityProcessDialog := dialog.NewCustom("Restarting Backend", "OK", container.NewVBox(widget.NewLabel("Restarting backend..."), widget.NewProgressBarInfinite()), fyne.CurrentApp().Driver().AllWindows()[0])
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

	tableContainer.Add(widget.NewLabel("Plugin List"))

	loadingBar := widget.NewProgressBarInfinite()

	table := UpdateUtility.ExtractTable(md)

	tableData := UpdateUtility.ParseTableIntoStruct(table, UpdateUtility.PluginRelativeUrlPrefix)

	localPluginFilesData := UpdateUtility.ParseLocalPluginFiles()

	checkAllButton := widget.NewButton("Check all Plugins for Updates", func() {
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

	// iterate over the table data and create a new widget for each row
	for _, row := range tableData {

		title := row.Title

		titleLabel := canvas.NewText(title, color.RGBA{255, 255, 255, 255})
		titleLabel.TextSize = theme.TextSubHeadingSize()
		titleLabel.TextStyle.Bold = true

		remoteVersionLabel := widget.NewLabel("Newest V: ")
		currentVersionLabel := canvas.NewText("  Current V: ", color.RGBA{255, 255, 255, 255})
		currentVersionLabel.Move(fyne.NewPos(theme.Padding(), 0))

		author := row.Author
		authorLabel := widget.NewLabel("Author:\n" + author)

		row.Widgets.RemoteVersion = remoteVersionLabel
		row.Widgets.CurrentVersion = currentVersionLabel

		titleLink := row.TitleLink

		titleButton := widget.NewButtonWithIcon("Update / Install", theme.DownloadIcon(), nil)
		titleButton.OnTapped = func() {
			version, class, _, fileContent := UpdateUtility.FetchAndAnalyzePluginUrl(titleLink)

			pluginFileName := UpdateUtility.PluginDir + Utilities.CamelToSnake(class) + ".py"

			localPluginFile := UpdateUtility.FindLocalPluginFileByClass(localPluginFilesData, class)
			if localPluginFile.Class != "" {
				pluginFileName = localPluginFile.FilePath
			}

			// write the file to disk
			err := os.WriteFile(pluginFileName, fileContent, 0644)
			if err != nil {
				log.Fatalf("Error writing file: %v", err)
			}

			//currentVersionLabel.SetText("Current V: " + version)
			currentVersionLabel.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			currentVersionLabel.Text = "  Current V: " + version
			currentVersionLabel.Refresh()

			titleButton.Importance = widget.LowImportance
			titleButton.SetText("Installed")

			// show success installed dialog
			dialog.ShowInformation("Plugin Installed", class+" has been installed. The Plugin is disabled by default.\n"+
				"Please restart Whispering Tiger to load the Plugin.\n",
				pluginListWindow)

			// add to FreshInstalledPlugins list
			FreshInstalledPlugins = append(FreshInstalledPlugins, class)
		}

		row.Widgets.UpdateButton = titleButton

		descriptionText := strings.ReplaceAll(row.Description, "</br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br/>", "\n")
		// remove html tags from description using regex
		re := regexp.MustCompile("<.*?>")
		descriptionText = re.ReplaceAllString(descriptionText, "")

		descriptionLabel := widget.NewLabel(descriptionText)
		descriptionLabel.Wrapping = fyne.TextWrapWord

		grid.Add(container.NewVBox(row.Widgets.UpdateButton, row.Widgets.RemoteVersion, row.Widgets.CurrentVersion))

		openPageButton := widget.NewButton("Open Page", func() {
			err := fyne.CurrentApp().OpenURL(parseURL(titleLink))
			if err != nil {
				dialog.ShowError(err, pluginListWindow)
			}
		})

		rightColumn := container.NewBorder(authorLabel, nil, nil, nil, openPageButton)

		previewLink := row.PreviewLink

		previewImageContainer := container.NewStack()
		previewBorder := container.NewBorder(container.NewPadded(titleLabel), nil, nil, previewImageContainer, descriptionLabel)

		go func() {
			if previewLink != "" {
				previewFileUri, err := storage.ParseURI(previewLink)
				if err == nil {
					// is preview an image?
					if previewFileUri.Extension() == ".png" || previewFileUri.Extension() == ".jpg" || previewFileUri.Extension() == ".jpeg" {
						var previewImageStatic *canvas.Image = nil
						previewImageStatic = canvas.NewImageFromURI(previewFileUri)
						if previewImageContainer != nil && previewImageStatic != nil {
							previewImageStatic.ScaleMode = canvas.ImageScaleFastest
							previewImageStatic.FillMode = canvas.ImageFillContain
							previewImageStatic.SetMinSize(fyne.NewSize(220, 120))
							previewImageContainer.Add(previewImageStatic)
						}
					}
					if previewFileUri.Extension() == ".gif" {
						previewImageAni, err := CustomWidget.NewAnimatedGif(previewFileUri)
						if err == nil {
							if previewImageContainer != nil {
								previewImageAni.SetMinSize(fyne.NewSize(230, 130))
								previewImageContainer.Add(previewImageAni)
								previewImageAni.Start()
							}
						}
					}
				}
			}
		}()

		descriptionBorder := container.NewBorder(nil, nil, nil, rightColumn, previewBorder)

		grid.Add(descriptionBorder)

		grid.Add(widget.NewSeparator())
		grid.Add(widget.NewSeparator())
	}

	loadingBar.Hide()

	// run the check all button once at window showing
	checkAllButton.OnTapped()
}
