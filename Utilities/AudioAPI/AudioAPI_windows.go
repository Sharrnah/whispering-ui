//go:build windows

package AudioAPI

import (
	"github.com/gen2brain/malgo"
)

var AudioBackends = []AudioBackend{
	//{malgo.BackendNull, "null", "Null"},
	{malgo.BackendWasapi, "wasapi", "WASAPI"},
	{malgo.BackendWinmm, "mme", "MME"},
	//{malgo.BackendDsound, "directsound", "DirectSound"},
}
