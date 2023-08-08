package Advanced

import (
	"bufio"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func fetchAndAnalyzeGist(gistURL string) (string, string) {
	resp, err := http.Get(gistURL)
	if err != nil {
		fmt.Printf("Error fetching gist: %v\n", err)
		return "err", "err"
	}
	defer resp.Body.Close()

	version, class := "", ""

	scanner := bufio.NewScanner(resp.Body)
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

	if versionLine == "" && classLine == "" {
		return versionLine, classLine
	}

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	versionMatches := re.FindStringSubmatch(versionLine)
	if len(versionMatches) > 0 {
		version = versionMatches[0]
	}

	re = regexp.MustCompile(`class (\w+)`)
	classMatches := re.FindStringSubmatch(classLine)
	if len(classMatches) > 1 {
		class = classMatches[1]
	}

	return version, class
}

func CreatePluginListWindow() {
	fileUrl := "https://raw.githubusercontent.com/Sharrnah/whispering/main/documentation/plugins.md"

	pluginListWindow := fyne.CurrentApp().NewWindow("Plugin List")
	pluginListWindow.Resize(fyne.NewSize(1400, 800))

	// Create a new Fyne container for the table
	tableContainer := container.NewVBox()

	tableContainer.Add(widget.NewLabel("Plugin List"))

	md, err := DownloadFile(fileUrl)
	if err != nil {
		panic(err)
	}

	table := extractTable(md)

	tableData := parseTableIntoStruct(table)
	//fmt.Println(tableData)

	checkAllButton := widget.NewButton("Check all Plugins for Updates", func() {
		for _, row := range tableData {
			fmt.Println("Checking update for: " + row.TitleLink)
			if row.Widgets.RemoteVersion != nil {
				remoteVersion, class := fetchAndAnalyzeGist(row.TitleLink)
				row.Widgets.RemoteVersion.SetText("Newest V: " + remoteVersion)
				fmt.Println("found remote version: " + remoteVersion + ", class: " + class)
			}
		}
	})

	//grid := container.NewGridWithColumns(3)
	grid := container.New(layout.NewFormLayout())
	// iterate over the table data and create a new widget for each row
	for _, row := range tableData {

		title := row.Title
		titleLabel := widget.NewLabel(title)
		titleLabel.Wrapping = fyne.TextWrapWord

		remoteVersionLabel := widget.NewLabel("Newest V: ")
		currentVersionLabel := widget.NewLabel("Current V: ")

		titleLink := row.TitleLink
		titleButton := widget.NewButtonWithIcon("Update / Install", theme.DownloadIcon(), func() {
			fmt.Println("clicked Link")
			fmt.Println(titleLink)
		})

		row.Widgets.UpdateButton = titleButton
		row.Widgets.RemoteVersion = remoteVersionLabel
		row.Widgets.CurrentVersion = currentVersionLabel

		descriptionText := strings.ReplaceAll(row.Description, "</br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br>", "\n")
		descriptionText = strings.ReplaceAll(descriptionText, "<br/>", "\n")
		// remove html tags from description using regex
		re := regexp.MustCompile("<.*?>")
		descriptionText = re.ReplaceAllString(descriptionText, "")

		descriptionLabel := widget.NewLabel(descriptionText)
		descriptionLabel.Wrapping = fyne.TextWrapWord
		//authorLabel := widget.NewLabel(row.Author)

		//grid.Add(titleLabel)
		grid.Add(container.NewVBox(titleLabel, row.Widgets.UpdateButton, row.Widgets.RemoteVersion, row.Widgets.CurrentVersion))
		descriptionScroller := container.NewVScroll(descriptionLabel)
		descriptionScroller.Resize(fyne.NewSize(descriptionScroller.Size().Width, 400))
		grid.Add(descriptionScroller)

		grid.Add(widget.NewSeparator())
		grid.Add(widget.NewSeparator())
		//rowContainer := container.NewBorder(nil, nil, titleButton, authorLabel, descriptionLabel)

		//tableContainer.Add(rowContainer)
	}

	// Set the content of the window to the table container
	scrollContainer := container.NewVScroll(grid)
	verticalContent := container.NewBorder(checkAllButton, nil, nil, nil, scrollContainer)
	pluginListWindow.SetContent(verticalContent)

	// Show and run the application
	pluginListWindow.Show()
}

type TableDataWidgets struct {
	UpdateButton   *widget.Button
	CurrentVersion *widget.Label
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

func parseTableIntoStruct(table string) []TableData {
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

		description := strings.TrimSpace(cells[3])
		author := strings.TrimSpace(cells[4])

		row := TableData{
			Title:       strings.ReplaceAll(title, "**", ""),
			TitleLink:   titleLink + "/raw",
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
