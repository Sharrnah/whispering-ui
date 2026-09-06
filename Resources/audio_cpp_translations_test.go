package Resources

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudioCppTranslationKeysExistInEveryCatalog(t *testing.T) {
	required := []string{
		"CPU (most compatible, slower)",
		"CUDA (NVIDIA, recommended)",
		"Vulkan (AMD / Intel recommended; NVIDIA supported)",
		"Metal (Apple GPU)",
		"audio.cpp (native GGUF runtime)",
		"Qwen3-ASR 0.6B — recommended, faster and lower memory",
		"Qwen3-ASR 1.7B — higher accuracy, more memory",
		"Q8_0 GGUF — recommended, lower memory",
		"F16 GGUF — higher precision, more memory",
		"Original model dtype — recommended quality",
		"F16 GGUF — lower memory",
		"Fast multilingual preset voices — no voice cloning",
		"estimated OS order",
		"not detected by audio.cpp",
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
