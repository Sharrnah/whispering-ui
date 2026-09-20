package Websocket

import (
	"encoding/json"
	"flag"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Fields"
	"whispering-tiger-ui/Logging"
	"whispering-tiger-ui/RemoteAudioView"
	"whispering-tiger-ui/RuntimeBackend"
	"whispering-tiger-ui/SendMessageChannel"
	"whispering-tiger-ui/Settings"
	"whispering-tiger-ui/Utilities"
)

const (
	maxMessageSize = 64 * Utilities.MiB
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
	go RemoteAudioView.StopIntegrated()
	c.InterruptChan <- os.Interrupt
}

// Websocket Client

func (c *Client) Start() {
	defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
		scope.SetTag("GoRoutine", "Websocket\\client->Start")
	})
	previouslyConnected := false

	runBackend := Settings.Config.Run_backend
	if !runBackend {
		RemoteAudioView.SetReceiver(func(raw []byte) { ReceiveMessageChannel <- raw })
	}

	statusBar := widget.NewProgressBarInfinite()
	connectingStateContainer := container.NewVBox()
	connectingStateDialog := dialog.NewCustomWithoutButtons(
		"",
		container.NewBorder(statusBar, nil, nil, nil, connectingStateContainer),
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
	connectingStateDialog.SetButtons([]fyne.CanvasObject{
		&widget.Button{
			Text: lang.L("Show Log"),
			OnTapped: func() {
				logWindow := fyne.CurrentApp().NewWindow(lang.L("Logs"))
				copyLogButton := widget.NewButtonWithIcon(lang.L("Copy Log"), theme.ContentCopyIcon(), func() {
					fyne.CurrentApp().Driver().AllWindows()[0].Clipboard().SetContent(
						strings.Join(RuntimeBackend.BackendsList[0].RecentLog, "\n"),
					)
				})
				LogText := CustomWidget.NewLogTextWithData(Fields.DataBindings.LogBinding)
				LogText.AutoScroll = true
				LogText.ReadOnly = true
				sendErrorReportButton := widget.NewButtonWithIcon(lang.L("Send error report"), theme.MailSendIcon(), func() { RuntimeBackend.ErrorReportWithLog(logWindow) })
				logWindow.SetContent(container.NewBorder(nil, container.NewHBox(copyLogButton, sendErrorReportButton), nil, nil, LogText))
				logWindow.Resize(fyne.CurrentApp().Driver().AllWindows()[0].Canvas().Size())
				logWindow.Show()
			},
		},
	})

	go processingStopTimer()
	go realtimeLabelHideTimer()
	go ProcessReceiveMessageChannel()

	flag.Parse()
	log.SetFlags(0)

	//interrupt := make(chan os.Signal, 1)
	signal.Notify(c.InterruptChan, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: c.Addr, Path: "/"}
	if host, port, err := net.SplitHostPort(c.Addr); err == nil {
		if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
			loopback := "127.0.0.1"
			if ip.To4() == nil {
				loopback = "::1"
			}
			u.Host = net.JoinHostPort(loopback, port)
		}
	}
	log.Printf("connecting to %s", u.String())

	fyne.Do(func() {
		connectingStateContainer.Add(widget.NewLabel(lang.L("Connecting to Server", map[string]interface{}{"ServerUri": u.String()})))
		connectingStateDialog.Show()
	})

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

	fyne.Do(func() {
		connectingStateDialog.Hide()
	})
	previouslyConnected = true
	if !runBackend {
		fyne.Do(RemoteAudioView.StartIntegrated)
	}

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
						fyne.Do(func() {
							connectingStateDialog.Show()
						})
						previouslyConnected = false
					}
					c.Conn, _, err = dialer.Dial(u.String(), nil)
					time.Sleep(500 * time.Millisecond) // make sure to multiply by time.Millisecond
					fyne.Do(func() {
						connectingStateDialog.Hide()
					})
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
				if !runBackend {
					fyne.Do(RemoteAudioView.StartIntegrated)
				}
				continue
			}

			previouslyConnected = true

			// A WebSocket message may span many reads (notably returned WAV files).
			raw, readErr := io.ReadAll(r)
			if readErr != nil {
				log.Println("read message:", readErr)
				continue
			}
			if !RemoteAudioView.HandleIntegratedReceive(raw) {
				ReceiveMessageChannel <- raw
			}

		}
	}()

	go func() {
		defer Logging.GoRoutineErrorHandler(func(scope *sentry.Scope) {
			scope.SetTag("GoRoutine", "Websocket\\client->Start#sendMessageChannelRoutine")
		})
		for {
			select {
			//case <-done:
			//	return
			case message := <-c.sendMessageChan:
				if RemoteAudioView.HandleIntegratedMessage(&message) {
					continue
				}
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
