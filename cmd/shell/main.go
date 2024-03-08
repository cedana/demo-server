package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/usvc/go-config"
)

var conf = config.Map{
	"args": &config.StringSlice{
		Default: []string{},
	},
	"cmd": &config.String{
		Default: "/bin/bash",
	},
	"workdir": &config.String{
		Default: ".",
	},
}

func run(_ *cobra.Command, _ []string) error {
	cmd := conf.GetString("cmd")
	args := conf.GetStringSlice("args")
	workDir := conf.GetString("workdir")

	router := mux.NewRouter()
	xtermOpts := HandlerOpts{
		Arguments: args,
		Command:   cmd,
	}
	router.HandleFunc("/xterm.js", websocketHandler(xtermOpts))

	depDir := path.Join(workDir, "./node_modules")
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depDir))))

	assetsDir := path.Join(workDir, "./public")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(assetsDir)))

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%v", "0.0.0.0", 8376),
		Handler: handler(router),
	}
	return server.ListenAndServe()
}

func handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recover()
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	command := cobra.Command{
		Use:  "shell",
		RunE: run,
	}
	conf.ApplyToCobra(&command)
	command.Execute()
}
