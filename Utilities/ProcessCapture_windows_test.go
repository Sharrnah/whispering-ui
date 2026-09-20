//go:build windows && cgo

package Utilities

import (
	"encoding/binary"
	"github.com/gen2brain/malgo"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// Opt-in hardware check: emits a quiet synthetic tone, never opens a microphone.
func TestNativeProcessCapture(t *testing.T) {
	if os.Getenv("WT_TEST_PROCESS_AUDIO") != "1" {
		t.Skip("requires a Windows playback endpoint")
	}
	ctx, err := malgo.InitContext([]malgo.Backend{malgo.BackendWasapi}, malgo.ContextConfig{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.Free()
	defer ctx.Uninit()
	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.SampleRate = 16000
	cfg.Playback.Channels = 1
	cfg.Playback.Format = malgo.FormatS16
	phase := 0
	dev, err := malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{Data: func(out, _ []byte, _ uint32) {
		for i := 0; i+1 < len(out); i += 2 {
			sample := int16(math.Sin(float64(phase)*2*math.Pi*440/16000) * 3000)
			binary.LittleEndian.PutUint16(out[i:], uint16(sample))
			phase++
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Uninit()
	exe, _ := os.Executable()
	var audible atomic.Int64
	stop, err := StartApplicationAudioCapture(filepath.Base(exe), uint32(os.Getpid()), func(pcm []byte) {
		for i := 0; i+1 < len(pcm); i += 2 {
			sample := int16(binary.LittleEndian.Uint16(pcm[i:]))
			if sample > 100 || sample < -100 {
				audible.Add(1)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	if err = dev.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1500 * time.Millisecond)
	if audible.Load() < 1000 {
		t.Fatalf("no captured application tone: %d samples", audible.Load())
	}
	t.Logf("captured %d audible samples from the selected process", audible.Load())
}
