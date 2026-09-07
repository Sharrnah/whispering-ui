package RemoteAudio

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Invoked by the backend's manual transport interop check.
func TestPythonInterop(t *testing.T) {
	address := os.Getenv("WT_REMOTE_TEST_URL")
	if address == "" {
		t.Skip("requires a local Python remote-audio test server")
	}
	audio := &fakeAudio{end: make(chan struct{})}
	transcript := make(chan Event, 1)
	client := &Client{Audio: audio, OnEvent: func(event Event) {
		if event.Type == "transcript" {
			transcript <- event
		}
	}}
	done := make(chan error, 1)
	go func() { done <- client.Run(context.Background(), address, "interop-test", true) }()
	defer client.Close()
	select {
	case event := <-transcript:
		if event.Result["text"] != "Python to Go" {
			t.Fatal(event)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no transcript")
	}
	select {
	case <-audio.end:
	case <-time.After(5 * time.Second):
		t.Fatal("no audio")
	}
	client.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("cleanup timeout")
	}
	audio.mu.Lock()
	defer audio.mu.Unlock()
	if !bytes.Equal(audio.pcm, []byte{1, 0, 2, 0}) {
		t.Fatal("PCM differs across Python/Go")
	}
}

type fakeAudio struct {
	mu     sync.Mutex
	pcm    []byte
	closed bool
	end    chan struct{}
}

func (f *fakeAudio) Capture(send func([]byte)) error {
	send(make([]byte, 320))
	send(make([]byte, 704))
	return nil
}
func (f *fakeAudio) Play(pcm []byte, rate uint32, channels uint16) error {
	f.mu.Lock()
	f.pcm = append(f.pcm, pcm...)
	f.mu.Unlock()
	return nil
}
func (f *fakeAudio) End()   { close(f.end) }
func (f *fakeAudio) Stop()  {}
func (f *fakeAudio) Close() { f.mu.Lock(); f.closed = true; f.mu.Unlock() }

func TestAudioValidation(t *testing.T) {
	for _, packet := range [][]byte{nil, make([]byte, 8), make([]byte, 11), make([]byte, 65537)} {
		if _, _, _, err := DecodeAudio(packet); err == nil {
			t.Fatal("accepted malformed audio")
		}
	}
	packet := make([]byte, 12)
	binary.LittleEndian.PutUint32(packet, 24000)
	binary.LittleEndian.PutUint16(packet[4:], 1)
	binary.LittleEndian.PutUint16(packet[6:], 1)
	pcm, rate, channels, err := DecodeAudio(packet)
	if err != nil || len(pcm) != 4 || rate != 24000 || channels != 1 {
		t.Fatal("valid audio rejected", err)
	}
}

func TestClientRoundTripAndCancellation(t *testing.T) {
	serverErrors := make(chan string, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var hello map[string]interface{}
		_ = json.Unmarshal(raw, &hello)
		if hello["token"] != "secret" || hello["version"] != float64(1) {
			serverErrors <- "bad handshake"
			return
		}
		_ = conn.WriteJSON(map[string]interface{}{"type": "ready", "version": 1})
		kind, pcm, err := conn.ReadMessage()
		if err != nil || kind != websocket.BinaryMessage || len(pcm) != 1024 {
			serverErrors <- "bad capture framing"
			return
		}
		packet := make([]byte, 12)
		binary.LittleEndian.PutUint32(packet, 24000)
		binary.LittleEndian.PutUint16(packet[4:], 1)
		binary.LittleEndian.PutUint16(packet[6:], 1)
		copy(packet[8:], []byte{1, 0, 2, 0})
		_ = conn.WriteMessage(websocket.BinaryMessage, packet)
		_ = conn.WriteJSON(map[string]string{"type": "audio_end"})
		_, _, _ = conn.ReadMessage()
	}))
	defer server.Close()
	audio := &fakeAudio{end: make(chan struct{})}
	client := &Client{Audio: audio}
	done := make(chan error, 1)
	go func() {
		done <- client.Run(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), "secret", true)
	}()
	select {
	case <-audio.end:
	case msg := <-serverErrors:
		t.Fatal(msg)
	case <-time.After(5 * time.Second):
		t.Fatal("round trip timed out")
	}
	client.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("client failed to close")
	}
	audio.mu.Lock()
	defer audio.mu.Unlock()
	if !audio.closed || !bytes.Equal(audio.pcm, []byte{1, 0, 2, 0}) {
		t.Fatal("audio or cleanup mismatch")
	}
}
