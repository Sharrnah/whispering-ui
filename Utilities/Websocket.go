package Utilities

import (
	"github.com/gorilla/websocket"
	"net"
	"net/url"
	"time"
)

func CheckPortInUse(addr string) bool {
	timeout := 1 * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		// If there is an error, we assume the port is not in use
		return false
	}

	// Close the connection if the port is in use
	conn.Close()
	return true
}

func SendQuitMessage(addr string) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer func() {
		time.Sleep(500 * time.Millisecond) // Add a delay before closing the connection
		conn.Close()
	}()

	message := map[string]interface{}{
		"type": "quit",
		"data": true,
	}

	err = conn.WriteJSON(message)
	if err != nil {
		return err
	}

	return nil
}
