package Messages

import (
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

//goland:noinspection SpellCheckingInspection
type Window struct {
	Hwnd  string `json:"hwnd"`
	Title string `json:"title"`
}

type WindowsStruct struct {
	Windows []Window `json:"data"`
}

var WindowsList WindowsStruct

func (res WindowsStruct) Update() *WindowsStruct {
	// fill combo-box with ocr windows
	Fields.Field.OcrWindowCombo.Options = nil
	for _, window := range WindowsList.Windows {
		Fields.Field.OcrWindowCombo.Options = append(Fields.Field.OcrWindowCombo.Options, window.Title)
	}

	if Fields.Field.OcrWindowCombo.Selected != Settings.Config.Ocr_window_name {
		if !Utilities.Contains(Fields.Field.OcrWindowCombo.Options, Settings.Config.Ocr_window_name) {
			Fields.Field.OcrWindowCombo.Options = append(Fields.Field.OcrWindowCombo.Options, Settings.Config.Ocr_window_name)
		}
	}

	if Fields.Field.OcrWindowCombo.LastTappedPointEvent != nil {
		Fields.Field.OcrWindowCombo.ShopPopup()
	}

	return &res
}

// ############################

type OcrLanguages struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
type InstalledOcrLanguagesListing struct {
	Languages []OcrLanguages `json:"data"`
}

var OcrLanguagesList InstalledOcrLanguagesListing

func (res InstalledOcrLanguagesListing) Update() *InstalledOcrLanguagesListing {
	// fill combo-box with ocr windows
	Fields.Field.OcrLanguageCombo.Options = nil
	for _, language := range res.Languages {
		Fields.Field.OcrLanguageCombo.Options = append(Fields.Field.OcrLanguageCombo.Options, language.Name)
	}

	return &res
}

func (res InstalledOcrLanguagesListing) GetCodeByName(name string) string {
	for _, entry := range res.Languages {
		if entry.Name == name {
			return entry.Code
		}
	}
	return ""
}

func (res InstalledOcrLanguagesListing) GetNameByCode(code string) string {
	for _, entry := range res.Languages {
		if entry.Code == code {
			return entry.Name
		}
	}
	return ""
}
