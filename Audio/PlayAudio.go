package Audio

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/ebitengine/oto/v3"
	"github.com/gen2brain/malgo"
	"io"
	"os"
	"time"
	"whispering-tiger-ui/Logging"
)

type TtsResultRaw struct {
	WavData string `json:"wav_data"` // base64 encoded binary data
}

func (res *TtsResultRaw) PlayWAVFromBase64() error {
	// Decode the base64-encoded string into a byte slice
	decodedBytes, err := base64.StdEncoding.DecodeString(res.WavData)
	if err != nil {
		Logging.CaptureException(err)
		return err
	}

	bytesReader := bytes.NewReader(decodedBytes)

	//	//byteWriter := bufio.Writer{}
	//	out, _ := os.Create("test.wav")
	//	wavDecoder := wav.NewDecoder(bytesReader)
	//	fmt.Printf("is this file valid: %t", wavDecoder.IsValidFile())
	//	wavBuf, err := wavDecoder.FullPCMBuffer()
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	wavEncoder := wav.NewEncoder(out,
	//		wavBuf.Format.SampleRate,
	//		int(wavDecoder.BitDepth),
	//		wavBuf.Format.NumChannels,
	//		int(wavDecoder.WavAudioFormat))
	//
	//	wavEncoder.Write(wavBuf)
	//
	//	out.Close()

	// initialize malgo
	var backends = []malgo.Backend{malgo.BackendNull}

	ctx, err := malgo.InitContext(backends, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})
	if err != nil {
		fmt.Println(err)
		Logging.CaptureException(err)
		Logging.Flush(Logging.FlushTimeoutDefault)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	infos, err := ctx.Devices(malgo.Playback)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.DeviceID = infos[19].ID.Pointer()
	//deviceConfig.DeviceType = malgo.DeviceType(22)
	//deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 44800
	deviceConfig.Alsa.NoMMap = 1

	// This is the function that's used for sending more data to the device for playback.
	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		_, err = io.ReadFull(bytesReader, pOutputSample)
		//pOutputSample, err = io.ReadAll(bytesReader)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return nil
}

func (res *TtsResultRaw) PlayWAVFromBase64Oto() error {
	//var channels, sampleRate uint32

	// Decode the base64-encoded string into a byte slice
	decodedBytes, err := base64.StdEncoding.DecodeString(res.WavData)
	if err != nil {
		return err
	}

	bytesReader := bytes.NewReader(decodedBytes)

	// Create a new audio context
	op := &oto.NewContextOptions{}
	op.SampleRate = 44100
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE
	context, readyChan, err := oto.NewContext(op)
	//context, readyChan, err := oto.NewContext(44800, 2, 2)
	if err != nil {
		return err
	}
	<-readyChan

	// Create a new player for the WAV data
	player := context.NewPlayer(bytesReader)

	// Play starts playing the sound and returns without waiting for it (Play() is async).
	player.Play()

	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	err = player.Close()
	if err != nil {
		return err
	}

	return nil
}
