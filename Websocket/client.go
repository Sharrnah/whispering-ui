package Websocket

import (
	"encoding/json"
	"flag"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Settings"
)

func messageLoader(c interface{}, message []byte) interface{} {
	err := json.Unmarshal(message, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

type Client struct {
	Addr            string
	Conn            *websocket.Conn
	sendMessageChan chan Fields.SendMessageStruct
	InterruptChan   chan os.Signal
}

func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
		Conn: nil,
		//SendMessageChan: make(chan Fields.SendMessageStruct),
		sendMessageChan: Fields.SendMessageChannel,
		InterruptChan:   make(chan os.Signal, 1),
	}
}

func (c *Client) Close() {
	c.InterruptChan <- os.Interrupt
}

// Websocket Client

func (c *Client) Start() {
	previouslyConnected := false

	statusBar := widget.NewProgressBarInfinite()
	connectingStateContainer := container.NewVBox()
	connectingStateDialog := dialog.NewCustom(
		"",
		"Hide",
		container.NewBorder(statusBar, nil, nil, nil, connectingStateContainer),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)

	go processingStopTimer()
	go realtimeLabelHideTimer()

	flag.Parse()
	log.SetFlags(0)

	//interrupt := make(chan os.Signal, 1)
	signal.Notify(c.InterruptChan, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: c.Addr, Path: "/"}
	log.Printf("connecting to %s", u.String())
	connectingStateContainer.Add(widget.NewLabel("Connecting to " + u.String()))
	connectingStateDialog.Show()

	var err error = nil
	c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	// retry
	for err != nil {
		log.Println("dial:", err)
		time.Sleep(500)
		log.Println("retrying... ")
		c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	}

	connectingStateDialog.Hide()
	previouslyConnected = true

	defer c.Conn.Close()

	done := make(chan struct{})

	go func() {
		// send remote settings request if running remote backend
		if !Settings.Config.Run_backend {
			sendMessage := Fields.SendMessageStruct{
				Type: "setting_update_req",
			}
			sendMessage.SendMessage()
		}

		defer close(done)
		for {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				// retry
				for err != nil {
					time.Sleep(500)
					log.Println("retrying... ")
					if previouslyConnected {
						connectingStateDialog.Show()
						previouslyConnected = false
					}
					c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
					connectingStateDialog.Hide()
				}
				continue
				//return
			}

			previouslyConnected = true

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
		case message := <-c.sendMessageChan:
			HandleSendMessage(&message)
			if message.Value != SkipMessage {
				sendMessage, _ := json.Marshal(message)
				err := c.Conn.WriteMessage(websocket.TextMessage, sendMessage)
				if err != nil {
					log.Println("write:", err)
					//return
				}
			}

		case <-c.InterruptChan:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
