package ProfileSettings

import (
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities/AudioAPI"
)

var DefaultProfileSetting = Settings.Conf{
	SettingsFilename:         "",
	Websocket_ip:             "127.0.0.1",
	Websocket_port:           5000,
	Run_backend:              true,
	Device_index:             -1,
	Device_out_index:         -1,
	Audio_api:                AudioAPI.AudioBackends[0].Name,
	Audio_input_device:       "",
	Audio_output_device:      "",
	Ai_device:                "cpu",
	Model:                    "tiny",
	Txt_translator:           "NLLB200_CT2",
	Txt_translator_size:      "small",
	Txt_translator_device:    "cpu",
	Txt_translator_precision: "float32",
	Txt_translate_realtime:   false,

	Txt_second_translation_enabled:   false,
	Txt_second_translation_languages: "eng_Latn",
	Txt_second_translation_wrap:      " | ",

	//Tts_enabled:      true,
	Tts_type:         "silero",
	Tts_ai_device:    "cpu",
	Current_language: "",

	Osc_ip:                             "127.0.0.1",
	Osc_port:                           9000,
	Osc_address:                        "/chatbox/input",
	Osc_min_time_between_messages:      1.5,
	Osc_typing_indicator:               true,
	Osc_convert_ascii:                  false,
	Osc_chat_limit:                     144,
	Osc_type_transfer:                  "translation_result",
	Osc_type_transfer_split:            " 🌐 ",
	Osc_send_type:                      "chunks",
	Osc_time_limit:                     15.0,
	Osc_scroll_time_limit:              1.5,
	Osc_initial_time_limit:             15.0,
	Osc_scroll_size:                    3,
	Osc_max_scroll_size:                30,
	Osc_delay_until_audio_playback:     false,
	Osc_delay_until_audio_playback_tag: "tts",
	Osc_delay_timeout:                  10.0,

	Osc_server_ip:   "127.0.0.1",
	Osc_server_port: 9001,
	Osc_sync_mute:   false,
	Osc_sync_afk:    false,

	Ocr_window_name: "VRChat",
	Ocr_lang:        "en",

	Logprob_threshold:   "-1.0",
	No_speech_threshold: "0.6",

	Vad_enabled:              true,
	Vad_on_full_clip:         false,
	Vad_confidence_threshold: 0.4,
	Vad_frames_per_buffer:    512,
	Vad_thread_num:           1,
	Push_to_talk_key:         "",

	Speaker_diarization:  false,
	Speaker_change_split: true,
	Min_speaker_length:   0.5,
	Min_speakers:         1,
	Max_speakers:         3,

	Denoise_audio:                "",
	Denoise_audio_post_filter:    false,
	Denoise_audio_before_trigger: false,

	Whisper_task:                  "transcribe",
	Whisper_precision:             "float32",
	Stt_type:                      "faster_whisper",
	Temperature_fallback:          true,
	Phrase_time_limit:             30.0,
	Pause:                         1.0,
	Energy:                        300,
	Beam_size:                     5,
	Length_penalty:                1.0,
	Beam_search_patience:          1.0,
	Repetition_penalty:            1.0,
	No_repeat_ngram_size:          0,
	Whisper_cpu_threads:           0,
	Whisper_num_workers:           1,
	Condition_on_previous_text:    false,
	Prompt_reset_on_temperature:   0.5,
	Realtime:                      false,
	Realtime_frame_multiply:       15,
	Realtime_frequency_time:       1.0,
	Realtime_whisper_model:        "",
	Realtime_whisper_precision:    "float32",
	Realtime_whisper_beam_size:    1,
	Realtime_temperature_fallback: false,
	Whisper_apply_voice_markers:   false,
	Max_sentence_repetition:       -1,
	Transcription_auto_save_file:  "",
	Thread_per_transcription:      true,

	Silence_cutting_enabled:   true,
	Silence_offset:            -40.0,
	Max_silence_length:        30.0,
	Keep_silence_length:       0.20,
	Normalize_enabled:         true,
	Normalize_lower_threshold: -24.0,
	Normalize_upper_threshold: -16.0,
	Normalize_gain_factor:     2.0,
}

var Presets = map[string]Settings.Conf{
	"": DefaultProfileSetting,
	"NVIDIA-HighPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cuda"
		profile.Model = "large-v3"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = true

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = true
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cuda"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float16"
		return profile
	}(),
	"NVIDIA-LowPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cuda"
		profile.Model = "medium"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = true

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = true
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cpu"
		profile.Txt_translator_size = "small"
		profile.Txt_translator_precision = "int8"
		return profile
	}(),
	"NVIDIA-HighPerformance-Realtime": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cuda"
		profile.Model = "large-v2"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 1
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "translate"
		profile.Realtime = true
		profile.Realtime_frame_multiply = 15
		profile.Realtime_frequency_time = 1.2
		profile.Realtime_temperature_fallback = false
		profile.Realtime_whisper_beam_size = 1
		profile.Repetition_penalty = 1.05
		profile.Thread_per_transcription = true
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = ""
		profile.Txt_translate_realtime = true

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 0.9
		profile.Phrase_time_limit = 10

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cuda"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float16"
		return profile
	}(),
	"NVIDIA-LowPerformance-Realtime": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cuda"
		profile.Model = "medium"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 1
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "translate"
		profile.Realtime = true
		profile.Realtime_frame_multiply = 15
		profile.Realtime_frequency_time = 1.2
		profile.Realtime_temperature_fallback = false
		profile.Realtime_whisper_beam_size = 1
		profile.Repetition_penalty = 1.05
		profile.Thread_per_transcription = true
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = ""
		profile.Txt_translate_realtime = true

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 0.9
		profile.Phrase_time_limit = 10

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cuda"
		profile.Txt_translator_size = "small"
		profile.Txt_translator_precision = "int8_float16"
		return profile
	}(),
	"AMDIntel-HighPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "transformer_whisper"
		profile.Ai_device = "direct-ml:0"
		profile.Model = "large-v3"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = true

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = true
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200"
		profile.Txt_translator_device = "direct-ml:0"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float16"
		return profile
	}(),
	"AMDIntel-LowPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "transformer_whisper"
		profile.Ai_device = "direct-ml:0"
		profile.Model = "medium"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = true

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = true
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cpu"
		profile.Txt_translator_size = "small"
		profile.Txt_translator_precision = "int8"
		return profile
	}(),
	"AMDIntel-HighPerformance-Realtime": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "transformer_whisper"
		profile.Ai_device = "direct-ml:0"
		profile.Model = "large-v2"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 1
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "translate"
		profile.Realtime = true
		profile.Realtime_frame_multiply = 15
		profile.Realtime_frequency_time = 1.2
		profile.Realtime_temperature_fallback = false
		profile.Realtime_whisper_beam_size = 1
		profile.Repetition_penalty = 1.05
		profile.Thread_per_transcription = true
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = ""
		profile.Txt_translate_realtime = true

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 0.9
		profile.Phrase_time_limit = 10

		profile.Txt_translator = "NLLB200"
		profile.Txt_translator_device = "direct-ml:0"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float16"
		return profile
	}(),
	"AMDIntel-LowPerformance-Realtime": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "transformer_whisper"
		profile.Ai_device = "direct-ml:0"
		profile.Model = "medium"
		profile.Whisper_precision = "float16"
		profile.Beam_size = 1
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "translate"
		profile.Realtime = true
		profile.Realtime_frame_multiply = 15
		profile.Realtime_frequency_time = 1.2
		profile.Realtime_temperature_fallback = false
		profile.Realtime_whisper_beam_size = 1
		profile.Repetition_penalty = 1.05
		profile.Thread_per_transcription = true
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = ""
		profile.Txt_translate_realtime = true

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 0.9
		profile.Phrase_time_limit = 10

		profile.Txt_translator = "NLLB200"
		profile.Txt_translator_device = "direct-ml:0"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float16"
		return profile
	}(),
	"CPU-HighPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cpu"
		profile.Model = "medium"
		profile.Whisper_precision = "float32"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cpu"
		profile.Txt_translator_size = "medium"
		profile.Txt_translator_precision = "float32"
		return profile
	}(),
	"CPU-LowPerformance-Accuracy": func() Settings.Conf {
		profile := DefaultProfileSetting

		profile.Stt_type = "faster_whisper"
		profile.Ai_device = "cpu"
		profile.Model = "small"
		profile.Whisper_precision = "float32"
		profile.Beam_size = 5
		profile.Condition_on_previous_text = false

		profile.Whisper_task = "transcribe"
		profile.Realtime = false
		profile.Repetition_penalty = 1.0
		profile.Thread_per_transcription = false
		profile.Whisper_cpu_threads = 5
		profile.Whisper_num_workers = 5
		profile.Word_timestamps = false

		profile.Tts_type = "silero"
		profile.Txt_translate_realtime = false

		profile.Temperature_fallback = false
		profile.Normalize_enabled = false
		profile.Pause = 1.4
		profile.Phrase_time_limit = 30

		profile.Txt_translator = "NLLB200_CT2"
		profile.Txt_translator_device = "cpu"
		profile.Txt_translator_size = "small"
		profile.Txt_translator_precision = "int8"
		return profile
	}(),
}
