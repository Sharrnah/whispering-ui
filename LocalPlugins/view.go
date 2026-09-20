// Package LocalPlugins manages integrations on the gaming PC, separately from
// the AI PC's existing plugin settings and lifecycle.
package LocalPlugins

import (
	"encoding/json"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/lang"
	"sync"
	"whispering-tiger-ui/RuntimeBackend"
)

var lifecycle sync.Mutex
var state struct {
	sync.Mutex
	host   *RuntimeBackend.PluginHost
	render func(Snapshot)
	status func(string)
	wanted bool
}

type Snapshot struct {
	Type      string                            `json:"type"`
	Message   string                            `json:"message"`
	Available []string                          `json:"available"`
	Enabled   []string                          `json:"enabled"`
	Settings  map[string]map[string]interface{} `json:"settings"`
}

func status(value string) {
	state.Lock()
	callback := state.status
	state.Unlock()
	if callback != nil {
		fyne.Do(func() { callback(value) })
	}
}
func Start() {
	state.Lock()
	state.wanted = true
	state.Unlock()
	go func() {
		lifecycle.Lock()
		defer lifecycle.Unlock()
		state.Lock()
		current := state.host
		wanted := state.wanted
		state.Unlock()
		if current != nil || !wanted {
			return
		}
		status(lang.L("Starting local plugins"))
		var launched *RuntimeBackend.PluginHost
		host, err := RuntimeBackend.StartPluginHost("LocalSettings/plugins.yaml", func(raw []byte) {
			var value Snapshot
			if err := json.Unmarshal(raw, &value); err != nil {
				status(err.Error())
				return
			}
			if value.Type == "error" {
				status(value.Message)
				return
			}
			if value.Type != "state" {
				return
			}
			state.Lock()
			render := state.render
			state.Unlock()
			if render != nil {
				fyne.Do(func() {
					state.Lock()
					wanted := state.wanted
					state.Unlock()
					if wanted {
						render(value)
					}
				})
			}
			status(lang.L("Local plugin host is running"))
		}, func(err error) {
			// Lifecycle lock prevents a late exit from clearing a replacement.
			go func() {
				lifecycle.Lock()
				defer lifecycle.Unlock()
				state.Lock()
				current := state.host == launched
				if current {
					state.host = nil
				}
				state.Unlock()
				if current {
					if err != nil {
						status(err.Error())
					} else {
						status(lang.L("Local plugins are stopped"))
					}
				}
			}()
		})
		if err != nil {
			status(err.Error())
			return
		}
		launched = host
		state.Lock()
		state.host = host
		state.Unlock()
	}()
}
func AutoStart() {
	if fyne.CurrentApp().Preferences().Bool("local.plugins.enabled") {
		Start()
	}
}
func Close() {
	state.Lock()
	state.wanted = false
	state.Unlock()
	lifecycle.Lock()
	defer lifecycle.Unlock()
	state.Lock()
	host := state.host
	state.host = nil
	state.Unlock()
	if host != nil {
		host.Close()
	}
	status(lang.L("Local plugins are stopped"))
}
func send(value interface{}) {
	state.Lock()
	host := state.host
	state.Unlock()
	if host == nil {
		return
	}
	if err := host.Send(value); err != nil {
		status(err.Error())
	}
}
func Transcript(result map[string]interface{}, final bool) {
	send(map[string]interface{}{"type": "transcript", "result": result, "final": final})
}

func Observe(render func(Snapshot), statusCallback func(string)) {
	state.Lock()
	state.render = render
	state.status = statusCallback
	state.Unlock()
	send(map[string]interface{}{"type": "state"})
}
func Send(value interface{}) { send(value) }

func Restart() { go func() { Close(); Start() }() }
