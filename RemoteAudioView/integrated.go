package RemoteAudioView

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"whispering-tiger-ui/LocalPlugins"
	"whispering-tiger-ui/RemoteAudio"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities/AudioAPI"
)

var integrated struct {
	sync.Mutex
	client *RemoteAudio.Client
	cancel context.CancelFunc
	done   chan struct{}
	emit   func([]byte)
	status binding.String
	ready  bool
}
var integratedLifecycle sync.Mutex
var remoteExports sync.Map

// Export destinations originate on this PC and never travel to the AI host.
func HandleIntegratedReceive(raw []byte) bool {
	if !Settings.Config.Run_backend {
		var event map[string]interface{}
		if json.Unmarshal(raw, &event) == nil {
			if event["type"] == "transcript" {
				LocalPlugins.Transcript(event, true)
			}
			if event["type"] == "processing_data" {
				if text, ok := event["data"].(string); ok {
					LocalPlugins.Transcript(map[string]interface{}{"text": text}, false)
				}
			}
		}
	}
	var value struct {
		Type string `json:"type"`
		ID   string `json:"request_id"`
		WAV  string `json:"wav_data"`
	}
	if json.Unmarshal(raw, &value) != nil || value.Type != "tts_save" || value.ID == "" {
		return false
	}
	path, ok := remoteExports.LoadAndDelete(value.ID)
	if !ok {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(value.WAV)
	if err == nil {
		err = os.WriteFile(path.(string), data, 0644)
	}
	if err != nil {
		setRemoteStatus(err.Error())
		emitRemote(map[string]interface{}{"type": "error", "data": err.Error()})
	}
	return true
}

func SetReceiver(emit func([]byte)) { integrated.Lock(); integrated.emit = emit; integrated.Unlock() }
func emitRemote(value interface{}) {
	raw, _ := json.Marshal(value)
	integrated.Lock()
	emit := integrated.emit
	integrated.Unlock()
	if emit != nil {
		emit(raw)
	}
}
func setRemoteStatus(value string) {
	integrated.Lock()
	status := integrated.status
	integrated.Unlock()
	if status != nil {
		_ = status.Set(value)
	}
}

func StopIntegrated() {
	integratedLifecycle.Lock()
	defer integratedLifecycle.Unlock()
	stopIntegrated()
}

func stopIntegrated() {
	integrated.Lock()
	cancel, client, done := integrated.cancel, integrated.client, integrated.done
	integrated.cancel = nil
	integrated.Unlock()
	if cancel != nil {
		cancel()
	}
	if client != nil {
		client.Close()
	}
	if done != nil {
		<-done
	}
}

// Called on the UI thread, where the current profile and device selections live.
func StartIntegrated() {
	LocalPlugins.AutoStart()
	startIntegrated(nil)
}

func startIntegrated(completed func(error)) {
	var completion sync.Once
	report := func(err error) {
		completion.Do(func() {
			if completed != nil {
				completed(err)
			}
		})
	}
	conf := Settings.Config
	prefs := fyne.CurrentApp().Preferences()
	token := prefs.String("remote.audio.token." + conf.Websocket_ip)
	port := prefs.IntWithFallback("remote.audio.port", 5001)
	if token == "" {
		setRemoteStatus(lang.L("Enter the AI PC pairing key to connect audio."))
		report(fmt.Errorf("%s", lang.L("Enter the AI PC pairing key to connect audio.")))
		return
	}
	address := "ws://" + net.JoinHostPort(conf.Websocket_ip, strconv.Itoa(port))
	go func() {
		integratedLifecycle.Lock()
		stopIntegrated()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		integrated.Lock()
		integrated.cancel = cancel
		integrated.done = done
		integrated.ready = false
		integrated.Unlock()
		integratedLifecycle.Unlock()
		defer close(done)
		for ctx.Err() == nil {
			setRemoteStatus(lang.L("Connecting"))
			devices, err := RemoteAudio.OpenDevices(AudioAPI.GetAudioBackendByName(conf.Audio_api).Backend)
			if err == nil {
				devices.Process = conf.Audio_input_process
				devices.ProcessID = uint32(conf.Audio_input_process_id)
				err = devices.SelectNames(conf.Audio_input_device, conf.Audio_output_device)
				if err != nil {
					devices.Close()
				}
			}
			if err == nil {
				client := &RemoteAudio.Client{Audio: devices, FollowProfile: true}
				client.OnCapture = func() {
					integrated.Lock()
					integrated.ready = true
					integrated.Unlock()
					setRemoteStatus(lang.L("Connected"))
					report(nil)
				}
				client.OnEvent = func(event RemoteAudio.Event) {
					switch event.Type {
					case "ready":
						select {
						case SendMessageChannel.SendMessageChannel <- SendMessageChannel.SendMessageStruct{Type: "remote_audio_attach", Value: map[string]string{"token": event.AttachToken}}:
						case <-ctx.Done():
						}
					case "transcript":
						LocalPlugins.Transcript(event.Result, event.Final)
						if event.Final {
							emitRemote(event.Result)
						} else {
							emitRemote(map[string]interface{}{"type": "processing_data", "data": event.Result["text"]})
						}
					case "error":
						setRemoteStatus(event.Message)
						emitRemote(map[string]interface{}{"type": "error", "data": event.Message})
					}
				}
				integrated.Lock()
				integrated.client = client
				integrated.Unlock()
				err = client.Run(ctx, address, token, false)
				integrated.Lock()
				integrated.client = nil
				integrated.ready = false
				integrated.Unlock()
			}
			if ctx.Err() != nil {
				break
			}
			report(err)
			setRemoteStatus(fmt.Sprint(err))
			select {
			case <-ctx.Done():
			case <-time.After(3 * time.Second):
			}
		}
		setRemoteStatus(lang.L("Disconnected"))
	}()
}

// Integrated uses the profile's control address and the application's existing
// audio/language/model controls; only pairing is additional.
func Integrated() fyne.CanvasObject {
	prefs := fyne.CurrentApp().Preferences()
	token := widget.NewPasswordEntry()
	token.SetText(prefs.String("remote.audio.token." + Settings.Config.Websocket_ip))
	port := widget.NewEntry()
	port.SetText(strconv.Itoa(prefs.IntWithFallback("remote.audio.port", 5001)))
	status := binding.NewString()
	_ = status.Set(lang.L("Disconnected"))
	integrated.Lock()
	integrated.status = status
	integrated.Unlock()
	connect := widget.NewButton(lang.L("Connect"), func() {
		n, err := strconv.Atoi(port.Text)
		if err != nil || n < 1 || n > 65535 {
			setRemoteStatus(lang.L("Invalid port"))
			return
		}
		prefs.SetString("remote.audio.token."+Settings.Config.Websocket_ip, token.Text)
		prefs.SetInt("remote.audio.port", n)
		StartIntegrated()
	})
	description := widget.NewLabel(lang.L("Remote audio uses this profile's IP address. Choose local audio devices in Application Options and languages in the usual tabs."))
	description.Wrapping = fyne.TextWrapWord
	return container.NewVBox(description, widget.NewForm(widget.NewFormItem(lang.L("Pairing key"), token), widget.NewFormItem(lang.L("Port"), port)), container.NewHBox(connect, widget.NewButton(lang.L("Disconnect"), func() { go StopIntegrated() })), widget.NewLabelWithData(status))
}

// Local device identifiers must never be applied to the AI PC.
func HandleIntegratedMessage(message *SendMessageChannel.SendMessageStruct) bool {
	if Settings.Config.Run_backend {
		return false
	}
	if message.Type == "tts_req" || message.Type == "tts_req_last" {
		integrated.Lock()
		managed, ready := integrated.done != nil, integrated.ready
		integrated.Unlock()
		if managed && !ready {
			go emitRemote(map[string]interface{}{"type": "error", "data": lang.L("Remote audio") + ": " + lang.L("Disconnected")})
			return true
		}
		raw, _ := json.Marshal(message.Value)
		var value map[string]interface{}
		if json.Unmarshal(raw, &value) == nil {
			if path, ok := value["path"].(string); ok && path != "" {
				id := fmt.Sprintf("remote-export-%d", time.Now().UnixNano())
				remoteExports.Store(id, path)
				time.AfterFunc(10*time.Minute, func() { remoteExports.Delete(id) })
				value["request_id"] = id
				delete(value, "path")
				message.Value = value
			}
		}
	}
	if message.Type == "audio_stop" {
		integrated.Lock()
		client := integrated.client
		integrated.Unlock()
		if client != nil {
			_ = client.StopAudio()
			return true
		}
	}
	if message.Type != "audio_input_switch" && message.Type != "audio_output_switch" {
		return false
	}
	raw, _ := json.Marshal(message.Value)
	var values struct {
		RequestID string `json:"request_id"`
		API       string `json:"audio_api"`
		Input     string `json:"audio_input_device"`
		Process   string `json:"audio_input_process"`
		PID       int    `json:"audio_input_process_id"`
		Output    string `json:"audio_output_device"`
	}
	if err := json.Unmarshal(raw, &values); err != nil {
		return true
	}
	fyne.Do(func() {
		previous := Settings.Config
		Settings.Config.Audio_api = values.API
		if message.Type == "audio_input_switch" {
			Settings.Config.Audio_input_device = values.Input
			Settings.Config.Audio_input_process = values.Process
			Settings.Config.Audio_input_process_id = values.PID
		} else {
			Settings.Config.Audio_output_device = values.Output
		}
		startIntegrated(func(err error) {
			errorText := ""
			if err != nil {
				errorText = err.Error()
				fyne.Do(func() {
					Settings.Config.Audio_api = previous.Audio_api
					Settings.Config.Audio_input_device = previous.Audio_input_device
					Settings.Config.Audio_input_process = previous.Audio_input_process
					Settings.Config.Audio_input_process_id = previous.Audio_input_process_id
					Settings.Config.Audio_output_device = previous.Audio_output_device
					StartIntegrated()
				})
			}
			go emitRemote(map[string]interface{}{"type": message.Type + "_result", "data": map[string]interface{}{"request_id": values.RequestID, "success": err == nil, "error": errorText}})
		})
	})
	return true
}
