# Integrated Text-to-Speech Models

## Silero-TTS
Silero-TTS is the simplest solution and supports different models for different languages like _English, Spanish, French, German and others_.
Silero-TTS supports SSML by which you can change the behaviour in the text.

Supported Tags are:
- break `<break time="2000ms" strength="x-weak"/>` where _time_ can be in milliseconds (**ms**) or seconds (**s**) and _strength_ can be **x-weak, weak, medium, strong, x-strong** 
- prosody `<prosody rate="x-slow" pitch="x-high">` where _rate_ can be **x-slow, slow, medium, fast, x-fast**, and _pitch_ can be **x-low, low, medium, high, x-high**
- p `<p>text</p>` Represents a paragraph, equivalent to x-strong pause.
- s `<s>text</s>` Represents a sentence, equivalent to strong pause.

## F5-TTS / E2-TTS
F5-TTS is a TTS Model that supports voice cloning based on an audio sample with fast inference and Multi-Style / Multi-Speaker Generation.

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
