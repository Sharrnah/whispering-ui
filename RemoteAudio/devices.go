//go:build cgo

package RemoteAudio

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"whispering-tiger-ui/Utilities"

	"github.com/gen2brain/malgo"
)

type Device struct {
	Name string
	ID   malgo.DeviceID
}

// Devices uses miniaudio's native backend (WASAPI on Windows; Linux chooses
// PulseAudio/ALSA). No Windows endpoint IDs cross the wire.
type Devices struct {
	Process           string
	ProcessID         uint32
	stopProcess       func()
	ctx               *malgo.AllocatedContext
	input, output     *malgo.Device
	InputID, OutputID *malgo.DeviceID
	Loopback          bool
	mu                sync.Mutex
	pcm               []byte
	rate              uint32
	channels          uint16
	playing           bool
	ended             bool
}

func OpenDevices(backends ...malgo.Backend) (*Devices, error) {
	ctx, err := malgo.InitContext(backends, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	return &Devices{ctx: ctx}, nil
}

func (d *Devices) List(output bool) ([]Device, error) {
	kind := malgo.Capture
	if output {
		kind = malgo.Playback
	}
	infos, err := d.ctx.Devices(kind)
	if err != nil {
		return nil, err
	}
	list := make([]Device, 0, len(infos))
	for _, info := range infos {
		list = append(list, Device{Name: info.Name(), ID: info.ID})
	}
	return list, nil
}

func (d *Devices) Capture(callback func([]byte)) error {
	if d.Process != "" {
		stop, err := Utilities.StartApplicationAudioCapture(d.Process, d.ProcessID, callback)
		d.stopProcess = stop
		return err
	}
	kind := malgo.Capture
	if d.Loopback {
		if runtime.GOOS != "windows" {
			return errors.New("select a Linux monitor input for system audio")
		}
		kind = malgo.Loopback
	}
	config := malgo.DefaultDeviceConfig(kind)
	config.SampleRate = 16000
	config.Capture.Format = malgo.FormatS16
	config.Capture.Channels = 1
	if d.InputID != nil {
		config.Capture.DeviceID = d.InputID.Pointer()
	}
	dev, err := malgo.InitDevice(d.ctx.Context, config, malgo.DeviceCallbacks{Data: func(out, in []byte, frames uint32) { callback(in) }})
	if err != nil {
		return err
	}
	d.input = dev
	return dev.Start()
}

func (d *Devices) Play(pcm []byte, rate uint32, channels uint16) error {
	if d.output != nil && (d.rate != rate || d.channels != channels) {
		// A format change within buffered playback must never drop an utterance.
		d.mu.Lock()
		remaining := len(d.pcm)
		d.mu.Unlock()
		if remaining != 0 {
			return errors.New("audio format changed during buffered playback")
		}
		d.output.Uninit()
		d.output = nil
	}
	if d.output == nil {
		d.rate = rate
		d.channels = channels
		config := malgo.DefaultDeviceConfig(malgo.Playback)
		config.SampleRate = rate
		config.Playback.Format = malgo.FormatS16
		config.Playback.Channels = uint32(channels)
		if d.OutputID != nil {
			config.Playback.DeviceID = d.OutputID.Pointer()
		}
		dev, err := malgo.InitDevice(d.ctx.Context, config, malgo.DeviceCallbacks{Data: func(out, in []byte, frames uint32) {
			clear(out)
			d.mu.Lock()
			defer d.mu.Unlock()
			if !d.playing {
				return
			}
			n := copy(out, d.pcm)
			d.pcm = d.pcm[n:]
			if len(d.pcm) == 0 {
				d.playing = false
				d.ended = false
			}
		}})
		if err != nil {
			return err
		}
		d.output = dev
		if err = dev.Start(); err != nil {
			return err
		}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.pcm)+len(pcm) > int(rate)*int(channels)*2*120 {
		return errors.New("playback buffer is full")
	}
	d.pcm = append(d.pcm, pcm...)
	// Rebuffer after an underrun; completion flushes short utterances.
	if len(d.pcm) >= int(rate)*int(channels)*2 || d.ended {
		d.playing = true
	}
	return nil
}

func (d *Devices) End() {
	d.mu.Lock()
	d.ended = len(d.pcm) > 0
	d.playing = len(d.pcm) > 0
	d.mu.Unlock()
}
func (d *Devices) Stop() { d.mu.Lock(); d.pcm = nil; d.playing = false; d.ended = false; d.mu.Unlock() }
func (d *Devices) Close() {
	if d.stopProcess != nil {
		d.stopProcess()
		d.stopProcess = nil
	}
	if d.input != nil {
		d.input.Uninit()
		d.input = nil
	}
	if d.output != nil {
		d.output.Uninit()
		d.output = nil
	}
	d.Stop()
	if d.ctx != nil {
		_ = d.ctx.Uninit()
		d.ctx.Free()
		d.ctx = nil
	}
}

// SelectNames resolves only this machine's native devices.
func (d *Devices) SelectNames(input, output string) error {
	d.Loopback = strings.HasSuffix(input, " [Loopback]")
	input = strings.TrimSuffix(input, " [Loopback]")
	for _, pair := range []struct {
		name   string
		output bool
		target **malgo.DeviceID
	}{{input, d.Loopback, &d.InputID}, {output, true, &d.OutputID}} {
		if pair.name == "" || pair.name == "Default" || (!pair.output && d.Process != "") {
			continue
		}
		list, err := d.List(pair.output)
		if err != nil {
			return err
		}
		found := false
		for _, dev := range list {
			if dev.Name == pair.name {
				id := dev.ID
				*pair.target = &id
				found = true
				break
			}
		}
		if !found {
			return errors.New("audio device unavailable: " + pair.name)
		}
	}
	return nil
}
