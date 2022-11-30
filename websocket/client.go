package websocket

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"whispering-tiger-ui/Fields"
)

// Websocket Client

var addr = flag.String("addr", "127.0.0.1:5000", "http service address")

func messageLoader(c interface{}, message []byte) interface{} {
	err := json.Unmarshal(message, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func Start() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	// retry
	for err != nil {
		log.Println("dial:", err)
		time.Sleep(500)
		log.Println("retrying... ")
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	}

	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var msg MessageStruct
			msg.GetMessage(message)
			msg.HandleReceiveMessage()
			//log.Printf("recv: %s", msg)
		}
	}()

	for {
		select {
		case <-done:
			return
		case message := <-Fields.SendMessageChannel:
			HandleSendMessage(&message)
			if message.Value != nil {
				sendMessage, _ := json.Marshal(message)
				err := c.WriteMessage(websocket.TextMessage, sendMessage)
				if err != nil {
					log.Println("write:", err)
					//return
				}
			}

		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
