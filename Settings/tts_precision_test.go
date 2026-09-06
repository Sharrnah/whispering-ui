package Settings

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTTSProfile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "profile.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadYamlSettingsMigratesLegacyTTSPrecision(t *testing.T) {
	path := writeTTSProfile(t, "tts_type: qwen3_tts\nspecial_settings:\n  tts_qwen3_tts:\n    precision: bfloat16\n")
	conf := Conf{}
	if err := conf.LoadYamlSettings(path); err != nil {
		t.Fatal(err)
	}
	if conf.Tts_precision != "bfloat16" {
		t.Fatalf("tts_precision = %q, want bfloat16", conf.Tts_precision)
	}
}

func TestTopLevelTTSPrecisionWinsAndSynchronizesLegacyField(t *testing.T) {
	path := writeTTSProfile(t, "tts_type: qwen3_tts\ntts_precision: float32\nspecial_settings:\n  tts_qwen3_tts:\n    precision: bfloat16\n")
	conf := Conf{}
	if err := conf.LoadYamlSettings(path); err != nil {
		t.Fatal(err)
	}
	if conf.Tts_precision != "float32" {
		t.Fatalf("tts_precision = %q, want float32", conf.Tts_precision)
	}
	qwenSettings := conf.Special_settings["tts_qwen3_tts"].(map[string]interface{})
	if qwenSettings["precision"] != "float32" {
		t.Fatalf("legacy precision = %#v, want float32", qwenSettings["precision"])
	}
}
