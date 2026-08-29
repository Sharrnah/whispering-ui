package Resources

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAdditionalAudioRouteStringsExistInEveryLocale(t *testing.T) {
	required := []string{
		"Additional Audio Sources",
		"Add Audio Source",
		"Additional Audio Sources Share the Loaded Speech Model",
		"Apply Audio Sources",
		"Audio Processing",
		"Audio Source",
		"Audio Source Name",
		"Audio Source Name Is Required",
		"Main Microphone",
		"Microphone Plugin Routing",
		"New Audio Sources Copy Current Profile Defaults",
		"No Additional Audio Sources Configured",
		"Plugin Routing",
		"Plugin Routing Chooses Audio Sources for Enabled Plugins",
		"DeepFilterNet",
		"Noise Reduce",
		"Remove Audio Source",
		"Show Results in Whispering Tiger",
		"Smart Turn Detection",
		"Task",
		"Turn Detection Minimum Length",
		"Turn Pause Length",
		"Turn Probability Threshold",
		"VRChat Notification Sound",
		"VRChat Typing Indicator",
	}
	files, err := filepath.Glob(filepath.Join("translations", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no translation files found")
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		translations := map[string]json.RawMessage{}
		if err := json.Unmarshal(content, &translations); err != nil {
			t.Fatalf("%s: invalid JSON: %v", file, err)
		}
		for _, key := range required {
			if _, ok := translations[key]; !ok {
				t.Errorf("%s: missing translation key %q", file, key)
			}
		}
	}
}
