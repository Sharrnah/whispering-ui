# Hardware and AI runtimes

Whispering Tiger runs on Windows and Linux. GPU support depends on the selected **runtime and model**, not just the GPU brand. STT, translation, TTS and OCR can use different devices in the same profile.

| Runtime | Windows | Linux | Used for |
| --- | --- | --- | --- |
| audio.cpp (GGUF) | CPU, NVIDIA CUDA, Vulkan on AMD/Intel/NVIDIA | CPU, Vulkan on AMD/Intel/NVIDIA | Selected STT and TTS models |
| PyTorch / Transformers | CPU, NVIDIA CUDA for compatible models | CPU, NVIDIA CUDA for compatible models | STT, translation, TTS and OCR, depending on model |
| CTranslate2 | CPU, NVIDIA CUDA | CPU, NVIDIA CUDA | Faster Whisper and Faster NLLB-200 |

## AMD and Intel GPUs

Select **audio.cpp (native GGUF runtime)** as the STT or TTS type, choose a model, then select **Vulkan** and your GPU. Install your graphics driver's Vulkan support; a Vulkan SDK is not needed.

This does not enable Vulkan for PyTorch, CTranslate2, text translation or OCR. Use a supported device for each of those tasks, such as CPU. Available precisions and speed depend on the model, GPU and driver.

## NVIDIA GPUs

Use **CUDA** where the selected runtime supports it. On Linux, the managed audio.cpp runtime offers **Vulkan or CPU**; other compatible models can use the packaged CUDA backend. The Linux CUDA package includes its runtime libraries and needs a compatible NVIDIA GPU and driver; follow the package's installation notes.

## CPU and memory

CPU operation is useful without a supported GPU, but large models can be slow. Start with a small model and watch the profile's memory estimate. Loading STT, translation and TTS together adds their memory requirements. The estimate is approximate.

Quantized models can use less memory. An F16/BF16 GGUF package describes weight storage, not a guarantee of native FP16/BF16 arithmetic on the selected Vulkan device.

## Other runtimes and offline use

- **HIP/ROCm:** audio.cpp accepts a custom server build, but there is no managed HIP download or normal UI option.
- **DirectML:** not an audio.cpp backend and not a supported option in the current UI.
- Download the backend, runtimes and chosen models before using them offline. Some model options require manually supplied local weights.
- Online plugins need a connection and may send audio or text to their provider. Model licenses are separate from the application's source license.

[Installation](../README.md#installation) · [TTS models](documentations/integrated-tts.md) · [Audio setup](audio-config.md) · [Runtime implementation](https://github.com/Sharrnah/whispering/blob/main/Models/audio_cpp_runtime.py)
