package Websocket

import (
	"bytes"
	"encoding/json"
	"flag"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

const (
	maxMessageSize = 10 / Utilities.MiB
)

type Client struct {
	Addr            string
	Conn            *websocket.Conn
	sendMessageChan chan SendMessageChannel.SendMessageStruct
	InterruptChan   chan os.Signal
}

func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
		Conn: nil,
		//SendMessageChan: make(chan Fields.SendMessageStruct),
		sendMessageChan: SendMessageChannel.SendMessageChannel,
		InterruptChan:   make(chan os.Signal, 1),
	}
}

func (c *Client) Close() {
	c.InterruptChan <- os.Interrupt
}

// Websocket Client

func (c *Client) Start() {
	defer Utilities.PanicLogger()

	previouslyConnected := false

	runBackend := Settings.Config.Run_backend

	statusBar := widget.NewProgressBarInfinite()
	connectingStateContainer := container.NewVBox()
	connectingStateDialog := dialog.NewCustom(
		"",
		lang.L("Hide"),
		container.NewBorder(statusBar, nil, nil, nil, connectingStateContainer),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)

	go processingStopTimer()
	go realtimeLabelHideTimer()
	go ProcessReceiveMessageChannel()

	flag.Parse()
	log.SetFlags(0)

	//interrupt := make(chan os.Signal, 1)
	signal.Notify(c.InterruptChan, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: c.Addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	connectingStateContainer.Add(widget.NewLabel(lang.L("Connecting to Server", map[string]interface{}{"ServerUri": u.String()})))
	connectingStateDialog.Show()

	// create websocket dialer
	dialer := websocket.DefaultDialer
	dialer.EnableCompression = true
	dialer.HandshakeTimeout = 120 * time.Second

	var err error = nil
	c.Conn, _, err = dialer.Dial(u.String(), nil)
	//c.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	// retry
	for err != nil {
		log.Println("dial:", err)
		time.Sleep(500)
		log.Println("retrying... ")
		c.Conn, _, err = dialer.Dial(u.String(), nil)
	}
	time.Sleep(100)

	connectingStateDialog.Hide()
	previouslyConnected = true

	defer c.Conn.Close()
	c.Conn.SetReadLimit(maxMessageSize)

	done := make(chan struct{})

	go func() {
		// send remote settings request if running remote backend
		if runBackend {
			// send info that backend is running locally
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type:  "ui_connected",
				Value: true,
			}
			sendMessage.SendMessage()
		} else {
			sendMessage := SendMessageChannel.SendMessageStruct{
				Type: "setting_update_req",
			}
			sendMessage.SendMessage()
		}

		defer close(done)
		for {
			_, r, err := c.Conn.NextReader()
			if err != nil {
				log.Println("read:", err)
				// retry
				for err != nil {
					log.Println("retrying after disconnect... ")
					if previouslyConnected {
						connectingStateDialog.Show()
						previouslyConnected = false
					}
					c.Conn, _, err = dialer.Dial(u.String(), nil)
					time.Sleep(500 * time.Millisecond) // make sure to multiply by time.Millisecond
					connectingStateDialog.Hide()
				}
				if runBackend {
					log.Println("send ui_connected")
					// send info that backend is running locally
					sendMessage := SendMessageChannel.SendMessageStruct{
						Type:  "ui_connected",
						Value: true,
					}
					sendMessage.SendMessage()
				} else {
					sendMessage := SendMessageChannel.SendMessageStruct{
						Type: "setting_update_req",
					}
					sendMessage.SendMessage()
				}
				continue
			}

			previouslyConnected = true

			// Read the message using io.Reader
			// buf := make([]byte, 1024)
			buf := make([]byte, 4096)
			var buffer []byte // Holds incoming data
			for {
				// Read data into buf as before
				n, err := r.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Println("Error reading message:", err)
					}
					break
				}
				buffer = append(buffer, buf[:n]...)

				// Create a new reader and decoder for the current state of buffer
				reader := bytes.NewReader(buffer)
				decoder := json.NewDecoder(reader)

				for decoder.More() {
					var msgStruct interface{} // Use the specific struct you expect to decode, or interface{} for generic JSON objects
					err := decoder.Decode(&msgStruct)

					if err != nil {
						log.Println("Error decoding JSON:", err)
						// If an error occurs, break out of the loop. You'll need to handle partial messages
						break
					}

					// Successfully decoded a message, now send it to ReceiveMessageChannel
					// Since decoder doesn't provide directly the raw message, you might need to re-marshal if you need the exact JSON
					processedJSON, err := json.Marshal(msgStruct)
					if err != nil {
						log.Println("Error marshaling decoded JSON:", err)
						break
					}
					ReceiveMessageChannel <- processedJSON

					// Update buffer to remove processed message
					// This is a bit tricky since decoder does not tell us directly how much of the input it consumed
					// We'll have to rely on re-marshaling and knowing the structure of our JSON to get the byte length, or
					// we assume the decoder consumed exactly what was needed for the message it decoded.
					newPos := reader.Size() - int64(reader.Len()) // Calculate how much of the buffer we've processed
					buffer = buffer[newPos:]                      // This assumes all data before newPos is processed, which might not always be accurate
				}

				// Handle case where buffer might be empty or contain partial data
				if len(buffer) == 0 || !decoder.More() {
					// Clear the buffer or handle partial data
					buffer = nil // Reset the buffer if we think we've processed all messages
				}
			}
		}
	}()

	go func() {
		defer Utilities.PanicLogger()
		for {
			select {
			//case <-done:
			//	return
			case message := <-c.sendMessageChan:
				HandleSendMessage(&message)
				if message.Value != SkipMessage {
					sendMessage, err := json.Marshal(message)
					if err != nil {
						log.Println("Error marshaling message:", err)
						//return
					} else {
						if c.Conn != nil { // make sure connection is not closed before sending message
							err := c.Conn.WriteMessage(websocket.TextMessage, sendMessage)
							if err != nil {
								log.Println("write:", err)
								//return
							}
						}
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
	}()

	// keep function running until interrupted
	for {
		for {
			select {
			case <-done:
				return
			}
		}
	}
}
