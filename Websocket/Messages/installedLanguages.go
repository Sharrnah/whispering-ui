package Messages

import (
	"strings"
	"whispering-tiger-ui/CustomWidget"
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
	Fields.Field.TargetLanguageTxtTranslateCombo.Options = nil
	Fields.Field.SourceLanguageCombo.Options = nil
	Fields.Field.SourceLanguageCombo.OptionsTextValue = nil
	Fields.Field.SourceLanguageTxtTranslateCombo.Options = nil

	// fill language text translate target combo-boxes (without None value)
	for _, element := range InstalledLanguages.Languages {
		Fields.Field.TargetLanguageCombo.Options = append(Fields.Field.TargetLanguageCombo.Options, element.Name)
		Fields.Field.TargetLanguageCombo.OptionsTextValue = append(Fields.Field.TargetLanguageCombo.OptionsTextValue, CustomWidget.TextValueOption{
			Text:  element.Name,
			Value: element.Code,
		})
		Fields.Field.TargetLanguageTxtTranslateCombo.Options = append(Fields.Field.TargetLanguageTxtTranslateCombo.Options, element.Name)
		Fields.Field.TargetLanguageTxtTranslateCombo.OptionsTextValue = append(Fields.Field.TargetLanguageTxtTranslateCombo.OptionsTextValue, CustomWidget.TextValueOption{
			Text:  element.Name,
			Value: element.Code,
		})
	}
	Fields.Field.TargetLanguageCombo.ResetOptionsFilter()
	Fields.Field.TargetLanguageTxtTranslateCombo.ResetOptionsFilter()

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
		if elementCode == "" || elementCode == "auto" {
			elementName = "Auto"
		}
		Fields.Field.SourceLanguageCombo.Options = append(Fields.Field.SourceLanguageCombo.Options, elementName)
		Fields.Field.SourceLanguageCombo.OptionsTextValue = append(Fields.Field.SourceLanguageCombo.OptionsTextValue, CustomWidget.TextValueOption{
			Text:  elementName,
			Value: elementCode,
		})
		Fields.Field.SourceLanguageTxtTranslateCombo.Options = append(Fields.Field.SourceLanguageTxtTranslateCombo.Options, elementName)
		Fields.Field.SourceLanguageTxtTranslateCombo.OptionsTextValue = append(Fields.Field.SourceLanguageTxtTranslateCombo.OptionsTextValue, CustomWidget.TextValueOption{
			Text:  elementName,
			Value: elementCode,
		})
		Fields.Field.SourceLanguageCombo.ResetOptionsFilter()
		Fields.Field.SourceLanguageTxtTranslateCombo.ResetOptionsFilter()
	}

	return &res
}

func (res InstalledLanguagesListing) GetCodeByName(name string) string {
	for _, entry := range res.Languages {
		if strings.ToLower(entry.Name) == strings.ToLower(name) {
			return entry.Code
		}
	}
	return ""
}

func (res InstalledLanguagesListing) GetNameByCode(code string) string {
	if strings.ToLower(code) == "auto" {
		return "Auto"
	}
	for _, entry := range res.Languages {
		if strings.ToLower(entry.Code) == strings.ToLower(code) {
			return entry.Name
		}
	}
	return ""
}
