fyne bundle --prefix Resource --package Resources --output Resources/bundleImage.go app-icon.png
echo fyne bundle --output bundle.go --append GoNoto.ttf
fyne bundle --prefix Resource --package Resources --output Resources/bundleAudio.go test.wav
