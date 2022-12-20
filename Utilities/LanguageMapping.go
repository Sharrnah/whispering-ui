package Utilities

type LanguageMap struct {
	Name string
	ISO1 string
	ISO3 []string
}

type LanguageMapping struct {
	LanguageMappings []LanguageMap
}

var LanguageMapList = LanguageMapping{LanguageMappings: []LanguageMap{
	{Name: "Acehnese", ISO1: "ace", ISO3: []string{"ace_Arab", "ace_Latn"}},
	{Name: "Mesopotamian Arabic", ISO1: "acm", ISO3: []string{"acm_Arab"}},
	{Name: "Ta’izzi-Adeni Arabic", ISO1: "acq", ISO3: []string{"acq_Arab"}},
	{Name: "Tunisian Arabic", ISO1: "aeb", ISO3: []string{"aeb_Arab"}},
	{Name: "Afrikaans", ISO1: "af", ISO3: []string{"afr_Latn"}},
	{Name: "South Levantine Arabic", ISO1: "ajp", ISO3: []string{"ajp_Arab"}},
	{Name: "Akan", ISO1: "ak", ISO3: []string{"aka_Latn"}},
	{Name: "Amharic", ISO1: "am", ISO3: []string{"amh_Ethi"}},
	{Name: "North Levantine Arabic", ISO1: "apc", ISO3: []string{"apc_Arab"}},
	{Name: "Modern Standard Arabic", ISO1: "arb", ISO3: []string{"arb_Arab"}},
	{Name: "Najdi Arabic", ISO1: "ars", ISO3: []string{"ars_Arab"}},
	{Name: "Moroccan Arabic", ISO1: "ary", ISO3: []string{"ary_Arab"}},
	{Name: "Egyptian Arabic", ISO1: "arz", ISO3: []string{"arz_Arab"}},
	{Name: "Assamese", ISO1: "as", ISO3: []string{"asm_Beng"}},
	{Name: "Asturian", ISO1: "ast", ISO3: []string{"ast_Latn"}},
	{Name: "Awadhi", ISO1: "awa", ISO3: []string{"awa_Deva"}},
	{Name: "Central Aymara", ISO1: "ayr", ISO3: []string{"ayr_Latn"}},
	{Name: "South Azerbaijani", ISO1: "azb", ISO3: []string{"azb_Arab"}},
	{Name: "North Azerbaijani", ISO1: "azj", ISO3: []string{"azj_Latn"}},
	{Name: "Bashkir", ISO1: "ba", ISO3: []string{"bak_Cyrl"}},
	{Name: "Bambara", ISO1: "bm", ISO3: []string{"bam_Latn"}},
	{Name: "Balinese", ISO1: "ban", ISO3: []string{"ban_Latn"}},
	{Name: "Belarusian", ISO1: "be", ISO3: []string{"bel_Cyrl"}},
	{Name: "Bemba", ISO1: "bem", ISO3: []string{"bem_Latn"}},
	{Name: "Bengali", ISO1: "bn", ISO3: []string{"ben_Beng"}},
	{Name: "Bhojpuri", ISO1: "bho", ISO3: []string{"bho_Deva"}},
	{Name: "Banjar", ISO1: "bjn", ISO3: []string{"bjn_Arab", "bjn_Latn"}},
	{Name: "Standard Tibetan", ISO1: "bo", ISO3: []string{"bod_Tibt"}},
	{Name: "Bosnian", ISO1: "bs", ISO3: []string{"bos_Latn"}},
	{Name: "Buginese", ISO1: "bug", ISO3: []string{"bug_Latn"}},
	{Name: "Bulgarian", ISO1: "bg", ISO3: []string{"bul_Cyrl"}},
	{Name: "Catalan", ISO1: "ca", ISO3: []string{"cat_Latn"}},
	{Name: "Cebuano", ISO1: "ceb", ISO3: []string{"ceb_Latn"}},
	{Name: "Czech", ISO1: "cs", ISO3: []string{"ces_Latn"}},
	{Name: "Chokwe", ISO1: "cjk", ISO3: []string{"cjk_Latn"}},
	{Name: "Central Kurdish", ISO1: "ckb", ISO3: []string{"ckb_Arab"}},
	{Name: "Crimean Tatar", ISO1: "crh", ISO3: []string{"crh_Latn"}},
	{Name: "Welsh", ISO1: "cy", ISO3: []string{"cym_Latn"}},
	{Name: "Danish", ISO1: "da", ISO3: []string{"dan_Latn"}},
	{Name: "German", ISO1: "de", ISO3: []string{"deu_Latn"}},
	{Name: "Southwestern Dinka", ISO1: "dik", ISO3: []string{"dik_Latn"}},
	{Name: "Dyula", ISO1: "dyu", ISO3: []string{"dyu_Latn"}},
	{Name: "Dzongkha", ISO1: "dz", ISO3: []string{"dzo_Tibt"}},
	{Name: "Greek", ISO1: "el", ISO3: []string{"ell_Grek"}},
	{Name: "English", ISO1: "en", ISO3: []string{"eng_Latn"}},
	{Name: "Esperanto", ISO1: "eo", ISO3: []string{"epo_Latn"}},
	{Name: "Estonian", ISO1: "et", ISO3: []string{"est_Latn"}},
	{Name: "Basque", ISO1: "eu", ISO3: []string{"eus_Latn"}},
	{Name: "Ewe", ISO1: "ee", ISO3: []string{"ewe_Latn"}},
	{Name: "Faroese", ISO1: "fo", ISO3: []string{"fao_Latn"}},
	{Name: "Western Persian", ISO1: "pes", ISO3: []string{"pes_Arab"}},
	{Name: "Fijian", ISO1: "fj", ISO3: []string{"fij_Latn"}},
	{Name: "Finnish", ISO1: "fi", ISO3: []string{"fin_Latn"}},
	{Name: "Fon", ISO1: "fon", ISO3: []string{"fon_Latn"}},
	{Name: "French", ISO1: "fr", ISO3: []string{"fra_Latn"}},
	{Name: "Friulian", ISO1: "fur", ISO3: []string{"fur_Latn"}},
	{Name: "Nigerian Fulfulde", ISO1: "fuv", ISO3: []string{"fuv_Latn"}},
	{Name: "Scottish Gaelic", ISO1: "gd", ISO3: []string{"gla_Latn"}},
	{Name: "Irish", ISO1: "ga", ISO3: []string{"gle_Latn"}},
	{Name: "Galician", ISO1: "gl", ISO3: []string{"glg_Latn"}},
	{Name: "Guarani", ISO1: "gn", ISO3: []string{"grn_Latn"}},
	{Name: "Gujarati", ISO1: "gu", ISO3: []string{"guj_Gujr"}},
	{Name: "Haitian Creole", ISO1: "ht", ISO3: []string{"hat_Latn"}},
	{Name: "Hausa", ISO1: "ha", ISO3: []string{"hau_Latn"}},
	{Name: "Hebrew", ISO1: "he", ISO3: []string{"heb_Hebr"}},
	{Name: "Hindi", ISO1: "hi", ISO3: []string{"hin_Deva"}},
	{Name: "Chhattisgarhi", ISO1: "hne", ISO3: []string{"hne_Deva"}},
	{Name: "Croatian", ISO1: "hr", ISO3: []string{"hrv_Latn"}},
	{Name: "Hungarian", ISO1: "hu", ISO3: []string{"hun_Latn"}},
	{Name: "Armenian", ISO1: "hy", ISO3: []string{"hye_Armn"}},
	{Name: "Igbo", ISO1: "ig", ISO3: []string{"ibo_Latn"}},
	{Name: "Ilocano", ISO1: "ilo", ISO3: []string{"ilo_Latn"}},
	{Name: "Indonesian", ISO1: "id", ISO3: []string{"ind_Latn"}},
	{Name: "Icelandic", ISO1: "is", ISO3: []string{"isl_Latn"}},
	{Name: "Italian", ISO1: "it", ISO3: []string{"ita_Latn"}},
	{Name: "Javanese", ISO1: "jv", ISO3: []string{"jav_Latn"}},
	{Name: "Japanese", ISO1: "ja", ISO3: []string{"jpn_Jpan"}},
	{Name: "Kabyle", ISO1: "kab", ISO3: []string{"kab_Latn"}},
	{Name: "Jingpho", ISO1: "kac", ISO3: []string{"kac_Latn"}},
	{Name: "Kamba", ISO1: "kam", ISO3: []string{"kam_Latn"}},
	{Name: "Kannada", ISO1: "kn", ISO3: []string{"kan_Knda"}},
	{Name: "Kashmiri", ISO1: "ks", ISO3: []string{"kas_Arab", "kas_Deva"}},
	{Name: "Georgian", ISO1: "ka", ISO3: []string{"kat_Geor"}},
	{Name: "Central Kanuri", ISO1: "knc", ISO3: []string{"knc_Arab", "knc_Latn"}},
	{Name: "Kazakh", ISO1: "kk", ISO3: []string{"kaz_Cyrl"}},
	{Name: "Kabiyè", ISO1: "kbp", ISO3: []string{"kbp_Latn"}},
	{Name: "Kabuverdianu", ISO1: "kea", ISO3: []string{"kea_Latn"}},
	{Name: "Khmer", ISO1: "km", ISO3: []string{"khm_Khmr"}},
	{Name: "Kikuyu", ISO1: "ki", ISO3: []string{"kik_Latn"}},
	{Name: "Kinyarwanda", ISO1: "rw", ISO3: []string{"kin_Latn"}},
	{Name: "Kyrgyz", ISO1: "ky", ISO3: []string{"kir_Cyrl"}},
	{Name: "Kimbundu", ISO1: "kmb", ISO3: []string{"kmb_Latn"}},
	{Name: "Kikongo", ISO1: "kg", ISO3: []string{"kon_Latn"}},
	{Name: "Korean", ISO1: "ko", ISO3: []string{"kor_Hang"}},
	{Name: "Northern Kurdish", ISO1: "kmr", ISO3: []string{"kmr_Latn"}},
	{Name: "Lao", ISO1: "lo", ISO3: []string{"lao_Laoo"}},
	{Name: "Standard Latvian", ISO1: "lvs", ISO3: []string{"lvs_Latn"}},
	{Name: "Ligurian", ISO1: "lij", ISO3: []string{"lij_Latn"}},
	{Name: "Limburgish", ISO1: "li", ISO3: []string{"lim_Latn"}},
	{Name: "Lingala", ISO1: "ln", ISO3: []string{"lin_Latn"}},
	{Name: "Lithuanian", ISO1: "lt", ISO3: []string{"lit_Latn"}},
	{Name: "Lombard", ISO1: "lmo", ISO3: []string{"lmo_Latn"}},
	{Name: "Latgalian", ISO1: "ltg", ISO3: []string{"ltg_Latn"}},
	{Name: "Luxembourgish", ISO1: "lb", ISO3: []string{"ltz_Latn"}},
	{Name: "Luba-Kasai", ISO1: "lua", ISO3: []string{"lua_Latn"}},
	{Name: "Ganda", ISO1: "lg", ISO3: []string{"lug_Latn"}},
	{Name: "Luo", ISO1: "luo", ISO3: []string{"luo_Latn"}},
	{Name: "Mizo", ISO1: "lus", ISO3: []string{"lus_Latn"}},
	{Name: "Magahi", ISO1: "mag", ISO3: []string{"mag_Deva"}},
	{Name: "Maithili", ISO1: "mai", ISO3: []string{"mai_Deva"}},
	{Name: "Malayalam", ISO1: "ml", ISO3: []string{"mal_Mlym"}},
	{Name: "Marathi", ISO1: "mr", ISO3: []string{"mar_Deva"}},
	{Name: "Minangkabau", ISO1: "min", ISO3: []string{"min_Arab", "min_Latn"}},
	{Name: "Macedonian", ISO1: "mk", ISO3: []string{"mkd_Cyrl"}},
	{Name: "Plateau Malagasy", ISO1: "plt", ISO3: []string{"plt_Latn"}},
	{Name: "Maltese", ISO1: "mt", ISO3: []string{"mlt_Latn"}},
	{Name: "Meitei", ISO1: "mni", ISO3: []string{"mni_Beng"}},
	{Name: "Halh Mongolian", ISO1: "khk", ISO3: []string{"khk_Cyrl"}},
	{Name: "Mossi", ISO1: "mos", ISO3: []string{"mos_Latn"}},
	{Name: "Maori", ISO1: "mi", ISO3: []string{"mri_Latn"}},
	{Name: "Standard Malay", ISO1: "zsm", ISO3: []string{"zsm_Latn"}},
	{Name: "Burmese", ISO1: "my", ISO3: []string{"mya_Mymr"}},
	{Name: "Dutch", ISO1: "nl", ISO3: []string{"nld_Latn"}},
	{Name: "Norwegian Nynorsk", ISO1: "nn", ISO3: []string{"nno_Latn"}},
	{Name: "Norwegian Bokmål", ISO1: "nb", ISO3: []string{"nob_Latn"}},
	{Name: "Nepali", ISO1: "npi", ISO3: []string{"npi_Deva"}},
	{Name: "Northern Sotho", ISO1: "nso", ISO3: []string{"nso_Latn"}},
	{Name: "Nuer", ISO1: "nus", ISO3: []string{"nus_Latn"}},
	{Name: "Nyanja", ISO1: "ny", ISO3: []string{"nya_Latn"}},
	{Name: "Occitan", ISO1: "oc", ISO3: []string{"oci_Latn"}},
	{Name: "West Central Oromo", ISO1: "gaz", ISO3: []string{"gaz_Latn"}},
	{Name: "Odia", ISO1: "ory", ISO3: []string{"ory_Orya"}},
	{Name: "Pangasinan", ISO1: "pag", ISO3: []string{"pag_Latn"}},
	{Name: "Eastern Panjabi", ISO1: "pa", ISO3: []string{"pan_Guru"}},
	{Name: "Papiamento", ISO1: "pap", ISO3: []string{"pap_Latn"}},
	{Name: "Polish", ISO1: "pl", ISO3: []string{"pol_Latn"}},
	{Name: "Portuguese", ISO1: "pt", ISO3: []string{"por_Latn"}},
	{Name: "Dari", ISO1: "prs", ISO3: []string{"prs_Arab"}},
	{Name: "Southern Pashto", ISO1: "pbt", ISO3: []string{"pbt_Arab"}},
	{Name: "Ayacucho Quechua", ISO1: "quy", ISO3: []string{"quy_Latn"}},
	{Name: "Romanian", ISO1: "ro", ISO3: []string{"ron_Latn"}},
	{Name: "Rundi", ISO1: "rn", ISO3: []string{"run_Latn"}},
	{Name: "Russian", ISO1: "ru", ISO3: []string{"rus_Cyrl"}},
	{Name: "Sango", ISO1: "sg", ISO3: []string{"sag_Latn"}},
	{Name: "Sanskrit", ISO1: "sa", ISO3: []string{"san_Deva"}},
	{Name: "Santali", ISO1: "sat", ISO3: []string{"sat_Beng"}},
	{Name: "Sicilian", ISO1: "scn", ISO3: []string{"scn_Latn"}},
	{Name: "Shan", ISO1: "shn", ISO3: []string{"shn_Mymr"}},
	{Name: "Sinhala", ISO1: "si", ISO3: []string{"sin_Sinh"}},
	{Name: "Slovak", ISO1: "sk", ISO3: []string{"slk_Latn"}},
	{Name: "Slovenian", ISO1: "sl", ISO3: []string{"slv_Latn"}},
	{Name: "Samoan", ISO1: "sm", ISO3: []string{"smo_Latn"}},
	{Name: "Shona", ISO1: "sn", ISO3: []string{"sna_Latn"}},
	{Name: "Sindhi", ISO1: "sd", ISO3: []string{"snd_Arab"}},
	{Name: "Somali", ISO1: "so", ISO3: []string{"som_Latn"}},
	{Name: "Southern Sotho", ISO1: "st", ISO3: []string{"sot_Latn"}},
	{Name: "Spanish", ISO1: "es", ISO3: []string{"spa_Latn"}},
	{Name: "Tosk Albanian", ISO1: "als", ISO3: []string{"als_Latn"}},
	{Name: "Sardinian", ISO1: "sc", ISO3: []string{"srd_Latn"}},
	{Name: "Serbian", ISO1: "sr", ISO3: []string{"srp_Cyrl"}},
	{Name: "Swati", ISO1: "ss", ISO3: []string{"ssw_Latn"}},
	{Name: "Sundanese", ISO1: "su", ISO3: []string{"sun_Latn"}},
	{Name: "Swedish", ISO1: "sv", ISO3: []string{"swe_Latn"}},
	{Name: "Swahili", ISO1: "swh", ISO3: []string{"swh_Latn"}},
	{Name: "Silesian", ISO1: "szl", ISO3: []string{"szl_Latn"}},
	{Name: "Tamil", ISO1: "ta", ISO3: []string{"tam_Taml"}},
	{Name: "Tatar", ISO1: "tt", ISO3: []string{"tat_Cyrl"}},
	{Name: "Telugu", ISO1: "te", ISO3: []string{"tel_Telu"}},
	{Name: "Tajik", ISO1: "tg", ISO3: []string{"tgk_Cyrl"}},
	{Name: "Tagalog", ISO1: "tl", ISO3: []string{"tgl_Latn"}},
	{Name: "Thai", ISO1: "th", ISO3: []string{"tha_Thai"}},
	{Name: "Tigrinya", ISO1: "ti", ISO3: []string{"tir_Ethi"}},
	{Name: "Tamasheq", ISO1: "taq", ISO3: []string{"taq_Tfng", "taq_Latn"}},
	{Name: "Tok Pisin", ISO1: "tpi", ISO3: []string{"tpi_Latn"}},
	{Name: "Tswana", ISO1: "tn", ISO3: []string{"tsn_Latn"}},
	{Name: "Tsonga", ISO1: "ts", ISO3: []string{"tso_Latn"}},
	{Name: "Turkmen", ISO1: "tk", ISO3: []string{"tuk_Latn"}},
	{Name: "Tumbuka", ISO1: "tum", ISO3: []string{"tum_Latn"}},
	{Name: "Turkish", ISO1: "tr", ISO3: []string{"tur_Latn"}},
	{Name: "Twi", ISO1: "tw", ISO3: []string{"twi_Latn"}},
	{Name: "Central Atlas Tamazight", ISO1: "tzm", ISO3: []string{"tzm_Tfng"}},
	{Name: "Uyghur", ISO1: "ug", ISO3: []string{"uig_Arab"}},
	{Name: "Ukrainian", ISO1: "uk", ISO3: []string{"ukr_Cyrl"}},
	{Name: "Umbundu", ISO1: "umb", ISO3: []string{"umb_Latn"}},
	{Name: "Urdu", ISO1: "ur", ISO3: []string{"urd_Arab"}},
	{Name: "Northern Uzbek", ISO1: "uzn", ISO3: []string{"uzn_Latn"}},
	{Name: "Venetian", ISO1: "vec", ISO3: []string{"vec_Latn"}},
	{Name: "Vietnamese", ISO1: "vi", ISO3: []string{"vie_Latn"}},
	{Name: "Waray", ISO1: "war", ISO3: []string{"war_Latn"}},
	{Name: "Wolof", ISO1: "wo", ISO3: []string{"wol_Latn"}},
	{Name: "Xhosa", ISO1: "xh", ISO3: []string{"xho_Latn"}},
	{Name: "Eastern Yiddish", ISO1: "ydd", ISO3: []string{"ydd_Hebr"}},
	{Name: "Yoruba", ISO1: "yo", ISO3: []string{"yor_Latn"}},
	{Name: "Yue Chinese", ISO1: "yue", ISO3: []string{"yue_Hant"}},
	{Name: "Chinese", ISO1: "zh", ISO3: []string{"zho_Hans", "zho_Hant"}},
	{Name: "Zulu", ISO1: "zu", ISO3: []string{"zul_Latn"}},
}}

func (languageMapping *LanguageMap) GetISO1() string {
	return languageMapping.ISO1
}

func (languageMapping *LanguageMap) GetISO3() []string {
	return languageMapping.ISO3
}

func (languageMapping *LanguageMap) GetName() string {
	return languageMapping.Name
}

func (languageMapping *LanguageMapping) GetISO3(iso1 string) []string {
	for _, languageMap := range languageMapping.LanguageMappings {
		if languageMap.ISO1 == iso1 {
			return languageMap.GetISO3()
		}
	}
	return []string{}
}

func (languageMapping *LanguageMapping) GetISO1(iso3 string) string {
	for _, languageMap := range languageMapping.LanguageMappings {
		for _, languageIso3 := range languageMap.ISO3 {
			if languageIso3 == iso3 {
				return languageMap.GetISO1()
			}
		}
	}
	return ""
}

func (languageMapping *LanguageMapping) GetName(iso1or3 string) string {
	for _, languageMap := range languageMapping.LanguageMappings {
		if languageMap.ISO1 == iso1or3 {
			return languageMap.GetName()
		} else {
			for _, languageIso3 := range languageMap.ISO3 {
				if languageIso3 == iso1or3 {
					return languageMap.GetName()
				}
			}
		}
	}
	return ""
}
