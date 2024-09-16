//go:build linux

package AudioAPI

import (
	"github.com/gen2brain/malgo"
)

var AudioBackends = []AudioBackend{
	{malgo.BackendPulseaudio, "pulseaudio", "PulseAudio"},
	{malgo.BackendAlsa, "alsa", "ALSA"},
}
