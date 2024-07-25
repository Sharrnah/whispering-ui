package UpdateUtility

import (
	"bufio"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"whispering-tiger-ui/Utilities"
)

const PluginDir = "./Plugins/"

const PluginListUrl = "https://github.com/Sharrnah/whispering-plugins/raw/main/README.md"
const PluginRelativeUrlPrefix = "https://raw.githubusercontent.com/Sharrnah/whispering-plugins/main/"

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

func DownloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ExtractTable(md string) string {
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

func ParseTableIntoStruct(table string, relativeUrlPreviewPart string) []TableData {
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

// FetchAndAnalyzePluginUrl fetches the gist at the given URL and analyzes it for version and class information
// returns the version, class, hash and binary of the gist
func FetchAndAnalyzePluginUrl(url string) (string, string, string, []byte) {
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

type LocalPluginFilesData struct {
	Class        string
	FilePath     string
	LocalVersion string
	SHA256       string
}

func ParseLocalPluginFiles() []LocalPluginFilesData {
	var localPluginFiles []LocalPluginFilesData
	// build plugins list
	var pluginFiles []string
	files, err := os.ReadDir(PluginDir)
	if err != nil {
		println(err)
	}

	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "__init__") && (strings.HasSuffix(file.Name(), ".py")) {
			pluginFiles = append(pluginFiles, file.Name())
			pluginPath := PluginDir + file.Name()

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

func FindLocalPluginFileByClass(localPluginFiles []LocalPluginFilesData, class string) LocalPluginFilesData {
	for _, file := range localPluginFiles {
		if file.Class == class {
			return file
		}
	}
	return LocalPluginFilesData{}
}

type PluginUpdateInfo struct {
	remoteVersion string
	remoteHash    string
	localVersion  string
	localHash     string
	class         string
}

func PluginsUpdateCheck(pluginLink string, localPluginFilesData []LocalPluginFilesData) PluginUpdateInfo {
	remoteVersion, class, remoteHash, _ := FetchAndAnalyzePluginUrl(pluginLink)

	localPluginFile := FindLocalPluginFileByClass(localPluginFilesData, class)
	localVersion := localPluginFile.LocalVersion
	localHash := localPluginFile.SHA256

	return PluginUpdateInfo{
		class:         class,
		remoteVersion: remoteVersion,
		remoteHash:    remoteHash,
		localVersion:  localVersion,
		localHash:     localHash,
	}
}

func PluginsUpdateAvailable() bool {
	md, err := DownloadFile(PluginListUrl)
	if err != nil {
		print(err)
		return false
	}

	table := ExtractTable(md)
	tableData := ParseTableIntoStruct(table, PluginRelativeUrlPrefix)
	localPluginFilesData := ParseLocalPluginFiles()
	for _, row := range tableData {
		pluginLink := row.TitleLink
		pluginUpdateInfo := PluginsUpdateCheck(pluginLink, localPluginFilesData)
		if pluginUpdateInfo.remoteVersion != pluginUpdateInfo.localVersion && pluginUpdateInfo.localVersion != "" && pluginUpdateInfo.remoteVersion != "" {
			return true
		}
	}
	return false
}

func PluginsUpdateWidgetsRefresh(tableDataRow *TableData, localPluginFilesData []LocalPluginFilesData) {
	titleLink := tableDataRow.TitleLink
	fmt.Println("Checking update for: " + titleLink)
	if tableDataRow.Widgets.RemoteVersion != nil {
		pluginUpdateInfo := PluginsUpdateCheck(titleLink, localPluginFilesData)

		tableDataRow.Widgets.RemoteVersion.SetText(lang.L("Newest V") + ": " + pluginUpdateInfo.remoteVersion)
		//row.Widgets.CurrentVersion.SetText("Current V: " + localVersion)
		tableDataRow.Widgets.CurrentVersion.Text = "  " + lang.L("Current V") + ": " + pluginUpdateInfo.localVersion

		if pluginUpdateInfo.remoteVersion != pluginUpdateInfo.localVersion && pluginUpdateInfo.localVersion != "" {
			tableDataRow.Widgets.CurrentVersion.Color = color.RGBA{R: 240, G: 0, B: 0, A: 255}
			tableDataRow.Widgets.UpdateButton.Importance = widget.HighImportance
			tableDataRow.Widgets.UpdateButton.SetText(lang.L("Update"))
		} else {
			tableDataRow.Widgets.CurrentVersion.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			tableDataRow.Widgets.UpdateButton.Importance = widget.LowImportance
			if pluginUpdateInfo.remoteHash != pluginUpdateInfo.localHash {
				tableDataRow.Widgets.UpdateButton.Importance = widget.HighImportance
			}
			if (pluginUpdateInfo.localVersion == "" && pluginUpdateInfo.remoteVersion != "") || (pluginUpdateInfo.localVersion == "" && pluginUpdateInfo.remoteVersion == "") {
				tableDataRow.Widgets.UpdateButton.SetText(lang.L("Install"))
			} else {
				tableDataRow.Widgets.UpdateButton.SetText(lang.L("Reinstall"))
			}
		}
		tableDataRow.Widgets.CurrentVersion.Refresh()

		fmt.Println("found remote version: " + pluginUpdateInfo.remoteVersion + ", local version: " + pluginUpdateInfo.localVersion + " class: " + pluginUpdateInfo.class)
		fmt.Println("remote sha256: " + pluginUpdateInfo.remoteHash + ", local sha256: " + pluginUpdateInfo.localHash)
	}
}
