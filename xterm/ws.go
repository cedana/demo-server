package xterm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

type TTYSize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

type HandlerOpts struct {
	AllowedHostnames     []string
	Arguments            []string
	Command              string
	KeepalivePingTimeout time.Duration
	MaxBufferSizeBytes   int
}

func GetHandler(opts HandlerOpts) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		maxBufferSizeBytes := opts.MaxBufferSizeBytes
		keepalivePingTimeout := opts.KeepalivePingTimeout
		allowedHostnames := opts.AllowedHostnames
		upgrader := getConnectionUpgrader(allowedHostnames, maxBufferSizeBytes)
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		terminal := opts.Command
		args := opts.Arguments
		cmd := exec.Command(terminal, args...)
		cmd.Env = os.Environ()
		tty, _ := pty.Start(cmd)
		defer func() {
			cmd.Process.Kill()
			cmd.Process.Wait()
			tty.Close()
			connection.Close()
		}()

		var waiter sync.WaitGroup
		waiter.Add(1)

		// keep-alive loop
		lastPongTime := time.Now()
		connection.SetPongHandler(func(msg string) error {
			lastPongTime = time.Now()
			return nil
		})
		go func() {
			for {
				connection.WriteMessage(websocket.PingMessage, []byte("keepalive"))
				time.Sleep(keepalivePingTimeout / 2)
				if time.Since(lastPongTime) > keepalivePingTimeout {
					waiter.Done()
					return
				}
			}
		}()

		// tty -> xterm
		go func() {
			for {
				buffer := make([]byte, maxBufferSizeBytes)
				readLength, _ := tty.Read(buffer)
				connection.WriteMessage(websocket.BinaryMessage, buffer[:readLength])
			}
		}()

		// tty <- xterm
		go func() {
			for {
				messageType, data, err := connection.ReadMessage()
				if err != nil {
					return
				}
				dataLength := len(data)
				dataBuffer := bytes.Trim(data, "\x00")
				if dataLength == -1 { // invalid
					continue
				}

				// handle size
				if messageType == websocket.BinaryMessage {
					if dataBuffer[0] == 1 {
						ttySize := &TTYSize{}
						resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
						json.Unmarshal(resizeMessage, ttySize)
						pty.Setsize(tty, &pty.Winsize{
							Rows: ttySize.Rows,
							Cols: ttySize.Cols,
						})
						continue
					}
				}
				tty.Write(dataBuffer)
			}
		}()

		waiter.Wait()
	}
}
