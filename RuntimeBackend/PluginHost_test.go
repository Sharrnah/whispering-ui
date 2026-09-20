package RuntimeBackend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestPluginHostChild(t *testing.T) {
	if os.Getenv("WT_PLUGIN_CHILD") != "1" {
		return
	}
	fmt.Println(`{"type":"state","available":["Fixture"]}`)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	os.Exit(0)
}

func TestPluginHostPrivateIPCAndGracefulEOF(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestPluginHostChild$")
	cmd.Env = append(os.Environ(), "WT_PLUGIN_CHILD=1")
	replies := make(chan map[string]interface{}, 4)
	exited := make(chan error, 1)
	host, err := startPluginHostCommand(cmd, func(raw []byte) {
		var value map[string]interface{}
		if json.Unmarshal(raw, &value) == nil {
			replies <- value
		}
	}, func(err error) { exited <- err })
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()
	select {
	case <-replies:
	case <-time.After(5 * time.Second):
		t.Fatal("no ready response")
	}
	if err = host.Send(map[string]interface{}{"type": "transcript", "final": true, "result": map[string]string{"text": "Hello", "txt_translation": "Hallo"}}); err != nil {
		t.Fatal(err)
	}
	select {
	case reply := <-replies:
		result := reply["result"].(map[string]interface{})
		if result["txt_translation"] != "Hallo" {
			t.Fatal(reply)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("transcript not delivered")
	}
	host.Close()
	select {
	case err = <-exited:
		if err != nil {
			t.Fatalf("unclean shutdown: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("helper did not exit")
	}
	if host.Send(map[string]string{"type": "state"}) == nil {
		t.Fatal("closed helper accepted message")
	}
}
