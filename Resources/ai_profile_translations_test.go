package Resources

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIProfileTranslationKeysExistInEveryCatalog(t *testing.T) {
	required := []string{
		"GPU",
		"Float16 request (safe BF16/FP32 fallback)",
		"Text-to-Speech Model",
	}
	entries, err := Translations.ReadDir("translations")
	if err != nil {
		t.Fatal(err)
	}
	catalogs := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		catalogs++
		data, err := Translations.ReadFile("translations/" + entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		catalog := make(map[string]json.RawMessage)
		if err := json.Unmarshal(data, &catalog); err != nil {
			t.Fatalf("invalid translation catalog %s: %v", entry.Name(), err)
		}
		for _, key := range required {
			if _, ok := catalog[key]; !ok {
				t.Errorf("translation catalog %s is missing %q", entry.Name(), key)
			}
		}
	}
	if catalogs != 16 {
		t.Fatalf("checked %d translation catalogs, expected 16", catalogs)
	}
}
