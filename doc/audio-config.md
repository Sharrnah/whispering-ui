# Audio configuration (TTS to Mic, Game Audio translation, etc.)

## Windows

### Text-to-Speech as Microphone input for Voice Chat (e.g. VRChat, Discord, etc.)
For routing Text-to-Speech as Microphone input, you need to install [VB-Audio Virtual Cable](https://vb-audio.com/Cable/)
- select it as `Audio Output (speaker)` in Whispering Tiger
- select it in the Game as Microphone (or as default Microphone in Windows) accordingly.
- If you want to listen to it at the same time, enable `tts_use_secondary_playback` in Advanced -> Settings. `tts_secondary_playback_device` = `-1` will play it on your windows default device.

### Game Audio as Microphone Input in Whispering Tiger (e.g. for Translation)
For routing Game Audio as Input to Whispering Tiger, you need to route the PC Sound to a Microphone Input.
  
  You have multiple options:
  - While using the WASAPI Audio API, you can select your PC Audio Device as Loopback Device in Whispering Tiger.
    ![audio-devices-loopback.png](images%2Fsetup%2Faudio-devices-loopback.png)


  - Install [VB-Audio Virtual Cable](https://vb-audio.com/Cable/) and select it as Audio Output in Windows.
    - activate Stereo Mix / Listen to this device for the same Cable in the Recording Tab and select your real loudspeaker in the `Playback through this device` (to still hear the audio yourself)
    ![how+to+playback+mic.png](images%2Fhow%2Bto%2Bplayback%2Bmic.png)

  
  - Install and use [Steelseries Sonar](https://steelseries.com/gg/sonar/download) (easier to use), [VoiceMeeter](https://vb-audio.com/Voicemeeter/) (difficult to use) or other similar applications for audio routing.
    
  
  And finally select the audio device in Whispering Tiger as `Audio Input (mic)` accordingly.


## Linux

Use **PulseAudio** in the profile for PulseAudio or PipeWire's PulseAudio compatibility service.

- **Desktop/game audio as input:** select the monitor source of the output device playing that audio. If needed, use your system's audio routing controls to connect the application's recording stream to that monitor.
- **TTS into voice chat:** route Whispering Tiger's output to a virtual sink, then select that sink's monitor as the voice-chat application's input. PipeWire routing tools can also connect the streams directly.
- **Listen locally as well:** enable `tts_use_secondary_playback` under Advanced -> Settings and select your headphones with `tts_secondary_playback_device`.

Keep the TTS output out of the source you are transcribing to avoid feeding generated speech back into recognition. Windows tools such as VB-Cable and WASAPI do not apply on Linux.

## Additional audio sources

Open **Additional Audio Sources** on the Speech-to-Text page to add another input, such as game audio alongside your microphone. Configure each source's outputs and translation separately. Windows also offers process loopback for an individual application's audio.

[Profile setup](../README.md#setup) · [Hardware and runtimes](hardware-support.md)
