package Settings

// IsVibeVoiceStreaming also accepts profiles saved by the initial integration.
func IsVibeVoiceStreaming(modelType, modelName string) bool {
	if modelType == "vibevoice_asr_streaming" {
		return true
	}
	if modelType != "vibevoice_asr" {
		return false
	}
	switch modelName {
	case "VibeVoice-ASR-Streaming-1.5B", "VibeVoice-ASR-Streaming-7B", "custom-streaming":
		return true
	}
	return false
}

func CanonicalVibeVoiceSelection(modelType, modelName string) (string, string) {
	if modelType == "vibevoice_asr_streaming" {
		modelType = "vibevoice_asr"
		if modelName == "custom" {
			modelName = "custom-streaming"
		}
	}
	return modelType, modelName
}

// MainStreamingDisplayMode preserves the initial special-settings preference.
func MainStreamingDisplayMode() string {
	mode := Config.Streaming_display_mode
	if mode == "" {
		if group, ok := Config.Special_settings["stt_vibevoice_streaming"].(map[string]interface{}); ok {
			mode, _ = group["osc_mode"].(string)
		}
	}
	if mode == "rolling" || (mode == "configured" && Config.Osc_send_type == "rolling") {
		return "rolling"
	}
	return "blocks"
}
