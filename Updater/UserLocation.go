package Updater

import (
	"fmt"
	goLocale "github.com/jeandeaual/go-locale"
	"os"
	"strings"
)

func GetLanguage() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LANGUAGE")
	}
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		locales, err := goLocale.GetLocales()
		if err != nil {
			fmt.Println("Error getting locales:", err)
			return ""
		}
		if len(locales) == 0 {
			fmt.Println("No locale found.")
			return ""
		}
		lang = locales[0]
	}
	// make sure we have a valid locale in format xx_XX
	lang = strings.ReplaceAll(lang, "-", "_")
	return lang
}

func IsUSLocale(lang string) bool {
	// List of common US-specific locales
	usLocales := []string{"en_US", "es_US", "zh_US"}

	for _, locale := range usLocales {
		if strings.HasPrefix(lang, locale) {
			return true
		}
	}
	return false
}

func IsEULocale(lang string) bool {
	// List of common EU country codes
	euCountryCodes := []string{"AT", "BE", "BG", "CY", "CZ", "DE", "DK", "EE", "ES", "FI", "FR", "GR", "HR", "HU", "IE", "IT", "LT", "LU", "LV", "MT", "NL", "PL", "PT", "RO", "SE", "SI", "SK"}

	for _, countryCode := range euCountryCodes {
		if strings.Contains(lang, "_"+countryCode) {
			return true
		}
	}
	return false
}
