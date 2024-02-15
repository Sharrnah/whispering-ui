fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go app-icon.png
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append image-recognition-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append speech-to-text-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append text-to-speech-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append translate-icon.svg
fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go --append plugins-icon.svg

fyne bundle --prefix Resource --package Resources --output Resources/bundleFont.go GoNotoKurrent-Regular.ttf
fyne bundle --prefix Resource --package Resources --output Resources/bundleAudio.go test.wav
