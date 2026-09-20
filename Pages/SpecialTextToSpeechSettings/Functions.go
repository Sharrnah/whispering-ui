package SpecialTextToSpeechSettings

import (
	"reflect"
	"strconv"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func stringInterfaceMap(value interface{}) map[string]interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			name, ok := key.(string)
			if ok {
				converted[name] = item
			}
		}
		return converted
	default:
		return make(map[string]interface{})
	}
}

// SetSpecialTTSSetting updates the in-memory profile without sending a backend
// message. It is useful while constructing controls, before the WebSocket
// sender is guaranteed to be running.
func SetSpecialTTSSetting(specialSettingType string, stringName string, value interface{}) bool {
	// Ensure outer map exists
	if Settings.Config.Special_settings == nil {
		Settings.Config.Special_settings = make(map[string]interface{})
	}

	// Ensure inner map exists and has correct type
	inner, ok := Settings.Config.Special_settings[specialSettingType].(map[string]interface{})
	if !ok || inner == nil {
		inner = make(map[string]interface{})
		Settings.Config.Special_settings[specialSettingType] = inner
	}

	// If the value is the same, do nothing (safe compare for non-comparable types)
	if old, exists := inner[stringName]; exists && reflect.DeepEqual(old, value) {
		return false
	}

	// Update locally. The caller decides whether the backend must be notified.
	inner[stringName] = value
	return true
}

func UpdateSpecialTTSSettings(specialSettingType string, stringName string, value interface{}) {
	if !SetSpecialTTSSetting(specialSettingType, stringName, value) {
		return
	}
	inner := Settings.Config.Special_settings[specialSettingType].(map[string]interface{})

	sendMessage := SendMessageChannel.SendMessageStruct{
		Type:  "special_settings",
		Name:  specialSettingType,
		Value: inner,
	}
	sendMessage.SendMessage()
}

func UpdateTTSPrecision(specialSettingType string, precision string) {
	changed := Settings.Config.Tts_precision != precision
	Settings.Config.Tts_precision = precision
	UpdateSpecialTTSSettings(specialSettingType, "precision", precision)
	if changed {
		SendMessageChannel.SendMessageStruct{
			Type:  "setting_change",
			Name:  "tts_precision",
			Value: precision,
		}.SendMessage()
	}
}

func GetSpecialTTSSettings(specialSettingType string, stringName string) interface{} {
	// Ensure outer map exists
	if Settings.Config.Special_settings == nil {
		Settings.Config.Special_settings = make(map[string]interface{})
	}

	// Ensure inner map exists and has correct type
	inner, ok := Settings.Config.Special_settings[specialSettingType].(map[string]interface{})
	if !ok || inner == nil {
		inner = make(map[string]interface{})
		Settings.Config.Special_settings[specialSettingType] = inner
	}

	// Return value if present
	if val, ok := inner[stringName]; ok {
		return val
	}
	return nil
}

func GetSpecialSettingFallback(specialSettingType string, key string, fallback interface{}) interface{} {
	if Settings.Config.Special_settings == nil {
		return fallback
	}
	val, ok := Settings.Config.Special_settings[specialSettingType]
	if !ok || val == nil {
		return fallback
	}
	// Support both map[string]interface{} and map[interface{}]interface{} (e.g. YAML)
	var settingMap map[string]interface{}
	switch m := val.(type) {
	case map[string]interface{}:
		settingMap = m
	case map[interface{}]interface{}:
		settingMap = make(map[string]interface{}, len(m))
		for k, v := range m {
			ks, ok := k.(string)
			if !ok {
				continue
			}
			settingMap[ks] = v
		}
	default:
		return fallback
	}

	raw, ok := settingMap[key]
	if !ok || raw == nil {
		return fallback
	}
	return coerceSpecialSetting(raw, fallback)
}

func coerceSpecialSetting(raw, fallback interface{}) interface{} {
	// Coerce based on the fallback type to ensure safe return types.
	switch fb := fallback.(type) {
	case string:
		if s, ok := raw.(string); ok {
			return s
		}
		return fb
	case float64:
		switch r := raw.(type) {
		case float64:
			return r
		case float32:
			return float64(r)
		case int:
			return float64(r)
		case int64:
			return float64(r)
		case int32:
			return float64(r)
		case string:
			if f, err := strconv.ParseFloat(r, 64); err == nil {
				return f
			}
			return fb
		default:
			return fb
		}
	case int:
		switch r := raw.(type) {
		case int:
			return r
		case int64:
			return int(r)
		case int32:
			return int(r)
		case float64:
			return int(r)
		case float32:
			return int(r)
		case string:
			if i, err := strconv.Atoi(r); err == nil {
				return i
			}
			return fb
		default:
			return fb
		}
	case bool:
		if b, ok := raw.(bool); ok {
			return b
		}
		if s, ok := raw.(string); ok {
			if s == "true" {
				return true
			} else if s == "false" {
				return false
			}
		}
		return fb
	default:
		// Unknown target type, return fallback to be safe.
		return fallback
	}
}

// GetNestedSpecialSettingFallback reads the model-specific shape used by
// audio.cpp: special_settings.<group>.<model-family>.<setting>.
func GetNestedSpecialSettingFallback(group, modelFamily, key string, fallback interface{}) interface{} {
	if Settings.Config.Special_settings == nil {
		return fallback
	}
	root := stringInterfaceMap(Settings.Config.Special_settings[group])
	model := stringInterfaceMap(root[modelFamily])
	if value, ok := model[key]; ok {
		return coerceSpecialSetting(value, fallback)
	}
	return fallback
}

// UpdateNestedSpecialSettings preserves settings for every other audio.cpp
// model family and sends the complete group to the backend. Entry controls use
// the debounced path so editing a number or prompt does not emit one WebSocket
// request per keystroke.
func UpdateNestedSpecialSettings(group, modelFamily, key string, value interface{}, debounce bool) {
	if Settings.Config.Special_settings == nil {
		Settings.Config.Special_settings = make(map[string]interface{})
	}
	root := stringInterfaceMap(Settings.Config.Special_settings[group])
	model := stringInterfaceMap(root[modelFamily])
	if old, exists := model[key]; exists && reflect.DeepEqual(old, value) {
		return
	}
	model[key] = value
	root[modelFamily] = model
	Settings.Config.Special_settings[group] = root
	message := SendMessageChannel.SendMessageStruct{
		Type:  "special_settings",
		Name:  group,
		Value: root,
	}
	if debounce {
		message.SendMessageDebounced()
	} else {
		message.SendMessage()
	}
}

func ResetNestedSpecialSettings(group, modelFamily string) {
	if Settings.Config.Special_settings == nil {
		Settings.Config.Special_settings = make(map[string]interface{})
	}
	root := stringInterfaceMap(Settings.Config.Special_settings[group])
	delete(root, modelFamily)
	Settings.Config.Special_settings[group] = root
	SendMessageChannel.SendMessageStruct{
		Type:  "special_settings",
		Name:  group,
		Value: root,
	}.SendMessage()
}
