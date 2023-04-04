# <img src=app-icon.png width=90> Whispering Tiger UI (Live Translate/Transcribe)

Whispering Tiger UI is a **Native-UI** that can be used to control the **Whispering Tiger** application.

[Whispering Tiger](https://github.com/Sharrnah/whispering) is a free and Open-Source tool that can listen/watch to any **audio stream** or **in-game image** on your machine and prints out the transcription or translation
to a web browser using Websockets or over OSC (examples are **Streaming-overlays** or **VRChat**).

## Intended Use
Whispering Tiger UI is intended to make it easier to control and configure the **Whispering Tiger** application.

## Features
- **Native-UI** for Windows (and possibly Linux in the future)
- **Easy to use** for both **beginners** and **advanced users**
- **Access to all Whispering Tiger features**, which includes:
   - Transcription / Translation of audio streams
   - Recognition and Translation of in-game images
   - Displaying the results in a web browser or VRChat, using Websockets or OSC
- **Save** and **load** configurations
- **Preview** if your selected Audio devices are working
- **Plugin** support for **additional features** (e.g. **Large Language Models**, **Emotion Prediction** or **Currently Playing Song** Plugins)
  - [Find a list of Plugins here.](https://github.com/Sharrnah/whispering/blob/main/documentation/plugins.md) 
- **Auto-Update** to the latest version of **Whispering Tiger**.

## Download
[**Download Latest Version**](https://github.com/Sharrnah/whispering-ui/releases/latest)

## Installation
1. After the download the latest version from the [**Releases**], extract it to a folder of your choice on a drive with enough free space.
2. [Install CUDA for GPU Acceleration](https://developer.nvidia.com/cuda-downloads) (Optional but recommended for NVIDIA GPUs).
3. Run the **Whispering Tiger.exe** file.
4. Let it download the latest version of **Whispering Tiger**. (It will ask to download an Update.)
5. After the download is finished, you can create a Profile and start using the **Whispering Tiger** application.
   - On the first start, it will start downloading the A.I. Models which can take a while depending on your selected Model size. (currently it does not show the status of the model downloads)

## Screenshots
<img src=doc/images/profile-selection.png width=845 alt="profile selection">
<img src=doc/images/speech2text.png width=845 alt="Speech 2 Text Tab">
<img src=doc/images/text-translate.png width=845 alt="Text Translate Tab">
<img src=doc/images/text2speech.png width=845 alt="Text 2 Speech Tab">
<img src=doc/images/ocr.png width=845 alt="Optical Character Recognition Tab">
<img src=doc/images/plugins.png width=845 alt="Plugins Tab">
<img src=doc/images/advanced-settings.png width=845 alt="Advanced Settings Tab">
<img src=doc/images/about.png width=845 alt="About Info Tab">
