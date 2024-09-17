package AudioAPI

import (
	"github.com/gen2brain/malgo"
	"strings"
)

type AudioBackend struct {
	Backend malgo.Backend
	Id      string
	Name    string
}

// var AudioBackends first API in slice is default

func GetAudioBackendByID(id string) AudioBackend {
	for _, backend := range AudioBackends {
		if strings.ToLower(backend.Id) == strings.ToLower(id) {
			return backend
		}
	}
	return AudioBackends[0]
}

func GetAudioBackendByName(name string) AudioBackend {
	for _, backend := range AudioBackends {
		if strings.ToLower(backend.Name) == strings.ToLower(name) {
			return backend
		}
	}
	return AudioBackends[0]
}
