fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go app-icon.png
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/image-recognition-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/speech-to-text-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/text-to-speech-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/translate-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/plugins-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/heart.png
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append Resources/icons/swap-horizontal.svg

rem fyne bundle --prefix Resource --package Resources --output Resources/bundleFont.go GoNotoKurrent-Regular.ttf
fyne bundle --prefix Resource --package Resources --output Resources/bundleAudio.go Resources/audio/test.wav
