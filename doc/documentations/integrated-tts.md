# Integrated Text-to-Speech Models

## Content
- [Silero-TTS](#silero-tts)
- [Chatterbox-TTS](#chatterbox-tts)
  - [Add own voice](#add-own-voice)
  - [Generate Audio with multiple speakers](#generate-audio-with-multiple-speakers)
- [F5-TTS / E2-TTS](#f5-tts--e2-tts)
  - [Add own voice](#add-own-voice-1)
  - [Generate Audio with multiple speakers](#generate-audio-with-multiple-speakers-1)
- [Kokoro-TTS](#kokoro-tts)
- [Zonos-TTS](#zonos-tts)
  - [Add own voice](#add-own-voice-2)
- [Orpheus TTS](#orpheus-tts)

## Silero-TTS
Silero-TTS is the simplest solution and supports different models for different languages.
Silero-TTS supports SSML by which you can change the behaviour in the text.

<details>
  <summary>8 Supported Languages (Multiple models)</summary>
  <ul>
    <li>English</li>
    <li>German</li>
    <li>Russian</li>
    <li>Ukrainian</li>
    <li>Uzbek</li>
    <li>Indic</li>
    <li>Spanish</li>
    <li>French</li>
  </ul>
</details>

Demo:
<video src='https://github.com/user-attachments/assets/05803372-0e53-431a-a99e-e067de0e6982' width=300></video>

Supported Tags are:
- break `<break time="2000ms" strength="x-weak"/>` where _time_ can be in milliseconds (**ms**) or seconds (**s**) and _strength_ can be **x-weak, weak, medium, strong, x-strong** 
- prosody `<prosody rate="x-slow" pitch="x-high">` where _rate_ can be **x-slow, slow, medium, fast, x-fast**, and _pitch_ can be **x-low, low, medium, high, x-high**
- p `<p>text</p>` Represents a paragraph, equivalent to x-strong pause.
- s `<s>text</s>` Represents a sentence, equivalent to strong pause.

## Chatterbox-TTS
Chatterbox-TTS is a TTS Model that supports voice cloning based on an audio sample with fast inference and Multi-Style / Multi-Speaker Generation.

The custom implementation can auto-detect the language from the text. _(its recommended to set the language for better results)_
<details>
  <summary>23 Supported Languages (Single model)</summary>
  <ul>
    <li>Arabic</li>
    <li>Danish</li>
    <li>German</li>
    <li>Greek</li>
    <li>English</li>
    <li>Spanish</li>
    <li>Finnish</li>
    <li>French</li>
    <li>Hebrew</li>
    <li>Hindi</li>
    <li>Italian</li>
    <li>Japanese</li>
    <li>Korean</li>
    <li>Malay</li>
    <li>Dutch</li>
    <li>Norwegian</li>
    <li>Polish</li>
    <li>Portuguese</li>
    <li>Russian</li>
    <li>Swedish</li>
    <li>Swahili</li>
    <li>Turkish</li>
    <li>Chinese</li>
  </ul>
</details>

Demo:
<video src='https://github.com/user-attachments/assets/005cd29d-a53e-4e5c-b9be-632a95310f0e' width=300></video>

The speed is configured by the CFG / Pace setting.

### Add own voice
To add your own voice, go to the `.cache\chatterbox-tts-cache\voices` directory
- Copy a _.wav_ sample audio of the voice into it.

  Best results should be audio files as PCM S16 LE, Mono with a sample rate of 24000 Hz and 16 Bits per sample
  and a length of ~ 12 seconds.

- If an audio file does not give good results. Sometimes it also helps to cut the audio shorter.
- The TTS result will take over the accent of the sample audio provided. You can reduce accent strength by lowering temperature.

### Generate Audio with multiple speakers
To generate audio with different speakers, you can add the Speaker name at the beginning of a line like this:
```
[Justin] This is the text, spoken by the Justin speaker.
[Announcer_Ahri] And this text will be spoken by the Announcer_Ahri voice.
or
[Justin]
This is the text, spoken by the Justin speaker.
[Announcer_Ahri]
And this text will be spoken by the Announcer_Ahri voice.
```

## F5-TTS / E2-TTS
F5-TTS is a TTS Model that supports voice cloning based on an audio sample with fast inference and Multi-Style / Multi-Speaker Generation.

<details>
  <summary>11 Supported Languages (Multiple models with language combinations)</summary>
  <ul>
    <li>English & Chinese</li>
    <li>French</li>
    <li>German</li>
    <li>Greek</li>
    <li>Italian</li>
    <li>Japanese</li>
    <li>Spanish</li>
    <li>Russian</li>
    <li>Vietnamese</li>
    <li>Malaysian</li>
  </ul>
</details>

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

<details>
  <summary>8 Supported Languages (Single model)</summary>
  <ul>
    <li>English (US)</li>
    <li>English (British)</li>
    <li>Spanish</li>
    <li>French</li>
    <li>Hindi</li>
    <li>Italian</li>
    <li>Japanese</li>
    <li>Brazilian Portuguese</li>
    <li>Chinese</li>
  </ul>
</details>

Demo:
<video src='https://github.com/user-attachments/assets/8bd6ecb8-1f67-4b97-abac-dc218d8590fa' width=300></video>

## Zonos-TTS
Zonos-TTS is a TTS Model that supports voice cloning based on an audio samples.

<details>
  <summary>~91 Supported Languages (Single model) not counting variants</summary>
  <ul>
    <li>Afrikaans</li>
    <li>Amharic</li>
    <li>Aragonese</li>
    <li>Assamese</li>
    <li>Azerbaijani</li>
    <li>Bashkir</li>
    <li>Bulgarian</li>
    <li>Bengali</li>
    <li>Bishnupriya Manipuri</li>
    <li>Bosnian</li>
    <li>Catalan</li>
    <li>Chinese (Mandarin)</li>
    <li>Welsh</li>
    <li>Danish</li>
    <li>German</li>
    <li>English (Caribbean)</li>
    <li>English (UK)</li>
    <li>English (Scotland)</li>
    <li>English (GB Clan)</li>
    <li>English (GB CWMD)</li>
    <li>English (RP)</li>
    <li>English (US)</li>
    <li>Esperanto</li>
    <li>Spanish</li>
    <li>Spanish (Latin America)</li>
    <li>Estonian</li>
    <li>Basque</li>
    <li>Persian</li>
    <li>Persian (Latin)</li>
    <li>Finnish</li>
    <li>French (Belgium)</li>
    <li>French (Switzerland)</li>
    <li>French (France)</li>
    <li>Irish</li>
    <li>Scottish Gaelic</li>
    <li>Guarani</li>
    <li>Ancient Greek</li>
    <li>Gujarati</li>
    <li>Hakka</li>
    <li>Croatian</li>
    <li>Haitian Creole</li>
    <li>Hungarian</li>
    <li>Armenian</li>
    <li>Western Armenian</li>
    <li>Interlingua</li>
    <li>Indonesian</li>
    <li>Icelandic</li>
    <li>Italian</li>
    <li>Japanese</li>
    <li>Lojban</li>
    <li>Georgian</li>
    <li>Kazakh</li>
    <li>Kalaallisut</li>
    <li>Kannada</li>
    <li>Korean</li>
    <li>Konkani</li>
    <li>Kurdish</li>
    <li>Kyrgyz</li>
    <li>Latin</li>
    <li>Lingua Franca Nova</li>
    <li>Lithuanian</li>
    <li>Latvian</li>
    <li>Māori</li>
    <li>Macedonian</li>
    <li>Malayalam</li>
    <li>Marathi</li>
    <li>Maltese</li>
    <li>Burmese</li>
    <li>Norwegian Bokmål</li>
    <li>Classical Nahuatl</li>
    <li>Nepali</li>
    <li>Dutch</li>
    <li>Oromo</li>
    <li>Oriya</li>
    <li>Punjabi</li>
    <li>Papiamento</li>
    <li>Polish</li>
    <li>Portuguese</li>
    <li>Portuguese (Brazil)</li>
    <li>Paraguayan Guarani</li>
    <li>K'iche'</li>
    <li>Romanian</li>
    <li>Russian</li>
    <li>Russian (Latvia)</li>
    <li>Sindhi</li>
    <li>Shan</li>
    <li>Sinhala</li>
    <li>Slovak</li>
    <li>Slovenian</li>
    <li>Albanian</li>
    <li>Serbian</li>
    <li>Swedish</li>
    <li>Swahili</li>
    <li>Tamil</li>
    <li>Telugu</li>
    <li>Tswana</li>
    <li>Turkish</li>
    <li>Tatar</li>
    <li>Urdu</li>
    <li>Uzbek</li>
    <li>Vietnamese</li>
    <li>Vietnamese (Central)</li>
    <li>Vietnamese (South)</li>
    <li>Cantonese</li>
  </ul>
</details>

Demo:
<video src='https://github.com/user-attachments/assets/9e7121e5-6321-47d0-a3b2-e99d2dba46ed' width=300></video>

The model is also multi-lingual and supports emotion settings.

### Add own voice
To add your own voice, go to the `.cache\zonos-tts-cache\voices` directory
- Copy a _.wav_ sample audio of the voice into it.

  Best results should be audio files as PCM S16 LE, Mono with a sample rate of 24000 Hz and 16 Bits per sample.

## Orpheus TTS
Orpheus TTS is a TTS Model that supports natural intonation, emotion, and rhythm.

<details>
  <summary>1 Supported Language (Single model)</summary>
  <ul>
    <li>English</li>
  </ul>
</details>

Demo:
<video src='https://github.com/user-attachments/assets/a5d2b890-4b98-4d08-b5bb-5f3fd24fc51c' width=300></video>

Tags to control speech and emotion characteristics:

`<laugh>`, `<chuckle>`, `<sigh>`, `<cough>`, `<sniffle>`, `<groan>`, `<yawn>`, `<gasp>`
