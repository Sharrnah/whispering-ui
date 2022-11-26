package Messages

type InstalledLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

var InstalledLanguages []InstalledLanguage
