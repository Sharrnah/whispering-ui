# Realtime Mode Configuration

Realtime mode will display intermediate transcriptions and translations as they are being processed.

![realtime-profile-option.png](images%2Frealtime-profile-option.png)

This way you can already see results while the person is still speaking.

## Improving Realtime Speed
To improve the speed of the realtime mode, you can use the following advanced Settings (under the **Advanced -> Settings** tab):

| Setting                         | Description                                                                                                                                                                                                                                          |
|---------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `realtime_frame_multiply`       | Defines the minimum frame size of audio snippets.<br>_The default value should be fine for most cases._                                                                                                                                              |
| `realtime_frequency_time`       | Defines the time frequency (in seconds as fraction) when the recorded audio snippets are processed.<br>_Settings this too low can have a negative effect on realtime speed as the GPU might not be able to keep up with the speed._                  |
| `realtime_whisper_model`        | Optionally defines a separate Whisper model for realtime processing.<br>_Can improve speed in exchange for additional RAM usage and less precision._                                                                                                 |
| `realtime_whisper_precision`    | Only works when `realtime_whisper_model` is set.<br>_Defines the precision of the Whisper model._                                                                                                                                                    |
| `realtime_whisper_beam_size`    | Defines the beam size of the Whisper model.<br>_Can improve speed in exchange for less precision. (lower = faster)_                                                                                                                                  |
| `realtime_temperature_fallback` | Enables or Disables the temperature fallback when processing Realtime audio.<br>_Disabling this can speed up the processing and reduce stuck processing._                                                                                            |
| `condition_on_previous_text`    | Enables or Disables the condition on previous text.<br>_Disabling this can speed up the processing and reduce stuck processing while reducing relation of later results to previous transcripts. (is used for both realtime and regular processing)_ |


![realtime-settings.png](images%2Frealtime-settings.png)

Save the Settings by scrolling down and clicking on **Save**.

All realtime settings are applied immediately without requiring a restart
except the settings `realtime_whisper_model` and `realtime_whisper_precision` which require a restart.
