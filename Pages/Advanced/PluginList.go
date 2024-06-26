package Advanced

import (
	"bufio"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/Utilities"
)

const PLUGIN_DIR = "./Plugins/"

var FreshInstalledPlugins []string

func getVersionAndClassFromReader(pluginCode io.Reader) (string, string, string) {
	// Read the entire content into a byte slice
	content, err := io.ReadAll(pluginCode)
	if err != nil {
		log.Fatalf("Error reading content: %v", err)
	}

	version, class := "", ""
	scanner := bufio.NewScanner(bytes.NewReader(content))
	versionLine, classLine := "", ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			re := regexp.MustCompile(`(?:Version|V)\s*[:]*\s*\d+\.\d+\.\d+`)
			if re.MatchString(line) {
				versionLine = line
			}
		}
		if strings.Contains(line, "class") && strings.Contains(line, "(Plugins.Base)") {
			classLine = line
			break // class should always be after the version line, so we can break here
		}
	}

	hash, err := Utilities.FileHash(bytes.NewReader(content))
	if err != nil {
		fmt.Printf("Error calculating hash: %v\n", err)
	}

	if versionLine == "" && classLine == "" {
		return versionLine, classLine, hash
	}

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	versionMatches := re.FindStringSubmatch(versionLine)
	if len(versionMatches) > 0 {
		version = versionMatches[0]
	}

	re = regexp.MustCompile(`class\s+(\w+)\(Plugins\.Base\)`)
	classMatches := re.FindStringSubmatch(classLine)
	if len(classMatches) > 1 {
		class = classMatches[1]
	}

	return version, class, hash
}

// fetchAndAnalyzeGist fetches the gist at the given URL and analyzes it for version and class information
// returns the version, class, hash and binary of the gist
func fetchAndAnalyzeGist(url string) (string, string, string, []byte) {
	// Check for GitHub domain in the URL
	if strings.Contains(url, "github.com") {
		// Handle GitHub URLs
		if strings.Contains(url, "/blob/") {
			// Replace "/blob/" with "/raw/" for regular GitHub files
			url = strings.Replace(url, "/blob/", "/raw/", 1)
		} else if strings.Contains(url, "gist.") {
			// For Gist links, append "/raw" at the end of the URL
			url += "/raw"
		}
	}
	// Future extension: Add else if conditions for other domains like GitLab

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching gist: %v\n", err)
		return "err", "err", "", nil
	}
	defer resp.Body.Close()

	// Read the entire content into a byte slice
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading content: %v", err)
	}

	version, class, hash := getVersionAndClassFromReader(bytes.NewReader(content))

	return version, class, hash, content
}

func CreatePluginListWindow(closeFunction func()) {
	defer Utilities.PanicLogger()

	fileUrl := "https://github.com/Sharrnah/whispering-plugins/raw/main/README.md"
	relativeUrlPreviewPrefixPart := "https://raw.githubusercontent.com/Sharrnah/whispering-plugins/main/"

	pluginListWindow := fyne.CurrentApp().NewWindow("Plugin List")

	windowSize := fyne.NewSize(1150, 700)
	if len(fyne.CurrentApp().Driver().AllWindows()) > 0 {
		windowSize = fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size()
		windowSize.Height = windowSize.Height - 50
		windowSize.Width = windowSize.Width - 50
	}
	pluginListWindow.Resize(windowSize)

	pluginListWindow.CenterOnScreen()

	CloseFunctionCall := func() {
		closeFunction()
		if pluginListWindow != nil && pluginListWindow.Content().Visible() {
			pluginListWindow.Close()
		}
		if len(FreshInstalledPlugins) > 0 {
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

	md, err := DownloadFile(fileUrl)
	if err != nil {
		panic(err)
	}

	table := extractTable(md)

	tableData := parseTableIntoStruct(table, relativeUrlPreviewPrefixPart)

	localPluginFilesData := parseLocalPluginFiles()

	checkAllButton := widget.NewButton("Check all Plugins for Updates", func() {
		loadingBar.Show()
		for _, row := range tableData {
			titleLink := row.TitleLink
			fmt.Println("Checking update for: " + titleLink)
			if row.Widgets.RemoteVersion != nil {
				remoteVersion, class, hash, _ := fetchAndAnalyzeGist(titleLink)

				localPluginFile := findLocalPluginFileByClass(localPluginFilesData, class)

				localVersion := localPluginFile.LocalVersion

				row.Widgets.RemoteVersion.SetText("Newest V: " + remoteVersion)
				//row.Widgets.CurrentVersion.SetText("Current V: " + localVersion)
				row.Widgets.CurrentVersion.Text = "  Current V: " + localVersion

				if remoteVersion != localVersion && localVersion != "" {
					row.Widgets.CurrentVersion.Color = color.RGBA{R: 240, G: 0, B: 0, A: 255}
					row.Widgets.UpdateButton.Importance = widget.HighImportance
					row.Widgets.UpdateButton.SetText("Update")
				} else {
					row.Widgets.CurrentVersion.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
					row.Widgets.UpdateButton.Importance = widget.LowImportance
					if hash != localPluginFile.SHA256 {
						row.Widgets.UpdateButton.Importance = widget.HighImportance
					}
					if (localVersion == "" && remoteVersion != "") || (localVersion == "" && remoteVersion == "") {
						row.Widgets.UpdateButton.SetText("Install")
					} else {
						row.Widgets.UpdateButton.SetText("ReInstall")
					}
				}
				row.Widgets.CurrentVersion.Refresh()

				fmt.Println("found remote version: " + remoteVersion + ", local version: " + localVersion + " class: " + class)
				fmt.Println("remote sha256: " + hash + ", local sha256: " + localPluginFile.SHA256)
			}
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

		titleLabel := widget.NewLabel(title)
		titleLabel.Wrapping = fyne.TextWrapWord

		remoteVersionLabel := widget.NewLabel("Newest V: ")
		currentVersionLabel := canvas.NewText("  Current V: ", color.RGBA{255, 255, 255, 255})
		currentVersionLabel.Move(fyne.NewPos(10, 0))

		author := row.Author
		authorLabel := widget.NewLabel("Author:\n" + author)

		row.Widgets.RemoteVersion = remoteVersionLabel
		row.Widgets.CurrentVersion = currentVersionLabel

		titleLink := row.TitleLink

		titleButton := widget.NewButtonWithIcon("Update / Install", theme.DownloadIcon(), nil)
		titleButton.OnTapped = func() {
			version, class, _, fileContent := fetchAndAnalyzeGist(titleLink)

			pluginFileName := PLUGIN_DIR + Utilities.CamelToSnake(class) + ".py"

			localPluginFile := findLocalPluginFileByClass(localPluginFilesData, class)
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

		grid.Add(container.NewVBox(titleLabel, row.Widgets.UpdateButton, row.Widgets.RemoteVersion, row.Widgets.CurrentVersion))
		descriptionScroller := container.NewVScroll(descriptionLabel)
		descriptionScroller.Resize(fyne.NewSize(descriptionScroller.Size().Width, 400))

		openPageButton := widget.NewButton("Open Page", func() {
			err := fyne.CurrentApp().OpenURL(parseURL(titleLink))
			if err != nil {
				dialog.ShowError(err, pluginListWindow)
			}
		})

		rightColumn := container.NewBorder(authorLabel, nil, nil, nil, openPageButton)

		previewLink := row.PreviewLink

		previewImageContainer := container.NewStack()
		previewBorder := container.NewBorder(nil, nil, nil, previewImageContainer, descriptionScroller)

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

type TableDataWidgets struct {
	UpdateButton   *widget.Button
	CurrentVersion *canvas.Text
	RemoteVersion  *widget.Label
}

type TableData struct {
	Title       string
	TitleLink   string
	Preview     string
	PreviewLink string
	Description string
	Author      string
	Version     string
	Widgets     *TableDataWidgets
}

type LocalPluginFilesData struct {
	Class        string
	FilePath     string
	LocalVersion string
	SHA256       string
}

func parseLocalPluginFiles() []LocalPluginFilesData {
	var localPluginFiles []LocalPluginFilesData
	// build plugins list
	var pluginFiles []string
	files, err := os.ReadDir(PLUGIN_DIR)
	if err != nil {
		println(err)
	}

	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "__init__") && (strings.HasSuffix(file.Name(), ".py")) {
			pluginFiles = append(pluginFiles, file.Name())
			pluginPath := PLUGIN_DIR + file.Name()

			// read local file and read its contents
			data, err := os.ReadFile(pluginPath)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return nil
			}
			// Convert the byte slice to a string
			pluginCode := string(data)
			pluginCodeReader := strings.NewReader(pluginCode)
			pluginVersion, pluginClass, sha256 := getVersionAndClassFromReader(pluginCodeReader)

			localPluginFiles = append(localPluginFiles, LocalPluginFilesData{
				Class:        pluginClass,
				FilePath:     pluginPath,
				LocalVersion: pluginVersion,
				SHA256:       sha256,
			})
		}
	}
	return localPluginFiles
}

func findLocalPluginFileByClass(localPluginFiles []LocalPluginFilesData, class string) LocalPluginFilesData {
	for _, file := range localPluginFiles {
		if file.Class == class {
			return file
		}
	}
	return LocalPluginFilesData{}
}

func DownloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractTable(md string) string {
	lines := strings.Split(md, "\n")
	tableStarted := false
	table := ""

	for _, line := range lines {
		if strings.HasPrefix(line, "|") {
			tableStarted = true
		}
		if tableStarted {
			table += line + "\n"
			if strings.HasPrefix(line, "|---") {
				tableStarted = false
			}
		}
	}

	return table
}

func parseTableIntoStruct(table string, relativeUrlPreviewPart string) []TableData {
	lines := strings.Split(table, "\n")
	var tableData []TableData
	reLink := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)            // regex to match markdown links
	reImg := regexp.MustCompile(`<img src=(.*?) width=.*>`)       // regex to match images
	reVid := regexp.MustCompile(`<video src=\'(.*?)\' width=.*>`) // regex to match videos

	for i, line := range lines {
		if i < 2 { // skip the header and separator line
			continue
		}
		cells := strings.Split(line, "|")

		if len(cells) < 5 { // skip if the row has less than 4 cells
			continue
		}

		// Parse title and title link
		matches := reLink.FindStringSubmatch(cells[1])
		title := ""
		titleLink := ""
		if len(matches) > 2 {
			title = matches[1]
			titleLink = matches[2]
		}

		// Parse preview and preview link
		matches = reImg.FindStringSubmatch(cells[2])
		if len(matches) == 0 { // if no image, try to find a video
			matches = reVid.FindStringSubmatch(cells[2])
		}
		preview := ""
		previewLink := ""
		if len(matches) > 1 {
			previewLink = matches[1]
			preview = "Preview available"
		}

		// Check if Link is relative
		if !strings.HasPrefix(previewLink, "http") {
			previewLink = relativeUrlPreviewPart + previewLink
		}

		description := strings.TrimSpace(cells[3])
		author := strings.TrimSpace(cells[4])

		row := TableData{
			Title:       strings.ReplaceAll(title, "**", ""),
			TitleLink:   titleLink,
			Preview:     preview,
			PreviewLink: previewLink,
			Description: description,
			Author:      author,
			Version:     "",
			Widgets: &TableDataWidgets{
				UpdateButton:   nil,
				CurrentVersion: nil,
				RemoteVersion:  nil,
			},
		}
		tableData = append(tableData, row)
	}

	return tableData
}
