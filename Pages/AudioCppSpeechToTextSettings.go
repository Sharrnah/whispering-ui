package Pages

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"whispering-tiger-ui/CustomWidget"
	special "whispering-tiger-ui/Pages/SpecialTextToSpeechSettings"
	"whispering-tiger-ui/Settings"
)

const audioCppSTTSettingsGroup = "stt_audio_cpp"

type audioCppSTTFieldKind uint8

const (
	audioCppSTTSelect audioCppSTTFieldKind = iota
	audioCppSTTSlider
	audioCppSTTCheck
	audioCppSTTLongText
)

type audioCppSTTField struct {
	key      string
	label    string
	fallback interface{}
	kind     audioCppSTTFieldKind
	options  []CustomWidget.TextValueOption
	min      float64
	max      float64
	step     float64
	decimals int
}

type audioCppSTTGroup struct {
	title   string
	columns int
	fields  []audioCppSTTField
}

type audioCppSTTSchema struct {
	family      string
	title       string
	description string
	groups      []audioCppSTTGroup
}

func audioCppSTTChoice(key, label, fallback string, options ...CustomWidget.TextValueOption) audioCppSTTField {
	return audioCppSTTField{key: key, label: label, fallback: fallback, kind: audioCppSTTSelect, options: options}
}

func audioCppSTTIntSlider(key, label string, fallback int, min, max, step float64) audioCppSTTField {
	return audioCppSTTField{key: key, label: label, fallback: fallback, kind: audioCppSTTSlider, min: min, max: max, step: step}
}

func audioCppSTTFloatSlider(key, label string, fallback, min, max, step float64, decimals int) audioCppSTTField {
	return audioCppSTTField{key: key, label: label, fallback: fallback, kind: audioCppSTTSlider, min: min, max: max, step: step, decimals: decimals}
}

func audioCppSTTCheckbox(key, label string, fallback bool) audioCppSTTField {
	return audioCppSTTField{key: key, label: label, fallback: fallback, kind: audioCppSTTCheck}
}

func audioCppSTTParagraph(key, label, fallback string) audioCppSTTField {
	return audioCppSTTField{key: key, label: label, fallback: fallback, kind: audioCppSTTLongText}
}

func audioCppSTTMode(defaultValue string) audioCppSTTField {
	return audioCppSTTChoice("mode", "Processing mode", defaultValue,
		CustomWidget.TextValueOption{Text: "Automatic (offline final / streaming realtime)", Value: "auto"},
		CustomWidget.TextValueOption{Text: "Offline", Value: "offline"},
		CustomWidget.TextValueOption{Text: "Streaming", Value: "streaming"},
	)
}

func audioCppSTTChunkMode(defaultValue string) audioCppSTTField {
	return audioCppSTTChoice("audio_chunk_mode", "Long-audio chunking", defaultValue,
		CustomWidget.TextValueOption{Text: "Automatic", Value: "auto"},
		CustomWidget.TextValueOption{Text: "Voice activity detection", Value: "vad"},
		CustomWidget.TextValueOption{Text: "Fixed-size chunks", Value: "fixed"},
		CustomWidget.TextValueOption{Text: "No chunking", Value: "none"},
	)
}

func audioCppSTTSchemaForModel(model string) (audioCppSTTSchema, bool) {
	if schema, ok := additionalAudioCppSTTSchema(model); ok {
		return schema, true
	}
	switch model {
	case "Qwen3-ASR-0.6B-GGUF", "Qwen3-ASR-1.7B-GGUF":
		size := "0.6B"
		if model == "Qwen3-ASR-1.7B-GGUF" {
			size = "1.7B"
		}
		return audioCppSTTSchema{
			family: "qwen3_asr", title: "Qwen3-ASR " + size,
			description: "General multilingual recognition with context prompting. It can use VAD or fixed-size chunks for long recordings.",
			groups: []audioCppSTTGroup{{title: "Processing", columns: 2, fields: []audioCppSTTField{
				audioCppSTTMode("auto"),
				audioCppSTTChunkMode("auto"),
				audioCppSTTFloatSlider("audio_chunk_seconds", "Chunk length (seconds)", 30, 5, 120, 5, 0),
			}}},
		}, true
	case "Nemotron-3.5-ASR-Streaming-0.6B-GGUF":
		return audioCppSTTSchema{
			family: "nemotron_asr", title: "Nemotron 3.5 ASR Streaming 0.6B",
			description: "Low-latency RNN-T for 40 locales with native token timestamps. Lookahead trades a little latency for more future context.",
			groups: []audioCppSTTGroup{{title: "Streaming", columns: 2, fields: []audioCppSTTField{
				audioCppSTTMode("auto"),
				audioCppSTTIntSlider("lookahead_tokens", "Lookahead tokens", 0, 0, 32, 1),
				audioCppSTTCheckbox("keep_language_tags", "Keep language tags", false),
			}}},
		}, true
	case "VibeVoice-ASR-GGUF":
		return audioCppSTTSchema{
			family: "vibevoice_asr", title: "VibeVoice-ASR",
			description: "Large offline meeting transcription model with timestamps and speaker turns. The normal Beam and Repetition controls remain in effect.",
			groups: []audioCppSTTGroup{{title: "Long recordings", columns: 2, fields: []audioCppSTTField{
				audioCppSTTChunkMode("auto"),
				audioCppSTTFloatSlider("audio_chunk_seconds", "Chunk length (seconds)", 1200, 60, 3600, 60, 0),
			}}},
		}, true
	case "Voxtral-Mini-4B-Realtime-2602-GGUF":
		return audioCppSTTSchema{
			family: "voxtral_realtime", title: "Voxtral Mini 4B Realtime 2602",
			description: "Low-latency recognition for 13 languages. This checkpoint returns partial text but no timestamps.",
			groups: []audioCppSTTGroup{{title: "Processing", columns: 1, fields: []audioCppSTTField{
				audioCppSTTMode("auto"),
			}}},
		}, true
	case "Audio8-ASR-0.1B-GGUF":
		return audioCppSTTSchema{
			family: "audio8_asr", title: "Audio8-ASR 0.1B",
			description: "Compact offline recognition for English, Chinese, Cantonese, Japanese, Korean, French, and German. The CC-BY-NC-4.0 checkpoint must be converted or privately hosted by the user; it has no model-specific controls.",
		}, true
	case "Kroko-ASR-English-64L-GGUF":
		return audioCppSTTSchema{
			family: "kroko_asr", title: "Kroko ASR English 64L",
			description: "Very small English-only streaming RNN-T with word timestamps, phrase boosting, and automatic endpoint segments.",
			groups: []audioCppSTTGroup{
				{title: "Decoding", columns: 2, fields: []audioCppSTTField{
					audioCppSTTMode("auto"),
					audioCppSTTChoice("decoding_method", "Search method", "greedy_search",
						CustomWidget.TextValueOption{Text: "Greedy (fastest)", Value: "greedy_search"},
						CustomWidget.TextValueOption{Text: "Modified beam search", Value: "modified_beam_search"},
					),
					audioCppSTTIntSlider("num_beams", "Beam hypotheses", 4, 1, 64, 1),
				}},
				{title: "Phrase boosting", columns: 2, fields: []audioCppSTTField{
					audioCppSTTParagraph("hotwords", "Hotwords (one phrase per line)", ""),
					audioCppSTTFloatSlider("hotwords_score", "Hotword strength", 1.5, 0, 10, 0.1, 1),
				}},
				{title: "Automatic endpoints", columns: 2, fields: []audioCppSTTField{
					audioCppSTTCheckbox("enable_endpoint", "Enable endpoint segments", false),
					audioCppSTTFloatSlider("rule1_min_trailing_silence_sec", "Silence without decoded speech", 2.4, 0, 15, 0.1, 1),
					audioCppSTTFloatSlider("rule2_min_trailing_silence_sec", "Silence after decoded speech", 1.2, 0, 15, 0.1, 1),
					audioCppSTTFloatSlider("rule3_min_utterance_length_sec", "Maximum utterance length", 20, 1, 120, 1, 0),
				}},
			},
		}, true
	default:
		return audioCppSTTSchema{}, false
	}
}

func audioCppSTTNumber(value interface{}) float64 {
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case float64:
		return typed
	default:
		return 0
	}
}

func audioCppSTTControl(schema audioCppSTTSchema, field audioCppSTTField) fyne.CanvasObject {
	current := special.GetNestedSpecialSettingFallback(audioCppSTTSettingsGroup, schema.family, field.key, field.fallback)
	switch field.kind {
	case audioCppSTTSelect:
		selector := CustomWidget.NewTextValueSelect("audio_cpp_stt_"+schema.family+"_"+field.key, field.options, nil, 0)
		selector.SetSelected(fmt.Sprint(current))
		selector.OnChanged = func(option CustomWidget.TextValueOption) {
			special.UpdateNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family, field.key, option.Value, false)
		}
		return selector
	case audioCppSTTSlider:
		value := math.Max(field.min, math.Min(field.max, audioCppSTTNumber(current)))
		slider := widget.NewSlider(field.min, field.max)
		slider.Step = field.step
		slider.SetValue(value)
		valueLabel := widget.NewLabel(fmt.Sprintf("%.*f", field.decimals, value))
		valueLabel.Alignment = fyne.TextAlignTrailing
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.*f", field.decimals, value))
			if _, integer := field.fallback.(int); integer {
				special.UpdateNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family, field.key, int(math.Round(value)), false)
			} else {
				special.UpdateNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family, field.key, value, false)
			}
		}
		return container.NewBorder(nil, nil, nil, valueLabel, slider)
	case audioCppSTTCheck:
		check := widget.NewCheck("", nil)
		check.SetChecked(current.(bool))
		check.OnChanged = func(value bool) {
			special.UpdateNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family, field.key, value, false)
		}
		return check
	case audioCppSTTLongText:
		entry := widget.NewMultiLineEntry()
		entry.SetMinRowsVisible(2)
		entry.SetPlaceHolder(field.label)
		entry.SetText(current.(string))
		entry.OnChanged = func(value string) {
			special.UpdateNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family, field.key, value, true)
		}
		return entry
	default:
		return widget.NewLabel("")
	}
}

func audioCppSTTForm(schema audioCppSTTSchema, fields []audioCppSTTField) fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(fields)*2)
	for _, field := range fields {
		objects = append(objects, widget.NewLabel(field.label+":"), audioCppSTTControl(schema, field))
	}
	return container.New(layout.NewFormLayout(), objects...)
}

func audioCppSTTGroupContent(schema audioCppSTTSchema, group audioCppSTTGroup) fyne.CanvasObject {
	if group.columns < 2 || len(group.fields) < 2 {
		return audioCppSTTForm(schema, group.fields)
	}
	middle := (len(group.fields) + 1) / 2
	return container.NewGridWithColumns(2,
		audioCppSTTForm(schema, group.fields[:middle]),
		audioCppSTTForm(schema, group.fields[middle:]),
	)
}

func buildAudioCppSTTSettingsForModel(model string, rebuild func(string)) fyne.CanvasObject {
	schema, ok := audioCppSTTSchemaForModel(model)
	if !ok {
		return widget.NewLabel("No audio.cpp settings schema is available for " + model + ".")
	}
	if len(schema.groups) == 0 {
		return container.NewVBox()
	}
	items := make([]*widget.AccordionItem, 0, len(schema.groups))
	for _, group := range schema.groups {
		items = append(items, widget.NewAccordionItem(group.title, audioCppSTTGroupContent(schema, group)))
	}
	accordion := widget.NewAccordion(items...)
	accordion.Open(0)
	reset := widget.NewButton("Reset", func() {
		special.ResetNestedSpecialSettings(audioCppSTTSettingsGroup, schema.family)
		if rebuild != nil {
			rebuild(model)
		}
	})
	reset.Importance = widget.LowImportance
	return container.NewVBox(accordion, container.NewHBox(layout.NewSpacer(), reset))
}

func buildAudioCppSTTSpecialSettings() fyne.CanvasObject {
	host := container.NewStack()
	var rebuild func(string)
	rebuild = func(model string) {
		host.Objects = []fyne.CanvasObject{buildAudioCppSTTSettingsForModel(model, rebuild)}
		host.Refresh()
	}
	rebuild(Settings.Config.Model)
	return host
}
