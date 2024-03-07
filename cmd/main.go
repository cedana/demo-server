package main

import (
	"cedana-shell/xterm"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

func main() {
	command := cobra.Command{
		Use:  "cedana-shell",
		RunE: run,
	}
	conf.ApplyToCobra(&command)
	command.Execute()
}

func run(_ *cobra.Command, _ []string) error {
	command := conf.GetString("command")
	arguments := conf.GetStringSlice("arguments")
	allowedHostnames := conf.GetStringSlice("allowed-hostnames")
	keepalivePingTimeout := time.Duration(conf.GetInt("keepalive-ping-timeout")) * time.Second
	maxBufferSizeBytes := conf.GetInt("max-buffer-size-bytes")
	pathXTermJS := conf.GetString("path-xtermjs")
	serverAddress := conf.GetString("server-addr")
	serverPort := conf.GetInt("server-port")
	workDir := conf.GetString("workdir")
	router := mux.NewRouter()

	xtermjsHandlerOptions := xterm.HandlerOpts{
		AllowedHostnames:     allowedHostnames,
		Arguments:            arguments,
		Command:              command,
		KeepalivePingTimeout: keepalivePingTimeout,
		MaxBufferSizeBytes:   maxBufferSizeBytes,
	}
	router.HandleFunc(pathXTermJS, xterm.GetHandler(xtermjsHandlerOptions))

	publicAssetsDirectory := path.Join(workDir, "./public")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%v", serverAddress, serverPort),
		Handler: handler(router),
	}
	return server.ListenAndServe()
}

func handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
