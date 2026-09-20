// Package RemoteAudio implements the platform-independent Whispering Tiger LAN protocol.
package RemoteAudio

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Audio interface {
	Capture(func([]byte)) error
	Play([]byte, uint32, uint16) error
	End()
	Stop()
	Close()
}

type Event struct {
	AttachToken string                 `json:"attach_token"`
	Type        string                 `json:"type"`
	Version     int                    `json:"version"`
	Final       bool                   `json:"final"`
	Message     string                 `json:"message"`
	Result      map[string]interface{} `json:"result"`
}

type Client struct {
	OnCapture     func()
	FollowProfile bool
	Audio         Audio
	OnEvent       func(Event)
	mu            sync.Mutex
	conn          *websocket.Conn
	cancel        context.CancelFunc
	muted         atomic.Bool
}

func (c *Client) StopAudio() error {
	c.muted.Store(true)
	c.Audio.Stop()
	return c.Send(map[string]string{"type": "stop"})
}

func DecodeAudio(packet []byte) ([]byte, uint32, uint16, error) {
	if len(packet) < 10 || len(packet) > 65536 {
		return nil, 0, 0, errors.New("invalid audio packet size")
	}
	rate := binary.LittleEndian.Uint32(packet)
	channels := binary.LittleEndian.Uint16(packet[4:])
	if rate < 8000 || rate > 192000 || channels < 1 || channels > 2 || binary.LittleEndian.Uint16(packet[6:]) != 1 || (len(packet)-8)%int(channels*2) != 0 {
		return nil, 0, 0, errors.New("unsupported audio format")
	}
	return packet[8:], rate, channels, nil
}

func (c *Client) Send(value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return errors.New("not connected")
	}
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.conn.WriteJSON(value)
}

func (c *Client) Close() {
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

// Run owns device and socket lifetimes. Reconnecting creates fresh buffers.
func (c *Client) Run(parent context.Context, address, token string, speak bool) error {
	if c.Audio == nil {
		return errors.New("no audio implementation configured")
	}
	defer c.Audio.Close()
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	c.mu.Lock()
	c.cancel = cancel
	c.mu.Unlock()
	u, err := url.Parse(address)
	if err != nil || (u.Scheme != "ws" && u.Scheme != "wss") || u.Host == "" || u.User != nil {
		return errors.New("expected ws://host:5001 or wss://host:port")
	}
	dialer := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}
	conn.SetReadLimit(65536)
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	defer func() {
		conn.Close()
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
	}()
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()
	if err = c.Send(map[string]interface{}{"version": 1, "token": token, "speak": speak, "follow_profile": c.FollowProfile, "name": "Gaming PC"}); err != nil {
		return err
	}
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	var ready Event
	if json.Unmarshal(raw, &ready) != nil || ready.Type != "ready" || ready.Version != 1 {
		return errors.New("incompatible AI host")
	}
	conn.SetReadDeadline(time.Now().Add(45 * time.Second))
	conn.SetPingHandler(func(data string) error {
		conn.SetReadDeadline(time.Now().Add(45 * time.Second))
		return conn.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(3*time.Second))
	})
	if c.OnEvent != nil {
		c.OnEvent(ready)
	}
	packets := make(chan []byte, 32)
	// Capture callbacks never block on networking. Overflow disconnects instead of
	// retaining delayed speech or silently joining discontinuous audio.
	var pending []byte
	var captureMu sync.Mutex
	err = c.Audio.Capture(func(pcm []byte) {
		captureMu.Lock()
		defer captureMu.Unlock()
		pending = append(pending, pcm...)
		for len(pending) >= 1024 {
			packet := append([]byte(nil), pending[:1024]...)
			pending = pending[1024:]
			select {
			case packets <- packet:
			default:
				cancel()
				return
			}
		}
	})
	if err != nil {
		return err
	}
	writerDone := make(chan struct{})
	if c.OnCapture != nil {
		c.OnCapture()
	}
	defer func() { cancel(); <-writerDone }()
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case pcm := <-packets:
				c.mu.Lock()
				conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
				err := conn.WriteMessage(websocket.BinaryMessage, pcm)
				c.mu.Unlock()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
	for {
		kind, raw, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if kind == websocket.BinaryMessage {
			if c.muted.Load() {
				continue
			}
			pcm, rate, channels, err := DecodeAudio(raw)
			if err != nil {
				return err
			}
			if err = c.Audio.Play(pcm, rate, channels); err != nil {
				return fmt.Errorf("playback: %w", err)
			}
		} else {
			var event Event
			if err = json.Unmarshal(raw, &event); err != nil {
				return err
			}
			if event.Type == "audio_end" {
				c.Audio.End()
			}
			if event.Type == "audio_stopped" {
				c.muted.Store(false)
			}
			if c.OnEvent != nil {
				c.OnEvent(event)
			}
		}
	}
}
