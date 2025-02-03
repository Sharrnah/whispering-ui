package Fields

import (
	"strconv"
	"strings"
	"whispering-tiger-ui/Settings"
)

func AdditionalLanguagesCountString(prefix string, wrapping string) string {
	// count additional languages
	numOfAdditionalLanguages := 0
	numOfAdditionalLanguagesLabelText := ""
	for _, language := range strings.Split(Settings.Config.Txt_second_translation_languages, ",") {
		if language != "" {
			numOfAdditionalLanguages++
		}
	}
	if Settings.Config.Txt_second_translation_enabled && numOfAdditionalLanguages > 0 {
		if prefix != "" {
			numOfAdditionalLanguagesLabelText = prefix
		}
		// Split the wrapping into beginning and end characters
		wrappingStart := ""
		wrappingEnd := ""
		if len(wrapping) > 0 {
			wrappingStart = string(wrapping[0])
		}
		if len(wrapping) > 1 {
			wrappingEnd = string(wrapping[1])
		}
		numOfAdditionalLanguagesLabelText += wrappingStart + "+" + strconv.Itoa(numOfAdditionalLanguages) + wrappingEnd
	}
	return numOfAdditionalLanguagesLabelText
}
