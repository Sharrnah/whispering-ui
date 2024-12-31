# Advanced Features

## Larger UI Scaling (_VR-Mode_)
For a better experience in VR, you can use a larger UI scaling. To do so, You can set the Environment variable "WT_SCALE" to a multiplicator for the UI scaling. For example, if you set it to 2, the UI will be scaled by 2.

```batch
SET WT_SCALE=1.2
start "" "Whispering Tiger.exe"
```
Save the above code as a batch file (`"VR-Mode.bat"` for example), place it besides the `Whispering Tiger.exe` file and run it to start the application with the larger UI scaling.


## Overwrite UI Language
Normally the language displayed by the UI is determined by the OS language.
If you want to use a different language, create a .bat file like mentioned above for UI Scaling, but set the Environment variable "PREFERRED_LANGUAGE" to the language code.

```batch
SET PREFERRED_LANGUAGE=pl
start "" "Whispering Tiger.exe"
```

Only works for already translated languages.