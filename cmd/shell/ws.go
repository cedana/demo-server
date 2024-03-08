package main

import (
	"bytes"
	"encoding/json"
	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

type Size struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

type HandlerOpts struct {
	Arguments          []string
	Command            string
}

var maxBufferSize = 1024

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 0,
	ReadBufferSize:   maxBufferSize,
	WriteBufferSize:  maxBufferSize,
}

func websocketHandler(opts HandlerOpts) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		shell := opts.Command
		args := opts.Arguments
		cmd := exec.Command(shell, args...)
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

		// tty -> xterm
		go func() {
			for {
				buffer := make([]byte, maxBufferSize)
				readLength, err := tty.Read(buffer)
				if err != nil {
					waiter.Done()
					return
				}
				if err := connection.WriteMessage(websocket.BinaryMessage, buffer[:readLength]); err != nil {
					continue
				}
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
				// skip invalid len
				if dataLength == -1 {
					continue
				}

				if messageType == websocket.BinaryMessage {
					if dataBuffer[0] == 1 {
						ttySize := &Size{}
						if err := json.Unmarshal(dataBuffer[1:], ttySize); err != nil {
							continue
						}
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
