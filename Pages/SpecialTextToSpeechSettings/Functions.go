package SpecialTextToSpeechSettings

import (
	"reflect"
	"strconv"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func UpdateSpecialTTSSettings(specialSettingType string, stringName string, value interface{}) {
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
		return
	}

	// Update and notify
	inner[stringName] = value

	sendMessage := SendMessageChannel.SendMessageStruct{
		Type:  "special_settings",
		Name:  specialSettingType,
		Value: inner,
	}
	sendMessage.SendMessage()
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
