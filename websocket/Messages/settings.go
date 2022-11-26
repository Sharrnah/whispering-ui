package Messages

type WhisperLanguage struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type TranslateSetting struct {
	WhisperLanguages      []WhisperLanguage `json:"whisper_languages"`
	SelectedOcrWindowName string            `json:"ocr_window_name"`
}

var TranslateSettings TranslateSetting
