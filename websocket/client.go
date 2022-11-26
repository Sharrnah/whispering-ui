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
)

type MessageStruct struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`

	// only in case of whisper message
	Text                 string `json:"text,omitempty"`
	Language             string `json:"language,omitempty"` // speaker language
	TxtTranslation       string `json:"txt_translation,omitempty"`
	TxtTranslationSource string `json:"txt_translation_source,omitempty"`
	TxtTranslationTarget string `json:"txt_translation_target,omitempty"`

	// only in case of text translate message
	TranslateResult string `json:"translate_result,omitempty"`
	OriginalText    string `json:"original_text,omitempty"`
	TxtFromLang     string `json:"txt_from_lang,omitempty"`

	// only in case of FLAN message
	FlanAnswer string `json:"flan_answer,omitempty"`
}

var addr = flag.String("addr", "127.0.0.1:5000", "http service address")

func messageLoader(c interface{}, message []byte) interface{} {
	err := json.Unmarshal(message, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func (c *MessageStruct) GetMessage(messageData []byte) *MessageStruct {
	return messageLoader(c, messageData).(*MessageStruct)
}

func Start() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
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
			// log.Printf("recv: %s", message)

			var msg MessageStruct
			msg.GetMessage(message)
			msg.HandleMessage()
			log.Printf("recv: %s", msg)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		//case t := <-ticker.C:
		//	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		//	if err != nil {
		//		log.Println("write:", err)
		//		return
		//	}
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
