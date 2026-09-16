# Desktop builds

The UI now contains its Linux implementation directly. It recognizes the Linux
backend executable, selects `ai_platform_linux_amd64_cu128` from the existing
hosted `latest.yaml`, preserves ZIP executable permissions, and terminates the
Linux backend's child processes on forced shutdown. Windows retains
`audioWhisper.exe`, the `ai_platform` update entry and Windows Job Objects.
Linux device discovery also finds the audio.cpp CPU/Vulkan server bundled by
the backend packager under `toolchain/audio.cpp/`.

## Build locally, entirely in Docker

From the UI repository, with Docker running:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\BuildTools\build-docker.ps1
```

The default builds Linux/CUDA and Windows in isolated source copies. Use
`-Target linux` or `-Target windows` for one platform, and `-Flavor cpu` for the
Linux CPU backend selection. CUDA is the default; compiling the UI or backend
does not require an NVIDIA GPU. The UI build does not package a backend.

Artifacts appear in `Build/docker/<timestamp>/<platform>/`: the executable,
versioned UI ZIP, `SHA256SUMS`, and `build-info.json`. Linux ZIPs include the UI
license; build instructions stay in this repository. Keep generated artifacts out of Git.

Linux tests execute under Xvfb. Every Windows test package and the full Windows
application are cross-compiled with MinGW; Windows tests are **not executed** in
the Linux container. A Windows desktop/audio run remains a separate runtime
check. Native Windows `build.bat` still uses `fyne package --release`, including
the Windows executable icon and version resources. It now runs in a temporary
source copy and copies `Whispering Tiger.exe` back to the repository root.

`FyneApp.toml` supplies the embedded version and build number, icon and Fyne
migration settings. No CLI tool is fetched at `@latest`, and `go.mod`/`go.sum`
are not rewritten. The generated metadata file exists only in a build copy.
The Linux executable's `--version` output is checked after compilation, from a
directory that has no `FyneApp.toml`.

## Matching Windows and Linux versions

Set `Version` and `Build` in the UI repository's `FyneApp.toml` once per release.
For example, `Version = "1.3.11"` and `Build = 1` produce UI version `1.3.11.1`
on both platforms. Use `Build = 2` for the next build of that version.

```powershell
# Windows UI, from this repository:
.\build.bat

# Linux UI only, from this repository:
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\BuildTools\build-docker.ps1 -Target linux -Release
```

The combined backend repository's `builder/build-linux.ps1` reads the same UI
TOML. Its `-Version` argument selects the **backend** version, not the UI version.
Build either platform first; keep the same TOML values for both.

The installed Fyne CLI 1.7.0 increments `Build` **after** successful packaging,
even when `--app-build` is supplied. The executable contains the value from
before that increment. `build.bat` isolates that write in its temporary copy;
the Docker builders already leave the original TOML unchanged. Running the raw
`fyne package --release` command directly in the repository still increments it
and can make the next Linux build one number ahead. Use `build.bat` instead.

Linux uses `go build` with embedded Fyne metadata and release build tags; it does
not run `fyne package`. Both the application icon and metadata icon are compiled
into the executable. `fyne bundle` embeds image resources; it does not add
Windows-style executable icon resources to Linux files. The Linux distribution
remains portable, with no installer or automatic desktop registration. Whether
a file manager displays an icon on the executable itself is separate from the
running window's icon.

## Release builds and Drone

Local builds default to Linux previews with update checks disabled. Add
`-Release` to produce release metadata and enable Linux update checks. This
switch does not upload anything. Finish the manual Linux desktop/CUDA test and
prepare the backend entry in the hosted feed before publishing.

The UI `.drone.yml` uses the same build script with Go 1.26.5. Every build runs
Linux tests and checks Windows compilation. Tag events additionally produce
release builds and upload the versioned Linux ZIP/checksums through the existing
S3/Gitea steps. Tags must match `Version.Build` in `FyneApp.toml` (optional `v`
prefix). Do not push a release tag merely to run a preview. Gitea's base URL is
derived from the repository URL; the existing secret names are retained.

This updates the UI pipeline only. The Python backend's older `.drone.yml`
still needs its separate packaging migration. Neither pipeline automatically
merges a Linux backend entry into the hosted `latest.yaml`. Keep the existing
`app` and Windows `ai_platform` entries; the common `app.version` must match the
published UI's `Version.Build`. Publish the Linux UI on the GitHub release page
too if Gitea is a mirror: the UI's update notification opens GitHub.

## Source files to commit

Commit the Linux UI source/test changes, `build.bat`, `.drone.yml`, `.gitignore`,
`.gitattributes`, and this `BuildTools/` directory. The Linux additions are:

```text
main.go
Pages/Ocr.go
Pages/Profiles.go
ProfileForm/Builder.go
ProfileForm/Schema.go
ProfileForm/Schema_audio_cpp_test.go
RuntimeBackend/Whisper.go
RuntimeBackend/jobobject_stub.go
RuntimeBackend/jobobject_windows.go
RuntimeBackend/process_group_linux_test.go
UpdateUtility/UpdateCheck.go
Updater/Platform.go
Updater/Platform_test.go
Updater/Unzip.go
Updater/Unzip_test.go
Utilities/BackendPath.go
Utilities/BackendPath_test.go
Utilities/Hardwareinfo/NVIDIAMemory.go
Utilities/Hardwareinfo/UnknownGPU_test.go
Utilities/Hardwareinfo/AudioCppDevices.go
Utilities/Hardwareinfo/AudioCppDevices_test.go
```

Existing edits to `FyneApp.toml` and `doc/documentations/integrated-tts.md` were
preserved. Review their changes separately when choosing the release version.
Local profiles, archives and audio prototypes are not required
by these UI builds.

The combined Docker builder in the backend repository reads this UI repository
and uses its `BuildTools/build.py`. Temporary source copies live inside Docker;
no UI patch is applied. Older UI folders under the backend's `.linux-build/`
are historical build snapshots, not source repositories to edit.
Snapshots include tracked files and new, non-ignored Go sources, Resources,
and BuildTools files so current worktree additions can build before commit.

## Validation on 2026-09-08

Both preview and release configurations passed the complete Linux Go tests and
built Linux and Windows executables. Every Windows test package cross-compiled;
no Windows test was executed. The built Linux executables reported `1.3.10.47`
without a nearby TOML file. Preview ZIP integrity, SHA-256 files, Unix executable
permissions, Windows x86-64 PE format and standard Windows DLL imports were
checked. Locale JSON, Go formatting and Drone YAML parsed successfully, and
Drone's own substitution library verified the Gitea URL expression. The
combined backend/UI staging and UI build also completed successfully.

Local preview outputs: `Build/docker/20260908-233237/`. Local release-mode
validation outputs: `Build/docker/20260908-233831/`. These are UI-only artifacts;
no tag, upload or publication was performed. Real Linux desktop/CUDA testing and
a Windows desktop/audio run remain separate from these build checks.

On 2026-09-10, the current repository snapshot (including new non-ignored Go
helpers) passed all Linux tests and built Linux metadata version `1.3.10.50`.
The Windows release executable and every Windows test package cross-compiled
successfully using the same snapshot. Windows tests were not executed.
The audio.cpp bundle discovery follow-up passed those checks again; its Linux
UI ZIP contains only the executable and license, with build information outside.

On 2026-09-14, the Linux release build passed the complete Go suite and reported
`1.3.11.1`. Real Fyne CLI 1.7.0 Windows packaging, cross-compiled in Docker with
MinGW, produced PE file version `1.3.11.1` with icon, version and manifest resources;
its temporary TOML advanced to Build 2 while the repository stayed at Build 1.
The Windows wrapper's source selection, repeated builds, failure handling and
temporary cleanup passed its PowerShell checks in Docker. Windows execution was
not tested. An X11 probe verified the exact bundled icon pixels on a running
window; this does not verify Dolphin's executable-file icon.
