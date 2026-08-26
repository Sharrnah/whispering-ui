package ProfileForm

import "testing"

func TestTTSTypeOptionsIncludeAudio8(t *testing.T) {
	for _, option := range TTSTypeOptions() {
		if option.Value == "audio8_tts" {
			return
		}
	}
	t.Fatal("Audio8 TTS is missing from the profile TTS type options")
}
