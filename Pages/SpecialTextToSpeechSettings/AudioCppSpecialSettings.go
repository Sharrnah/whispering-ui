package SpecialTextToSpeechSettings

import (
	"fmt"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

const audioCppTTSSettingsGroup = "tts_audio_cpp"

type audioCppTTSFieldKind uint8

const (
	audioCppTTSSelect audioCppTTSFieldKind = iota
	audioCppTTSSearchSelect
	audioCppTTSSlider
	audioCppTTSCheck
	audioCppTTSLongText
	audioCppTTSAudioFile
)

type audioCppTTSField struct {
	key         string
	label       string
	fallback    interface{}
	kind        audioCppTTSFieldKind
	options     []CustomWidget.TextValueOption
	allowCustom bool
	min         float64
	max         float64
	step        float64
	decimals    int
}

type audioCppTTSGroup struct {
	title   string
	columns int
	fields  []audioCppTTSField
}

type audioCppTTSSchema struct {
	family      string
	title       string
	description string
	groups      []audioCppTTSGroup
}

var audioCppSupertonicLanguages = []string{"auto", "en", "ko", "ja", "ar", "bg", "cs", "da", "de", "el", "es", "et", "fi", "fr", "hi", "hr", "hu", "id", "it", "lt", "lv", "nl", "pl", "pt", "ro", "ru", "sk", "sl", "sv", "tr", "uk", "vi"}
var audioCppConfuciusLanguages = []string{"auto", "zh", "en", "ja", "ko", "de", "fr", "es", "id", "it", "th", "pt", "ru", "ms", "vi"}
var audioCppMagpieLanguages = []string{"auto", "ar-AE", "ar-MSA", "ar-SA", "de", "en", "es", "fr", "hi", "it", "ko", "pt-BR", "vi", "zh"}
var audioCppCommonLanguages = []string{
	"auto", "ar", "bg", "cs", "da", "de", "el", "en", "es", "et", "fi", "fr", "he", "hi", "hr", "hu",
	"id", "it", "ja", "km", "ko", "lo", "lt", "lv", "ms", "my", "nl", "no", "pl", "pt", "ro", "ru", "sk",
	"sl", "sv", "sw", "th", "tl", "tr", "uk", "vi", "yue", "zh",
}

var audioCppLanguageNames = map[string]string{
	"ar": "Arabic", "ar-ae": "Arabic (UAE)", "ar-msa": "Arabic (Modern Standard)", "ar-sa": "Arabic (Saudi Arabia)",
	"bg": "Bulgarian", "cs": "Czech", "da": "Danish", "de": "German", "el": "Greek", "en": "English",
	"es": "Spanish", "et": "Estonian", "fi": "Finnish", "fr": "French", "he": "Hebrew", "hi": "Hindi",
	"hr": "Croatian", "hu": "Hungarian", "id": "Indonesian", "it": "Italian", "ja": "Japanese", "km": "Khmer",
	"ko": "Korean", "lo": "Lao", "lt": "Lithuanian", "lv": "Latvian", "ms": "Malay", "my": "Burmese",
	"nl": "Dutch", "no": "Norwegian", "pl": "Polish", "pt": "Portuguese", "pt-br": "Portuguese (Brazil)",
	"ro": "Romanian", "ru": "Russian", "sk": "Slovak", "sl": "Slovenian", "sv": "Swedish", "sw": "Swahili",
	"th": "Thai", "tl": "Filipino", "tr": "Turkish", "uk": "Ukrainian", "vi": "Vietnamese", "yue": "Cantonese",
	"zh": "Chinese",
}

func audioCppTTSOptions(values ...string) []CustomWidget.TextValueOption {
	options := make([]CustomWidget.TextValueOption, 0, len(values))
	for _, value := range values {
		options = append(options, CustomWidget.TextValueOption{Text: value, Value: value})
	}
	return options
}

func audioCppTTSLanguageOptions(codes []string) []CustomWidget.TextValueOption {
	options := make([]CustomWidget.TextValueOption, 0, len(codes))
	for _, code := range codes {
		text := code
		switch strings.ToLower(code) {
		case "auto":
			text = "Automatic"
		case "none":
			text = "No language tag"
		default:
			if name := audioCppLanguageNames[strings.ToLower(code)]; name != "" {
				text = fmt.Sprintf("%s (%s)", name, code)
			}
		}
		options = append(options, CustomWidget.TextValueOption{Text: text, Value: code})
	}
	return options
}

func audioCppTTSChoice(key, label, fallback string, options ...CustomWidget.TextValueOption) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: fallback, kind: audioCppTTSSelect, options: options}
}

func audioCppTTSLanguage(key, fallback string, codes []string, allowCustom bool) audioCppTTSField {
	return audioCppTTSField{
		key: key, label: "Language", fallback: fallback, kind: audioCppTTSSearchSelect,
		options: audioCppTTSLanguageOptions(codes), allowCustom: allowCustom,
	}
}

func audioCppTTSFloatSlider(key, label string, fallback, min, max, step float64, decimals int) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: fallback, kind: audioCppTTSSlider, min: min, max: max, step: step, decimals: decimals}
}

func audioCppTTSIntSlider(key, label string, fallback int, min, max, step float64) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: fallback, kind: audioCppTTSSlider, min: min, max: max, step: step}
}

func audioCppTTSCheckbox(key, label string, fallback bool) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: fallback, kind: audioCppTTSCheck}
}

func audioCppTTSParagraph(key, label, fallback string) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: fallback, kind: audioCppTTSLongText}
}

func audioCppTTSAudioPath(key, label string) audioCppTTSField {
	return audioCppTTSField{key: key, label: label, fallback: "", kind: audioCppTTSAudioFile}
}

func audioCppTTSSchemaForModel(model string) (audioCppTTSSchema, bool) {
	if schema, ok := additionalAudioCppTTSSchema(model); ok {
		return schema, true
	}
	switch model {
	case "Supertonic-3-GGUF":
		return audioCppTTSSchema{
			family: "supertonic", title: "Supertonic 3",
			description: "Fast multilingual synthesis with ten built-in voices. Use the normal Voice and Rate controls for voice and speaking speed.",
			groups: []audioCppTTSGroup{{title: "Language and quality", columns: 2, fields: []audioCppTTSField{
				audioCppTTSLanguage("language", "auto", audioCppSupertonicLanguages, false),
				audioCppTTSIntSlider("num_inference_steps", "Quality steps", 8, 1, 50, 1),
			}}},
		}, true
	case "Confucius4-TTS-GGUF":
		return audioCppTTSSchema{
			family: "confucius4_tts", title: "Confucius4-TTS",
			description: "Experimental multilingual voice cloning. Select a reference with the normal Voice control.",
			groups: []audioCppTTSGroup{{title: "Language and generation", columns: 2, fields: []audioCppTTSField{
				audioCppTTSLanguage("language", "auto", audioCppConfuciusLanguages, false),
				audioCppTTSIntSlider("num_beams", "Beam count", 3, 1, 8, 1),
				audioCppTTSIntSlider("num_inference_steps", "Flow-matching steps", 25, 1, 50, 1),
				audioCppTTSFloatSlider("guidance_scale", "Guidance", 0.7, 0, 2, 0.05, 2),
				audioCppTTSFloatSlider("temperature", "Creativity", 0.8, 0.1, 2, 0.05, 2),
			}}},
		}, true
	case "DotTTS-SOAR-GGUF", "DotTTS-MeanFlow-GGUF":
		variant := "SOAR"
		if model == "DotTTS-MeanFlow-GGUF" {
			variant = "MeanFlow"
		}
		languageCodes := append([]string{"none"}, audioCppCommonLanguages[1:]...)
		return audioCppTTSSchema{
			family: "dots_tts", title: "DotTTS " + variant,
			description: "Multilingual synthesis, instruction control, and zero-shot cloning. A same-name .txt file beside the selected voice can supply its exact transcript.",
			groups: []audioCppTTSGroup{
				{title: "Task", columns: 2, fields: []audioCppTTSField{
					audioCppTTSLanguage("language", "none", languageCodes, true),
					audioCppTTSChoice("template_name", "Synthesis mode", "tts",
						CustomWidget.TextValueOption{Text: "Text to speech", Value: "tts"},
						CustomWidget.TextValueOption{Text: "Instruction-guided speech", Value: "instruction_tts"},
						CustomWidget.TextValueOption{Text: "Text to audio", Value: "text_to_audio"},
						CustomWidget.TextValueOption{Text: "Streaming interleave", Value: "tts_interleave"},
					),
				}},
				{title: "Voice and instruction", columns: 2, fields: []audioCppTTSField{
					audioCppTTSParagraph("reference_text", "Exact reference transcript", ""),
					audioCppTTSParagraph("instruction", "Voice / audio instruction", ""),
				}},
				{title: "Quality", columns: 2, fields: []audioCppTTSField{
					audioCppTTSIntSlider("num_inference_steps", "Flow steps", 10, 1, 50, 1),
					audioCppTTSFloatSlider("guidance_scale", "Guidance", 1.2, 0, 4, 0.05, 2),
					audioCppTTSFloatSlider("speaker_scale", "Speaker likeness", 1.5, 0, 3, 0.05, 2),
					audioCppTTSChoice("sampler_mode", "Flow solver", "euler", audioCppTTSOptions("euler", "midpoint", "rk4")...),
				}},
			},
		}, true
	case "DotTTS-Edit-GGUF":
		languageCodes := append([]string{"none"}, audioCppCommonLanguages[1:]...)
		return audioCppTTSSchema{
			family: "dots_tts_edit", title: "DotTTS Edit",
			description: "Edits an existing recording. Select the source audio and describe the edit; source and target transcripts are optional precision controls.",
			groups: []audioCppTTSGroup{
				{title: "Edit input", columns: 1, fields: []audioCppTTSField{
					audioCppTTSAudioPath("source_audio", "Source audio"),
				}},
				{title: "Edit text", columns: 2, fields: []audioCppTTSField{
					audioCppTTSParagraph("source_text", "Source transcript (optional)", ""),
					audioCppTTSParagraph("target_text", "Target transcript (optional)", ""),
					audioCppTTSParagraph("instruction", "Edit instruction", ""),
				}},
				{title: "Voice and quality", columns: 2, fields: []audioCppTTSField{
					audioCppTTSLanguage("language", "none", languageCodes, true),
					audioCppTTSChoice("use_xvector", "Use source speaker", "auto",
						CustomWidget.TextValueOption{Text: "Automatic", Value: "auto"},
						CustomWidget.TextValueOption{Text: "Always", Value: "on"},
						CustomWidget.TextValueOption{Text: "Never", Value: "off"},
					),
					audioCppTTSIntSlider("num_inference_steps", "Flow steps", 10, 1, 50, 1),
					audioCppTTSFloatSlider("guidance_scale", "Guidance", 1.2, 0, 4, 0.05, 2),
					audioCppTTSFloatSlider("speaker_scale", "Speaker likeness", 1.5, 0, 3, 0.05, 2),
					audioCppTTSChoice("sampler_mode", "Flow solver", "euler", audioCppTTSOptions("euler", "midpoint", "rk4")...),
				}},
			},
		}, true
	case "IndexTTS2-GGUF", "IndexTTS2.5-GGUF":
		version := "2"
		languages := []string{"auto", "zh", "en"}
		if model == "IndexTTS2.5-GGUF" {
			version = "2.5"
			languages = []string{"auto", "zh", "en", "ja", "es", "ar"}
		}
		return audioCppTTSSchema{
			family: "index_tts2", title: "IndexTTS " + version,
			description: "Voice cloning with reference-audio, text-derived, random, or manually mixed emotion. Select the cloning voice with the normal Voice control.",
			groups: []audioCppTTSGroup{
				{title: "Voice and pacing", columns: 2, fields: []audioCppTTSField{
					audioCppTTSLanguage("language", "auto", languages, false),
					audioCppTTSFloatSlider("duration_factor", "Duration factor", 1.0, 0.5, 2, 0.05, 2),
				}},
				{title: "Emotion source", columns: 2, fields: []audioCppTTSField{
					audioCppTTSAudioPath("emotion_audio", "Emotion reference audio"),
					audioCppTTSParagraph("emotion_text", "Emotion description", ""),
					audioCppTTSCheckbox("use_emotion_text", "Use emotion description", false),
					audioCppTTSCheckbox("use_random_emotion", "Generate random emotion", false),
					audioCppTTSFloatSlider("emotion_alpha", "Emotion strength", 1.0, 0, 1, 0.01, 2),
				}},
				{title: "Manual emotion mix", columns: 2, fields: []audioCppTTSField{
					audioCppTTSFloatSlider("emotion_happy", "Happy", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_angry", "Angry", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_sad", "Sad", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_afraid", "Afraid", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_disgusted", "Disgusted", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_melancholic", "Melancholic", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_surprised", "Surprised", 0, 0, 1, 0.01, 2),
					audioCppTTSFloatSlider("emotion_calm", "Calm", 0, 0, 1, 0.01, 2),
				}},
			},
		}, true
	case "MagpieTTS-Multilingual-357M-GGUF":
		return audioCppTTSSchema{
			family: "magpie_tts", title: "MagpieTTS Multilingual 357M",
			description: "Small multilingual model with five preset voices. Select the voice with the normal Voice control.",
			groups: []audioCppTTSGroup{{title: "Language and expression", columns: 2, fields: []audioCppTTSField{
				audioCppTTSLanguage("language", "auto", audioCppMagpieLanguages, false),
				audioCppTTSFloatSlider("temperature", "Creativity", 0.6, 0.1, 2, 0.05, 2),
				audioCppTTSIntSlider("top_k", "Candidate tokens", 80, 1, 200, 1),
				audioCppTTSFloatSlider("guidance_scale", "Guidance", 2.5, 0, 5, 0.05, 2),
			}}},
		}, true
	case "OmniVoice-GGUF":
		return audioCppTTSSchema{
			family: "omnivoice", title: "OmniVoice",
			description: "Massively multilingual voice cloning and text-based voice design. The searchable language field accepts any audio.cpp language name or code, including values outside the suggestions.",
			groups: []audioCppTTSGroup{
				{title: "Voice and language", columns: 2, fields: []audioCppTTSField{
					audioCppTTSLanguage("language", "auto", audioCppCommonLanguages, true),
					audioCppTTSCheckbox("denoise", "Denoise reference", true),
					audioCppTTSParagraph("reference_text", "Exact reference transcript", ""),
					audioCppTTSParagraph("voice_instruction", "Voice design instruction", ""),
				}},
				{title: "Quality and timing", columns: 2, fields: []audioCppTTSField{
					audioCppTTSIntSlider("num_inference_steps", "Generation steps", 32, 1, 64, 1),
					audioCppTTSFloatSlider("guidance_scale", "Guidance", 2.0, 0, 5, 0.05, 2),
					audioCppTTSFloatSlider("speed", "Speed", 1.0, 0.5, 2, 0.05, 2),
					audioCppTTSFloatSlider("duration", "Target duration (0 = automatic)", 0, 0, 60, 0.5, 1),
				}},
			},
		}, true
	case "VoxCPM1-0.5B-GGUF", "VoxCPM2-GGUF":
		family, title, description := "voxcpm1", "VoxCPM 1 (0.5B)", "Compact text-to-speech and short-reference voice cloning."
		voiceFields := []audioCppTTSField{audioCppTTSParagraph("reference_text", "Exact reference transcript", "")}
		if model == "VoxCPM2-GGUF" {
			family, title = "voxcpm2", "VoxCPM 2"
			description = "Higher-quality multilingual cloning and natural-language voice design."
			voiceFields = append(voiceFields, audioCppTTSParagraph("voice_instruction", "Voice design instruction", ""))
		}
		return audioCppTTSSchema{
			family: family, title: title, description: description,
			groups: []audioCppTTSGroup{
				{title: "Voice", columns: 2, fields: voiceFields},
				{title: "Quality", columns: 2, fields: []audioCppTTSField{
					audioCppTTSIntSlider("num_inference_steps", "Diffusion steps", 10, 1, 50, 1),
					audioCppTTSFloatSlider("guidance_scale", "Guidance", 2.0, 0, 5, 0.05, 2),
				}},
			},
		}, true
	default:
		return audioCppTTSSchema{}, false
	}
}

func audioCppTTSCanonicalModel(displayValue string) string {
	if canonical, ok := Fields.TtsModelSelectionValues[displayValue]; ok && len(canonical) >= 2 {
		return canonical[1]
	}
	value := strings.TrimSpace(displayValue)
	if index := strings.Index(value, " ("); index > 0 {
		value = value[:index]
	}
	return value
}

func selectedAudioCppTTSModel() string {
	if Fields.Field.TtsModelCombo != nil && Fields.Field.TtsModelCombo.Selected != "" {
		if model := audioCppTTSCanonicalModel(Fields.Field.TtsModelCombo.Selected); model != "" {
			return model
		}
	}
	if len(Settings.Config.Tts_model) >= 2 {
		return Settings.Config.Tts_model[1]
	}
	return "Supertonic-3-GGUF"
}

func audioCppTTSOptionValue(options []CustomWidget.TextValueOption, text string) (string, bool) {
	for _, option := range options {
		if strings.EqualFold(strings.TrimSpace(text), option.Text) || strings.EqualFold(strings.TrimSpace(text), option.Value) {
			return option.Value, true
		}
	}
	return "", false
}

func audioCppTTSNumber(value interface{}) float64 {
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case float64:
		return typed
	default:
		return 0
	}
}

func audioCppTTSControl(schema audioCppTTSSchema, field audioCppTTSField) fyne.CanvasObject {
	current := GetNestedSpecialSettingFallback(audioCppTTSSettingsGroup, schema.family, field.key, field.fallback)
	switch field.kind {
	case audioCppTTSSelect:
		selector := CustomWidget.NewTextValueSelect("audio_cpp_tts_"+schema.family+"_"+field.key, field.options, nil, 0)
		selector.SetSelected(fmt.Sprint(current))
		selector.OnChanged = func(option CustomWidget.TextValueOption) {
			UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, option.Value, false)
		}
		return selector
	case audioCppTTSSearchSelect:
		entry := CustomWidget.NewCompletionEntry(nil)
		entry.SetValueOptions(field.options)
		entry.ResetOptionsFilter()
		entry.PlaceHolder = "Type to filter languages"
		value := fmt.Sprint(current)
		entry.SetSelected(value)
		if entry.Text == "" && value != "" {
			entry.SetText(value)
		}
		filterChanged := entry.OnChanged
		entry.OnChanged = func(text string) {
			if filterChanged != nil {
				filterChanged(text)
			}
			if canonical, ok := audioCppTTSOptionValue(field.options, text); ok {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, canonical, false)
			} else if field.allowCustom {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, strings.TrimSpace(text), true)
			}
		}
		entry.OnSubmitted = func(text string) {
			if canonical, ok := audioCppTTSOptionValue(field.options, text); ok {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, canonical, false)
			} else if field.allowCustom {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, strings.TrimSpace(text), false)
			}
		}
		return entry
	case audioCppTTSSlider:
		value := math.Max(field.min, math.Min(field.max, audioCppTTSNumber(current)))
		slider := widget.NewSlider(field.min, field.max)
		slider.Step = field.step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(fmt.Sprintf("%.*f", field.decimals, value))
		valueLabel.Alignment = fyne.TextAlignTrailing
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.*f", field.decimals, value))
			if _, integer := field.fallback.(int); integer {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, int(math.Round(value)), false)
			} else {
				UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, value, false)
			}
		}
		return container.NewBorder(nil, nil, nil, valueLabel, slider)
	case audioCppTTSCheck:
		check := widget.NewCheck("", nil)
		check.SetChecked(current.(bool))
		check.OnChanged = func(value bool) {
			UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, value, false)
		}
		return check
	case audioCppTTSLongText:
		entry := widget.NewMultiLineEntry()
		entry.SetMinRowsVisible(2)
		entry.SetPlaceHolder(field.label)
		entry.SetText(current.(string))
		entry.OnChanged = func(value string) {
			UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, value, true)
		}
		return entry
	case audioCppTTSAudioFile:
		entry := widget.NewEntry()
		entry.SetPlaceHolder("Select WAV, MP3, FLAC, or OGG audio")
		entry.SetText(current.(string))
		entry.OnChanged = func(value string) {
			UpdateNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family, field.key, value, true)
		}
		browse := widget.NewButton("Browse…", func() {
			app := fyne.CurrentApp()
			if app == nil || len(app.Driver().AllWindows()) == 0 {
				return
			}
			picker := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				path := reader.URI().Path()
				_ = reader.Close()
				entry.SetText(path)
			}, app.Driver().AllWindows()[0])
			picker.SetFilter(storage.NewExtensionFileFilter([]string{".wav", ".mp3", ".flac", ".ogg"}))
			picker.Show()
		})
		return container.NewBorder(nil, nil, nil, browse, entry)
	default:
		return widget.NewLabel("")
	}
}

func audioCppTTSForm(schema audioCppTTSSchema, fields []audioCppTTSField) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(fields)*2)
	for _, field := range fields {
		objects = append(objects, widget.NewLabel(field.label+":"), audioCppTTSControl(schema, field))
	}
	return container.New(layout.NewFormLayout(), objects...)
}

func audioCppTTSGroupContent(schema audioCppTTSSchema, group audioCppTTSGroup) fyne.CanvasObject {
	if group.columns < 2 || len(group.fields) < 2 {
		return audioCppTTSForm(schema, group.fields)
	}
	middle := (len(group.fields) + 1) / 2
	return container.NewGridWithColumns(2,
		audioCppTTSForm(schema, group.fields[:middle]),
		audioCppTTSForm(schema, group.fields[middle:]),
	)
}

func buildAudioCppTTSSettingsForModel(model string, rebuild func(string)) fyne.CanvasObject {
	schema, ok := audioCppTTSSchemaForModel(model)
	if !ok {
		return widget.NewLabel("No audio.cpp settings schema is available for " + model + ".")
	}
	items := make([]*widget.AccordionItem, 0, len(schema.groups))
	for _, group := range schema.groups {
		items = append(items, widget.NewAccordionItem(group.title, audioCppTTSGroupContent(schema, group)))
	}
	accordion := widget.NewAccordion(items...)
	if len(items) > 0 {
		accordion.Open(0)
	}
	reset := widget.NewButton("Reset", func() {
		ResetNestedSpecialSettings(audioCppTTSSettingsGroup, schema.family)
		if rebuild != nil {
			rebuild(model)
		}
	})
	reset.Importance = widget.LowImportance
	return container.NewVBox(
		accordion,
		container.NewHBox(layout.NewSpacer(), reset),
	)
}

// BuildAudioCppSpecialSettings shows a curated set of user-facing controls.
// Low-level graph arenas, tensor reinterpretation, cache sizes, and token caps
// remain automatic because the packaged GGUF and audio.cpp model defaults own
// those implementation details.
func BuildAudioCppSpecialSettings() fyne.CanvasObject {
	host := container.NewStack()
	var rebuild func(string)
	rebuild = func(displayValue string) {
		model := audioCppTTSCanonicalModel(displayValue)
		if model == "" {
			model = selectedAudioCppTTSModel()
		}
		host.Objects = []fyne.CanvasObject{buildAudioCppTTSSettingsForModel(model, rebuild)}
		host.Refresh()
	}
	Fields.TtsModelSelectionChanged = rebuild
	rebuild(selectedAudioCppTTSModel())
	return host
}
