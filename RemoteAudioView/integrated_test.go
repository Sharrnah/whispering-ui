package RemoteAudioView

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
)

func TestRemoteExportKeepsDestinationOnClient(t *testing.T) {
	previous := Settings.Config.Run_backend
	Settings.Config.Run_backend = false
	defer func() { Settings.Config.Run_backend = previous }()
	path := filepath.Join(t.TempDir(), "speech.wav")
	message := SendMessageChannel.SendMessageStruct{Type: "tts_req_last", Value: map[string]interface{}{"path": path, "download": true, "to_device": false}}
	if HandleIntegratedMessage(&message) {
		t.Fatal("request should use the control connection")
	}
	value := message.Value.(map[string]interface{})
	if _, exists := value["path"]; exists {
		t.Fatal("client path leaked to host")
	}
	pcm := []byte("RIFF-test-audio")
	response, _ := json.Marshal(map[string]interface{}{"type": "tts_save", "request_id": value["request_id"], "wav_data": base64.StdEncoding.EncodeToString(pcm)})
	if !HandleIntegratedReceive(response) {
		t.Fatal("export was not routed")
	}
	actual, err := os.ReadFile(path)
	if err != nil || string(actual) != string(pcm) {
		t.Fatalf("export mismatch: %q %v", actual, err)
	}
	if HandleIntegratedReceive(response) {
		t.Fatal("completed export reused")
	}
}
