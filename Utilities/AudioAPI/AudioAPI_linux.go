//go:build linux

package AudioAPI

import (
	"github.com/gen2brain/malgo"
)

var AudioBackends = []AudioBackend{
	//{malgo.BackendNull, "null", "Null"},
	{malgo.BackendAlsa, "alsa", "ALSA"},
	{malgo.BackendPulseaudio, "pulseaudio", "PulseAudio"},
	{malgo.BackendJack, "jack", "Jack"},
}
