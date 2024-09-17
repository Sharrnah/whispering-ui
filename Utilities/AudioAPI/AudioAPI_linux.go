//go:build linux

package AudioAPI

import (
	"github.com/gen2brain/malgo"
)

var AudioBackends = []AudioBackend{
	{malgo.BackendAlsa, "alsa", "ALSA"},
	{malgo.BackendPulseaudio, "pulseaudio", "PulseAudio"},
}
