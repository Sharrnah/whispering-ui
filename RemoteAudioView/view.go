// Package RemoteAudioView is shared by the full UI and the model-free client.
package RemoteAudioView

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
	"whispering-tiger-ui/RemoteAudio"
)

func Client(window fyne.Window) (fyne.CanvasObject, func()) {
	prefs := fyne.CurrentApp().Preferences()
	address := widget.NewEntry()
	address.SetText(prefs.StringWithFallback("remote.audio.address", "ws://192.168.1.2:5001"))
	findHosts := widget.NewButton(lang.L("Find AI PCs"), nil)
	findHosts.OnTapped = func() {
		findHosts.Disable()
		go func() {
			hosts, err := RemoteAudio.Discover(1500 * time.Millisecond)
			fyne.Do(func() {
				findHosts.Enable()
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				if len(hosts) == 0 {
					dialog.ShowInformation(lang.L("Find AI PCs"), lang.L("No AI PCs found. Enable remote processing on the AI PC or enter its address."), window)
					return
				}
				options := make([]string, len(hosts))
				for i, host := range hosts {
					options[i] = host.Name + " (" + host.Address + ")"
				}
				selectHost := widget.NewSelect(options, nil)
				picker := dialog.NewCustom(lang.L("Find AI PCs"), lang.L("Close"), selectHost, window)
				selectHost.OnChanged = func(string) { address.SetText(hosts[selectHost.SelectedIndex()].Address); picker.Hide() }
				picker.Show()
			})
		}()
	}
	token := widget.NewPasswordEntry()
	token.SetText(prefs.String("remote.audio.token"))
	status := widget.NewLabel(lang.L("Disconnected"))
	transcript := widget.NewMultiLineEntry()
	transcript.Wrapping = fyne.TextWrapWord
	transcript.Disable()
	speak := widget.NewCheck(lang.L("Speak returned text"), nil)
	speak.SetChecked(prefs.Bool("remote.audio.speak"))
	loopback := widget.NewCheck(lang.L("Capture system audio"), nil)
	if runtime.GOOS != "windows" {
		loopback.Disable()
	}
	inputs := widget.NewSelect([]string{lang.L("Default")}, nil)
	inputs.SetSelectedIndex(0)
	outputs := widget.NewSelect([]string{lang.L("Default")}, nil)
	outputs.SetSelectedIndex(0)
	var inDevices, outDevices []RemoteAudio.Device
	var mu sync.Mutex
	var client *RemoteAudio.Client
	var cancelConnection context.CancelFunc
	refresh := func() {
		devices, err := RemoteAudio.OpenDevices()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		defer devices.Close()
		inDevices, err = devices.List(loopback.Checked)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		outDevices, err = devices.List(true)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		ins := []string{lang.L("Default")}
		outs := []string{lang.L("Default")}
		for i, d := range inDevices {
			ins = append(ins, fmt.Sprintf("%d: %s", i+1, d.Name))
		}
		for i, d := range outDevices {
			outs = append(outs, fmt.Sprintf("%d: %s", i+1, d.Name))
		}
		inputs.Options = ins
		outputs.Options = outs
		inputs.SetSelectedIndex(0)
		outputs.SetSelectedIndex(0)
	}
	loopback.OnChanged = func(bool) { refresh() }
	refresh()
	closeClient := func() {
		mu.Lock()
		c := client
		cancel := cancelConnection
		mu.Unlock()
		if cancel != nil {
			cancel()
		}
		if c != nil {
			c.Close()
		}
	}
	connect := widget.NewButton(lang.L("Connect"), nil)
	connect.OnTapped = func() {
		mu.Lock()
		busy := cancelConnection != nil
		mu.Unlock()
		if busy {
			closeClient()
			return
		}
		devices, err := RemoteAudio.OpenDevices()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if i := inputs.SelectedIndex() - 1; i >= 0 && i < len(inDevices) {
			id := inDevices[i].ID
			devices.InputID = &id
		}
		if i := outputs.SelectedIndex() - 1; i >= 0 && i < len(outDevices) {
			id := outDevices[i].ID
			devices.OutputID = &id
		}
		devices.Loopback = loopback.Checked
		inputID, outputID, selectedLoopback := devices.InputID, devices.OutputID, devices.Loopback
		connectionContext, cancel := context.WithCancel(context.Background())
		selectedAddress, selectedToken, selectedSpeak := strings.TrimSpace(address.Text), strings.TrimSpace(token.Text), speak.Checked
		prefs.SetString("remote.audio.address", selectedAddress)
		prefs.SetString("remote.audio.token", selectedToken)
		prefs.SetBool("remote.audio.speak", selectedSpeak)
		c := &RemoteAudio.Client{Audio: devices, OnEvent: func(event RemoteAudio.Event) {
			fyne.Do(func() {
				switch event.Type {
				case "ready":
					status.SetText(lang.L("Connected"))
				case "transcript":
					text, _ := event.Result["text"].(string)
					if translated, ok := event.Result["txt_translation"].(string); ok && translated != "" {
						text += "\n" + translated
					}
					if event.Final {
						existing := transcript.Text
						if len(existing) > 16000 {
							existing = ""
						}
						transcript.SetText(existing + text + "\n\n")
					} else {
						status.SetText(text)
					}
				case "error":
					status.SetText(event.Message)
				}
			})
		}}
		mu.Lock()
		client = c
		cancelConnection = cancel
		mu.Unlock()
		connect.SetText(lang.L("Disconnect"))
		status.SetText(lang.L("Connecting"))
		address.Disable()
		findHosts.Disable()
		token.Disable()
		inputs.Disable()
		outputs.Disable()
		loopback.Disable()
		speak.Disable()
		go func() {
			var err error
			for {
				err = c.Run(connectionContext, selectedAddress, selectedToken, selectedSpeak)
				if connectionContext.Err() != nil {
					err = nil
					break
				}
				if websocket.IsCloseError(err, 1008) {
					break
				}
				fyne.Do(func() { status.SetText(lang.L("Reconnecting")) })
				select {
				case <-connectionContext.Done():
					err = nil
				case <-time.After(3 * time.Second):
				}
				if connectionContext.Err() != nil {
					break
				}
				devices, err = RemoteAudio.OpenDevices()
				if err != nil {
					break
				}
				devices.InputID = inputID
				devices.OutputID = outputID
				devices.Loopback = selectedLoopback
				c = &RemoteAudio.Client{Audio: devices, OnEvent: c.OnEvent}
				mu.Lock()
				client = c
				mu.Unlock()
			}
			cancel()
			mu.Lock()
			client = nil
			cancelConnection = nil
			mu.Unlock()
			fyne.Do(func() {
				connect.SetText(lang.L("Connect"))
				status.SetText(lang.L("Disconnected"))
				address.Enable()
				findHosts.Enable()
				token.Enable()
				inputs.Enable()
				outputs.Enable()
				speak.Enable()
				if runtime.GOOS == "windows" {
					loopback.Enable()
				}
				if err != nil {
					status.SetText(err.Error())
				}
			})
		}()
	}
	text := widget.NewEntry()
	text.SetPlaceHolder(lang.L("Text to speak"))
	send := widget.NewButton(lang.L("Speak"), func() {
		mu.Lock()
		c := client
		mu.Unlock()
		if c != nil {
			value := text.Text
			go func() {
				if err := c.Send(map[string]string{"type": "speak", "text": value}); err != nil {
					fyne.Do(func() { dialog.ShowError(err, window) })
				}
			}()
		}
	})
	stop := widget.NewButton(lang.L("Stop audio"), func() {
		mu.Lock()
		c := client
		mu.Unlock()
		if c != nil {
			go func() { _ = c.StopAudio() }()
		}
	})
	form := widget.NewForm(widget.NewFormItem(lang.L("AI PC address"), address), widget.NewFormItem(lang.L("Pairing key"), token), widget.NewFormItem(lang.L("Audio input"), inputs), widget.NewFormItem(lang.L("Audio output"), outputs))
	copyText := widget.NewButton(lang.L("Copy"), func() { window.Clipboard().SetContent(transcript.Text) })
	top := container.NewVBox(widget.NewLabel(lang.L("Remote audio client")), findHosts, form, loopback, speak, container.NewHBox(connect, stop, copyText), status)
	bottom := container.NewBorder(nil, nil, nil, send, text)
	return container.NewBorder(top, bottom, nil, nil, transcript), closeClient
}

// Host controls only connect to the local administration socket. The key is
// never persisted in a shared profile or sent to ordinary remote UI clients.
func Host(window fyne.Window, address string) fyne.CanvasObject {
	status := widget.NewLabel(lang.L("Remote audio is disabled"))
	key := widget.NewEntry()
	key.Disable()
	var enable, disable *widget.Button
	apply := func(enabled bool) {
		enable.Disable()
		disable.Disable()
		go func() {
			response, err := configureHost(address, enabled)
			fyne.Do(func() {
				enable.Enable()
				disable.Enable()
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				if response.Enabled {
					key.SetText(response.Token)
					status.SetText(lang.L("Connect the gaming PC to port 5001"))
				} else {
					key.SetText("")
					status.SetText(lang.L("Remote audio is disabled"))
				}
			})
		}()
	}
	enable = widget.NewButton(lang.L("Allow remote processing"), func() { apply(true) })
	disable = widget.NewButton(lang.L("Disable remote processing"), func() { apply(false) })
	copyKey := widget.NewButton(lang.L("Copy pairing key"), func() { window.Clipboard().SetContent(key.Text) })
	return container.NewVBox(status, container.NewHBox(enable, disable), widget.NewForm(widget.NewFormItem(lang.L("Pairing key"), key)), copyKey)
}

type hostResponse struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	Error   string `json:"error"`
}

func configureHost(address string, enabled bool) (hostResponse, error) {
	var result hostResponse
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/", nil)
	if err != nil {
		return result, err
	}
	defer conn.Close()
	conn.SetReadLimit(4 * 1024 * 1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	if err = conn.WriteJSON(map[string]interface{}{"type": "remote_audio_host", "value": map[string]bool{"enabled": enabled}}); err != nil {
		return result, err
	}
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return result, err
		}
		if json.Unmarshal(data, &result) == nil && result.Type == "remote_audio_host" {
			if result.Error != "" {
				return result, fmt.Errorf("%s", result.Error)
			}
			return result, nil
		}
	}
}
