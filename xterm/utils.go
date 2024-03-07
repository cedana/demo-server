package xterm

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func getConnectionUpgrader(
	allowedHostnames []string,
	maxBufferSizeBytes int,
) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			requesterHostname := r.Host
			if strings.Contains(requesterHostname, ":") {
				requesterHostname = strings.Split(requesterHostname, ":")[0]
			}
			for _, allowedHostname := range allowedHostnames {
				if requesterHostname == allowedHostname {
					return true
				}
			}
			return false
		},
		HandshakeTimeout: 0,
		ReadBufferSize:   maxBufferSizeBytes,
		WriteBufferSize:  maxBufferSizeBytes,
	}
}
