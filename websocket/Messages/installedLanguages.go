package Messages

import (
	"whispering-tiger-ui/Fields"
)

type InstalledLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type InstalledLanguagesListing struct {
	Languages []InstalledLanguage `json:"data"`
}

var InstalledLanguages InstalledLanguagesListing

func (res InstalledLanguagesListing) Update() *InstalledLanguagesListing {
	Fields.Field.TargetLanguageCombo.Options = nil
	Fields.Field.SourceLanguageCombo.Options = nil

	// Add None entry
	InstalledLanguages.Languages = append([]InstalledLanguage{
		{
			Code: "",
			Name: "None",
		},
	}, InstalledLanguages.Languages...)

	// fill source language combo
	for _, element := range InstalledLanguages.Languages {
		elementName := element.Name
		elementCode := element.Code
		if elementCode == "" {
			elementName = "Auto"
		}
		Fields.Field.SourceLanguageCombo.Options = append(Fields.Field.SourceLanguageCombo.Options, elementName)
	}
	// fill target language combo
	for _, element := range InstalledLanguages.Languages {
		Fields.Field.TargetLanguageCombo.Options = append(Fields.Field.TargetLanguageCombo.Options, element.Name)
	}

	return &res
}

func (res InstalledLanguagesListing) GetCodeByName(name string) string {
	for _, entry := range res.Languages {
		if entry.Name == name {
			return entry.Code
		}
	}
	return ""
}

func (res InstalledLanguagesListing) GetNameByCode(code string) string {
	for _, entry := range res.Languages {
		if entry.Code == code {
			return entry.Name
		}
	}
	return ""
}
