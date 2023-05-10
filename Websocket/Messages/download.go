package Messages

import "whispering-tiger-ui/ModelDownloader"

type Download struct {
	Urls       []string `json:"urls"`
	ExtractDir string   `json:"extract_dir"`
	Checksum   string   `json:"checksum"`
}

type DownloadMessage struct {
	Type     string   `json:"type"`
	Download Download `json:"data"`
}

func (res DownloadMessage) StartDownload() error {
	dl := res.Download
	return ModelDownloader.DownloadFile(dl.Urls, dl.ExtractDir, dl.Checksum)
}
