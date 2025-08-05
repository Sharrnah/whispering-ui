# Integrated Text-to-Speech Models

## Silero-TTS
Silero-TTS is the simplest solution and supports different models for different languages like _English, Spanish, French, German and others_.
Silero-TTS supports SSML by which you can change the behaviour in the text.

Demo:
<video src='https://github.com/user-attachments/assets/05803372-0e53-431a-a99e-e067de0e6982' width=300></video>

Supported Tags are:
- break `<break time="2000ms" strength="x-weak"/>` where _time_ can be in milliseconds (**ms**) or seconds (**s**) and _strength_ can be **x-weak, weak, medium, strong, x-strong** 
- prosody `<prosody rate="x-slow" pitch="x-high">` where _rate_ can be **x-slow, slow, medium, fast, x-fast**, and _pitch_ can be **x-low, low, medium, high, x-high**
- p `<p>text</p>` Represents a paragraph, equivalent to x-strong pause.
- s `<s>text</s>` Represents a sentence, equivalent to strong pause.

## F5-TTS / E2-TTS
F5-TTS is a TTS Model that supports voice cloning based on an audio sample with fast inference and Multi-Style / Multi-Speaker Generation.

Demo:
<video src='https://github.com/user-attachments/assets/eac658cc-13aa-482d-93a8-fb38ca410dbc' width=300></video>

The speed can be set globally in the Settings.

### Add own voice
To add your own voice, go to the `.cache\f5tts-cache\voices` directory
- Copy a _.wav_ sample audio of the voice with a _.txt_ file with the same name containing the transcript of the spoken text into it.
  
  Best results should be audio files as PCM S16 LE, Mono with a sample rate of 24000 Hz and 16 Bits per sample.

- If an audio file does not give good results, make sure the transcript is good. Sometimes it also helps to cut the audio shorter.

### Generate Audio with multiple speakers
To generate audio with different speakers, you can add the Speaker name at the beginning of a line like this:
```
[Justin] This is the text, spoken by the Justin speaker.
[Announcer_Ahri] And this text will be spoken by the Announcer_Ahri voice.
```

## Kokoro-TTS
Kokoro-TTS is a multi-lingual TTS Model that supports fast inference with high quality.

Demo:
<video src='https://github.com/user-attachments/assets/8bd6ecb8-1f67-4b97-abac-dc218d8590fa' width=300></video>

## Zonos-TTS
Zonos-TTS is a TTS Model that supports voice cloning based on an audio samples.

Demo:
<video src='https://github.com/user-attachments/assets/b1641287-43c9-4b96-946a-3f62112037f2' width=300></video>

The model is also multi-lingual and supports emotion settings.

### Add own voice
To add your own voice, go to the `.cache\zonos-tts-cache\voices` directory
- Copy a _.wav_ sample audio of the voice into it.

  Best results should be audio files as PCM S16 LE, Mono with a sample rate of 24000 Hz and 16 Bits per sample.

## Orpheus TTS
Orpheus TTS is a TTS Model that supports natural intonation, emotion, and rhythm.

Demo:
<video src='https://github.com/user-attachments/assets/a5d2b890-4b98-4d08-b5bb-5f3fd24fc51c' width=300></video>

Tags to control speech and emotion characteristics:

`<laugh>`, `<chuckle>`, `<sigh>`, `<cough>`, `<sniffle>`, `<groan>`, `<yawn>`, `<gasp>`




