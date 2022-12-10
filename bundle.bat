fyne bundle --output bundle.go app-icon.png
echo fyne bundle --output bundle.go --append GoNoto.ttf
fyne bundle --package Pages --output Pages/bundleAudio.go test.wav
