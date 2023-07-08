# Realtime Mode Configuration

Realtime mode will display intermediate transcriptions and translations as they are being processed.

It is recommended to run the AI Model on a GPU (`CUDA`) for realtime mode.

![realtime-profile-option.png](images%2Frealtime-profile-option.png)

## Improving Realtime Speed
To improve the speed of the realtime mode, you can use the following advanced Settings (under the **Advanced -> Settings** tab):

| Setting                         | Description                                                                                                                                                                                                                                                                                                                                             |
|---------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `model`                         | Defines the Whisper model for processing.<br>_Can improve speed in exchange for additional RAM usage and less precision.<br>(if no `realtime_whisper_model` is used, this option is used for realtime processing as well.)_                                                                                                                             |
| `whisper_precision`             | Defines the precision of the Whisper model. _(`float16` should be faster then `float32` with CUDA)_                                                                                                                                                                                                                                                     |
| `stt_type`                      | `faster_whisper` is faster then `original_whisper` with the same accuracy.                                                                                                                                                                                                                                                                              |
| `temperature_fallback`          | Enables or Disables the temperature fallback when processing audio.<br>_Disabling this can speed up the processing and reduce stuck processing._                                                                                                                                                                                                        |
| `beam_size`                     | Defines the beam size of the Whisper model.<br>_Can improve speed in exchange for less precision. (lower = faster)_                                                                                                                                                                                                                                     |
| `condition_on_previous_text`    | Enables or Disables the condition on previous text.<br>_Disabling this can speed up the processing and reduce stuck processing while reducing relation of later results to previous transcripts. (is used for both realtime and regular processing)_                                                                                                    |
| `whisper_num_workers`           | Increasing this can _possibly_ speed up processing.                                                                                                                                                                                                                                                                                                     |
| `realtime_frame_multiply`       | Defines the minimum frame size of audio snippets.<br>_The default value should be fine for most cases._                                                                                                                                                                                                                                                 |
| `realtime_frequency_time`       | Defines the time frequency (in seconds as fraction) when the recorded audio snippets are processed.<br>_Settings this too low can have a negative effect on realtime speed as the GPU might not be able to keep up with the speed._<br><br>_Setting this to `0.5` for every 500 milliseconds can already work nicely with a good GPU and medium model._ |
| `realtime_whisper_beam_size`    | Defines the beam size of the Whisper model.<br>_Can improve speed in exchange for less precision. (lower = faster)_                                                                                                                                                                                                                                     |
| `realtime_temperature_fallback` | Enables or Disables the temperature fallback when processing Realtime audio.<br>_Disabling this can speed up the processing and reduce stuck processing._                                                                                                                                                                                               |

> **Note:**
> <br>
> Optionally you can also define a separate Whisper model for realtime processing by setting the following settings:
> 
> _(Its not really recommended though, as using this increases the used Memory, and using a smaller model only for realtime processing will increase differences between realtime results and the final result)_
> 
> | Setting                      | Description                                                                                                                                            |
> |--------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
> | `realtime_whisper_model`     | Optionally defines a separate Whisper model for realtime processing.<br>_Can improve speed in exchange for additional RAM usage and less precision._   |
> | `realtime_whisper_precision` | Only used when `realtime_whisper_model` is set.<br>_Defines the precision of the Whisper model. (`float16` should be faster then `float32` with CUDA)_ |

![realtime-settings.png](images%2Frealtime-settings.png)

Save the Settings by scrolling down and clicking on **Save**.

All realtime settings are applied immediately without requiring a restart
except the settings `realtime_whisper_model` and `realtime_whisper_precision` which require a restart.
