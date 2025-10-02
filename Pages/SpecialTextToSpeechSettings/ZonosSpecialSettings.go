package SpecialTextToSpeechSettings

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/CustomWidget"
)

func BuildZonosSpecialSettings() fyne.CanvasObject {
	languageSelect := CustomWidget.NewTextValueSelect("language",
		[]CustomWidget.TextValueOption{
			{Text: "Afrikaans", Value: "af"},
			{Text: "Amharic", Value: "am"},
			{Text: "Aragonese", Value: "an"},
			//{Text: "Arabic", Value: "ar"},
			{Text: "Assamese", Value: "as"},
			{Text: "Azerbaijani", Value: "az"},
			{Text: "Bashkir", Value: "ba"},
			{Text: "Bulgarian", Value: "bg"},
			{Text: "Bengali", Value: "bn"},
			{Text: "Bishnupriya Manipuri", Value: "bpy"},
			{Text: "Bosnian", Value: "bs"},
			{Text: "Catalan", Value: "ca"},
			{Text: "Chinese (Mandarin)", Value: "cmn"},
			//{Text: "Czech", Value: "cs"},
			{Text: "Welsh", Value: "cy"},
			{Text: "Danish", Value: "da"},
			{Text: "German", Value: "de"},
			//{Text: "Greek", Value: "el"},
			{Text: "English (Caribbean)", Value: "en-029"},
			{Text: "English (UK)", Value: "en-gb"},
			{Text: "English (Scotland)", Value: "en-gb-scotland"},
			{Text: "English (GB Clan)", Value: "en-gb-x-gbclan"},
			{Text: "English (GB CWMD)", Value: "en-gb-x-gbcwmd"},
			{Text: "English (RP)", Value: "en-gb-x-rp"},
			{Text: "English (US)", Value: "en-us"},
			{Text: "Esperanto", Value: "eo"},
			{Text: "Spanish", Value: "es"},
			{Text: "Spanish (Latin America)", Value: "es-419"},
			{Text: "Estonian", Value: "et"},
			{Text: "Basque", Value: "eu"},
			{Text: "Persian", Value: "fa"},
			{Text: "Persian (Latin)", Value: "fa-latn"},
			{Text: "Finnish", Value: "fi"},
			{Text: "French (Belgium)", Value: "fr-be"},
			{Text: "French (Switzerland)", Value: "fr-ch"},
			{Text: "French (France)", Value: "fr-fr"},
			{Text: "Irish", Value: "ga"},
			{Text: "Scottish Gaelic", Value: "gd"},
			{Text: "Guarani", Value: "gn"},
			{Text: "Ancient Greek", Value: "grc"},
			{Text: "Gujarati", Value: "gu"},
			{Text: "Hakka", Value: "hak"},
			//{Text: "Hindi", Value: "hi"},
			{Text: "Croatian", Value: "hr"},
			{Text: "Haitian Creole", Value: "ht"},
			{Text: "Hungarian", Value: "hu"},
			{Text: "Armenian", Value: "hy"},
			{Text: "Western Armenian", Value: "hyw"},
			{Text: "Interlingua", Value: "ia"},
			{Text: "Indonesian", Value: "id"},
			{Text: "Icelandic", Value: "is"},
			{Text: "Italian", Value: "it"},
			{Text: "Japanese", Value: "ja"},
			{Text: "Lojban", Value: "jbo"},
			{Text: "Georgian", Value: "ka"},
			{Text: "Kazakh", Value: "kk"},
			{Text: "Kalaallisut", Value: "kl"},
			{Text: "Kannada", Value: "kn"},
			{Text: "Korean", Value: "ko"},
			{Text: "Konkani", Value: "kok"},
			{Text: "Kurdish", Value: "ku"},
			{Text: "Kyrgyz", Value: "ky"},
			{Text: "Latin", Value: "la"},
			{Text: "Lingua Franca Nova", Value: "lfn"},
			{Text: "Lithuanian", Value: "lt"},
			{Text: "Latvian", Value: "lv"},
			{Text: "Māori", Value: "mi"},
			{Text: "Macedonian", Value: "mk"},
			{Text: "Malayalam", Value: "ml"},
			{Text: "Marathi", Value: "mr"},
			//{Text: "Malay", Value: "ms"},
			{Text: "Maltese", Value: "mt"},
			{Text: "Burmese", Value: "my"},
			{Text: "Norwegian Bokmål", Value: "nb"},
			{Text: "Classical Nahuatl", Value: "nci"},
			{Text: "Nepali", Value: "ne"},
			{Text: "Dutch", Value: "nl"},
			{Text: "Oromo", Value: "om"},
			{Text: "Oriya", Value: "or"},
			{Text: "Punjabi", Value: "pa"},
			{Text: "Papiamento", Value: "pap"},
			{Text: "Polish", Value: "pl"},
			{Text: "Portuguese", Value: "pt"},
			{Text: "Portuguese (Brazil)", Value: "pt-br"},
			{Text: "Paraguayan Guarani", Value: "py"},
			{Text: "K'iche'", Value: "quc"},
			{Text: "Romanian", Value: "ro"},
			{Text: "Russian", Value: "ru"},
			{Text: "Russian (Latvia)", Value: "ru-lv"},
			{Text: "Sindhi", Value: "sd"},
			{Text: "Shan", Value: "shn"},
			{Text: "Sinhala", Value: "si"},
			{Text: "Slovak", Value: "sk"},
			{Text: "Slovenian", Value: "sl"},
			{Text: "Albanian", Value: "sq"},
			{Text: "Serbian", Value: "sr"},
			{Text: "Swedish", Value: "sv"},
			{Text: "Swahili", Value: "sw"},
			{Text: "Tamil", Value: "ta"},
			{Text: "Telugu", Value: "te"},
			{Text: "Tswana", Value: "tn"},
			{Text: "Turkish", Value: "tr"},
			{Text: "Tatar", Value: "tt"},
			{Text: "Urdu", Value: "ur"},
			{Text: "Uzbek", Value: "uz"},
			{Text: "Vietnamese", Value: "vi"},
			{Text: "Vietnamese (Central)", Value: "vi-vn-x-central"},
			{Text: "Vietnamese (South)", Value: "vi-vn-x-south"},
			{Text: "Cantonese", Value: "yue"},
		},
		func(option CustomWidget.TextValueOption) {

		},
		0,
	)

	// Load language with safe fallback
	language := GetSpecialSettingFallback("tts_zonos", "language", "en-us").(string)
	languageSelect.SetSelected(language)

	seedInput := widget.NewEntry()
	seedInput.PlaceHolder = lang.L("Enter manual seed")
	// Load seed (optional)
	if seed := GetSpecialSettingFallback("tts_zonos", "seed", "").(string); seed != "" {
		seedInput.SetText(seed)
	}

	// Clamp helper to keep values in [0,1]
	clamp01 := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}

	happinessSlider := widget.NewSlider(0.0, 1.0)
	happinessSlider.Step = 0.0001
	happinessSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "happiness", 0.3077).(float64)))

	sadnessSlider := widget.NewSlider(0.0, 1.0)
	sadnessSlider.Step = 0.0001
	sadnessSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "sadness", 0.0256).(float64)))

	disgustSlider := widget.NewSlider(0.0, 1.0)
	disgustSlider.Step = 0.0001
	disgustSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "disgust", 0.0256).(float64)))

	fearSlider := widget.NewSlider(0.0, 1.0)
	fearSlider.Step = 0.0001
	fearSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "fear", 0.0256).(float64)))

	surpriseSlider := widget.NewSlider(0.0, 1.0)
	surpriseSlider.Step = 0.0001
	surpriseSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "surprise", 0.0256).(float64)))

	angerSlider := widget.NewSlider(0.0, 1.0)
	angerSlider.Step = 0.0001
	angerSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "anger", 0.0256).(float64)))

	otherSlider := widget.NewSlider(0.0, 1.0)
	otherSlider.Step = 0.0001
	otherSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "other", 0.2564).(float64)))

	neutralSlider := widget.NewSlider(0.0, 1.0)
	neutralSlider.Step = 0.0001
	neutralSlider.SetValue(clamp01(GetSpecialSettingFallback("tts_zonos", "neutral", 0.3077).(float64)))

	// ignore list checkboxes
	emotionCheck := widget.NewCheck("", func(b bool) {})
	pitchStdCheck := widget.NewCheck("", func(b bool) {})
	//vqscore8Check := widget.NewCheck("", func(b bool) {})
	//ctcLossCheck := widget.NewCheck("", func(b bool) {})

	speakerCheck := widget.NewCheck("", func(b bool) {})
	//fmaxCheck := widget.NewCheck("", func(b bool) {})
	speaking_rateCheck := widget.NewCheck("", func(b bool) {})
	//dnsmosOvrlCheck := widget.NewCheck("", func(b bool) {})

	updateSpecialTTSSettings := func() {
		var ignoreList []string
		if emotionCheck.Checked {
			ignoreList = append(ignoreList, "emotion")
		}
		if pitchStdCheck.Checked {
			ignoreList = append(ignoreList, "pitch_std")
		}
		//if vqscore8Check.Checked {
		//	ignoreList = append(ignoreList, "vqscore8")
		//}
		//if ctcLossCheck.Checked {
		//	ignoreList = append(ignoreList, "ctc_loss")
		//}
		if speakerCheck.Checked {
			ignoreList = append(ignoreList, "speaker")
		}
		//if fmaxCheck.Checked {
		//	ignoreList = append(ignoreList, "fmax")
		//}
		if speaking_rateCheck.Checked {
			ignoreList = append(ignoreList, "speaking_rate")
		}
		//if dnsmosOvrlCheck.Checked {
		//	ignoreList = append(ignoreList, "dnsmos_ovrl")
		//}
		if ignoreList == nil {
			ignoreList = []string{}
		}

		UpdateSpecialTTSSettings("tts_zonos", "language", languageSelect.GetSelected().Value)
		UpdateSpecialTTSSettings("tts_zonos", "seed", seedInput.Text)
		UpdateSpecialTTSSettings("tts_zonos", "happiness", happinessSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "sadness", sadnessSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "disgust", disgustSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "fear", fearSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "surprise", surpriseSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "anger", angerSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "other", otherSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "neutral", neutralSlider.Value)
		UpdateSpecialTTSSettings("tts_zonos", "ignore_list", ignoreList)

		//sendMessage := SendMessageChannel.SendMessageStruct{
		//	Type: "special_settings",
		//	Name: "tts_zonos",
		//	Value: struct {
		//		Language   string   `json:"language"`
		//		Seed       string   `json:"seed"`
		//		Happiness  float64  `json:"happiness"`
		//		Sadness    float64  `json:"sadness"`
		//		Disgust    float64  `json:"disgust"`
		//		Fear       float64  `json:"fear"`
		//		Surprise   float64  `json:"surprise"`
		//		Anger      float64  `json:"anger"`
		//		Other      float64  `json:"other"`
		//		Neutral    float64  `json:"neutral"`
		//		IgnoreList []string `json:"ignore_list"`
		//	}{
		//		Language:   languageSelect.GetSelected().Value,
		//		Seed:       seedInput.Text,
		//		Happiness:  happinessSlider.Value,
		//		Sadness:    sadnessSlider.Value,
		//		Disgust:    disgustSlider.Value,
		//		Fear:       fearSlider.Value,
		//		Surprise:   surpriseSlider.Value,
		//		Anger:      angerSlider.Value,
		//		Other:      otherSlider.Value,
		//		Neutral:    neutralSlider.Value,
		//		IgnoreList: ignoreList,
		//	},
		//}
		//sendMessage.SendMessage()
	}
	languageSelect.OnChanged = func(option CustomWidget.TextValueOption) {
		updateSpecialTTSSettings()
	}
	seedInput.OnChanged = func(s string) {
		updateSpecialTTSSettings()
	}

	happinessSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	sadnessSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	disgustSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	fearSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	surpriseSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	angerSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	otherSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}
	neutralSlider.OnChanged = func(f float64) {
		updateSpecialTTSSettings()
	}

	emotionCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	pitchStdCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	//vqscore8Check.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	//ctcLossCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }

	speakerCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	//fmaxCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	speaking_rateCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }
	//dnsmosOvrlCheck.OnChanged = func(b bool) { updateSpecialTTSSettings() }

	advancedSettings := container.New(layout.NewVBoxLayout(),
		widget.NewLabel(" "),
		container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.L("Language")+":"),
			languageSelect,
			widget.NewLabel(lang.L("Seed")+":"),
			seedInput,
		),
		widget.NewAccordion(
			widget.NewAccordionItem(lang.L("Emotion"),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Happiness")+":"),
						happinessSlider,
						widget.NewLabel(lang.L("Sadness")+":"),
						sadnessSlider,
						widget.NewLabel(lang.L("Disgust")+":"),
						disgustSlider,
						widget.NewLabel(lang.L("Fear")+":"),
						fearSlider,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("Surprise")+":"),
						surpriseSlider,
						widget.NewLabel(lang.L("Anger")+":"),
						angerSlider,
						widget.NewLabel(lang.L("Other")+":"),
						otherSlider,
						widget.NewLabel(lang.L("Neutral")+":"),
						neutralSlider,
					),
				),
			),
			widget.NewAccordionItem(lang.L("Ignore"),
				container.NewGridWithColumns(2,
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("emotion")+":"),
						emotionCheck,
						widget.NewLabel(lang.L("speaker")+":"),
						speakerCheck,
					),
					container.New(layout.NewFormLayout(),
						widget.NewLabel(lang.L("pitch_std")+":"),
						pitchStdCheck,
						widget.NewLabel(lang.L("speaking_rate")+":"),
						speaking_rateCheck,
						//widget.NewLabel(lang.L("fmax")+":"),
						//fmaxCheck,
						//widget.NewLabel(lang.L("dnsmos_ovrl")+":"),
						//dnsmosOvrlCheck,
						//widget.NewLabel(lang.L("ctc_loss")+":"),
						//ctcLossCheck,
						//widget.NewLabel(lang.L("vqscore_8")+":"),
						//vqscore8Check,
					),
				),
			),
		),
	)
	return advancedSettings
}
