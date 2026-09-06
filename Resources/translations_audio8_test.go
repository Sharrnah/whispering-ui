package Resources

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudio8TranslationKeysExistInEveryCatalog(t *testing.T) {
	required := []string{
		"Audio8 TTS Preview (Voice Cloning)",
		"Attention",
		"Auto (FlashAttention 2 when supported)",
		"Auto (clone with exact transcript)",
		"Automatic (BF16 on supported CUDA)",
		"BFloat16",
		"Clone Mode",
		"Compatible PyTorch attention",
		"Disabled (unconditioned voice)",
		"Enable",
		"Exact transcript of the selected reference audio (optional in Auto mode)",
		"Fallback Segment Characters",
		"Fallback Segment Pause (ms)",
		"Float32",
		"FlashAttention 2",
		"General",
		"Generation",
		"Max new tokens",
		"Precision",
		"Reference Transcript",
		"Required (exact transcript)",
		"Reset",
		"Reset to defaults",
		"Reuse Generated Voice",
		"Sampling",
		"Save last generation as clone reference",
		"Seed",
		"Streaming",
		"Temperature",
		"Top K",
		"Top P",
		"Voice Cloning",
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
