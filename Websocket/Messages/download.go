package Messages

import "whispering-tiger-ui/ModelDownloader"

type Download struct {
	Urls          []string `json:"urls"`
	ExtractDir    string   `json:"extract_dir"`
	Checksum      string   `json:"checksum"`
	Title         string   `json:"title"`
	ExtractFormat string   `json:"extract_format"`
}

type DownloadMessage struct {
	Type     string   `json:"type"`
	Download Download `json:"data"`
}

func (res DownloadMessage) StartDownload() error {
	dl := res.Download
	return ModelDownloader.DownloadFile(dl.Urls, dl.ExtractDir, dl.Checksum, dl.Title, dl.ExtractFormat)
}
