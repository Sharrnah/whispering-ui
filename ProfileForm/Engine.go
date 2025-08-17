package ProfileForm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Settings"

	"fyne.io/fyne/v2/widget"
)

// AllProfileControls aggregates widget references so other logic can coordinate them
// without relying on fragile form indices. This is intentionally explicit and
// mirrors the current form layout sections.
type AllProfileControls struct {
	// Connection / backend
	WebsocketIP   *widget.Entry
	WebsocketPort *widget.Entry
	RunBackend    *widget.Check

	// Audio
	AudioAPI    *CustomWidget.TextValueSelect
	AudioInput  *CustomWidget.TextValueSelect
	AudioOutput *CustomWidget.TextValueSelect

	// VAD / Recording
	VadEnable       *widget.Check
	VadOnFullClip   *widget.Check
	VadRealtime     *widget.Check
	PushToTalk      *CustomWidget.HotKeyEntry
	VadConfidence   *widget.Slider
	Energy          *widget.Slider
	DenoiseAudio    *CustomWidget.TextValueSelect
	PauseSeconds    *widget.Slider
	PhraseTimeLimit *widget.Slider

	// STT
	STTType      *CustomWidget.TextValueSelect
	STTDevice    *CustomWidget.TextValueSelect
	STTModelSize *CustomWidget.TextValueSelect
	STTPrecision *CustomWidget.TextValueSelect

	// Text translation
	TxtType      *CustomWidget.TextValueSelect
	TxtDevice    *CustomWidget.TextValueSelect
	TxtSize      *CustomWidget.TextValueSelect
	TxtPrecision *CustomWidget.TextValueSelect

	// TTS
	TTSType   *CustomWidget.TextValueSelect
	TTSDevice *CustomWidget.TextValueSelect

	// OCR
	OCRType      *CustomWidget.TextValueSelect
	OCRDevice    *CustomWidget.TextValueSelect
	OCRPrecision *CustomWidget.TextValueSelect
}

// FormEngine centralizes cross-field operations and holds the controls + coordinator
type FormEngine struct {
	Controls *AllProfileControls
	Coord    *Coordinator
	// Bindings map profile setting field names (lowercase) to concrete controls
	// so generic load/save can work without manual wiring per field in loader code.
	Bindings map[string]any
}

func NewFormEngine(c *AllProfileControls, coord *Coordinator) *FormEngine {
	return &FormEngine{Controls: c, Coord: coord, Bindings: map[string]any{}}
}

// Helper to set options for a select with graceful fallback to a valid selection
func (e *FormEngine) SetOptionsWithFallback(sel *CustomWidget.TextValueSelect, options []CustomWidget.TextValueOption) {
	if e == nil || sel == nil {
		return
	}
	if e.Coord != nil {
		e.Coord.SetOptionsWithFallback(sel, options)
		return
	}
	// fallback without coordinator
	current := sel.GetSelected()
	sel.Options = options
	if current == nil || !sel.ContainsEntry(current, CustomWidget.CompareValue) {
		if len(options) > 0 {
			sel.SetSelectedIndex(0)
		}
	} else {
		sel.SetSelected(current.Value)
	}
	sel.Refresh()
}

// Register binds a settings field name to a control for generic load/save
func (e *FormEngine) Register(settingKey string, control any) {
	if e == nil || settingKey == "" || control == nil {
		return
	}
	if e.Bindings == nil {
		e.Bindings = map[string]any{}
	}
	e.Bindings[settingKey] = control
}

// helper to set TextValueSelect by value, or by label text fallback
func (e *FormEngine) selectSetByValueOrText(sel *CustomWidget.TextValueSelect, value string, fallbackText string) {
	if sel == nil {
		return
	}
	if value != "" {
		sel.SetSelected(value)
		// if not contained (SetSelected may ignore), try fallback logic
		if cur := sel.GetSelected(); cur == nil || cur.Value != value {
			// try by text equality
			if fallbackText != "" {
				for _, opt := range sel.Options {
					if opt.Text == fallbackText {
						sel.SetSelected(opt.Value)
						break
					}
				}
			}
		}
	} else if fallbackText != "" {
		for _, opt := range sel.Options {
			if opt.Text == fallbackText {
				sel.SetSelected(opt.Value)
				break
			}
		}
	} else {
		// empty value: prefer selecting a "Disabled" option if present (Value == "")
		for _, opt := range sel.Options {
			if opt.Value == "" {
				sel.SetSelected("")
				break
			}
		}
	}
}

// LoadFromSettings iterates all registered bindings and applies the settings values into controls
func (e *FormEngine) LoadFromSettings(conf *Settings.Conf) {
	if e == nil || conf == nil {
		return
	}
	for key, ctrl := range e.Bindings {
		switch c := ctrl.(type) {
		case *widget.Entry:
			// read as string via fmt.Sprint to handle ints as well
			val := e.getOptionByLowercase(conf, key)
			if val != nil {
				c.SetText(fmt.Sprint(val))
			}
		case *widget.Check:
			val := e.getOptionByLowercase(conf, key)
			if b, ok := val.(bool); ok {
				c.SetChecked(b)
			}
		case *widget.Slider:
			val := e.getOptionByLowercase(conf, key)
			switch v := val.(type) {
			case float64:
				c.SetValue(v)
			case int:
				c.SetValue(float64(v))
			case nil:
				// ignore
			default:
				// try parse if string
				if s, ok := v.(string); ok {
					if f, err := strconv.ParseFloat(s, 64); err == nil {
						c.SetValue(f)
					}
				}
			}
		case *CustomWidget.TextValueSelect:
			// default: set by value. For audio devices we also have name fields as fallback
			val := e.getOptionByLowercase(conf, key)
			strVal := fmt.Sprint(val)
			fallbackText := ""
			if key == "device_index" {
				if name := e.getOptionByLowercase(conf, "audio_input_device"); name != nil {
					fallbackText = fmt.Sprint(name)
				}
			}
			if key == "device_out_index" {
				if name := e.getOptionByLowercase(conf, "audio_output_device"); name != nil {
					fallbackText = fmt.Sprint(name)
				}
			}
			e.selectSetByValueOrText(c, strVal, fallbackText)
		case *CustomWidget.HotKeyEntry:
			val := e.getOptionByLowercase(conf, key)
			if s, ok := val.(string); ok {
				c.SetText(s)
			}
		}
	}
}

// getOptionByLowercase retrieves a field value from Settings.Conf by matching the lowercase key
func (e *FormEngine) getOptionByLowercase(conf *Settings.Conf, key string) any {
	if conf == nil || key == "" {
		return nil
	}
	rv := reflect.ValueOf(conf).Elem()
	rt := rv.Type()
	lk := strings.ToLower(key)
	for i := 0; i < rt.NumField(); i++ {
		if strings.ToLower(rt.Field(i).Name) == lk {
			return rv.Field(i).Interface()
		}
	}
	// final fallback using existing GetOption
	if v, err := conf.GetOption(key); err == nil {
		return v
	}
	return nil
}

// SaveToSettings iterates all registered bindings and writes control values back to settings
func (e *FormEngine) SaveToSettings(conf *Settings.Conf) {
	if e == nil || conf == nil {
		return
	}
	for key, ctrl := range e.Bindings {
		switch c := ctrl.(type) {
		case *widget.Entry:
			Settings.Config.SetOption(key, c.Text)
			conf.SetOption(key, c.Text)
		case *widget.Check:
			conf.SetOption(key, c.Checked)
		case *widget.Slider:
			// value is float64; Conf.SetOption converts where needed
			conf.SetOption(key, c.Value)
		case *CustomWidget.TextValueSelect:
			sel := c.GetSelected()
			if sel == nil {
				conf.SetOption(key, "")
				continue
			}
			// handle audio devices indexes specially (store int and name)
			if key == "device_index" || key == "device_out_index" {
				if iv, err := strconv.Atoi(sel.Value); err == nil {
					conf.SetOption(key, iv)
				} else {
					conf.SetOption(key, sel.Value)
				}
				if key == "device_index" {
					conf.SetOption("audio_input_device", sel.Text)
				} else {
					conf.SetOption("audio_output_device", sel.Text)
				}
			} else {
				conf.SetOption(key, sel.Value)
			}
		case *CustomWidget.HotKeyEntry:
			conf.SetOption(key, c.Text)
		}
	}
}
