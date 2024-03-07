package main

import (
	"github.com/usvc/go-config"
)

var conf = config.Map{
	"allowed-hostnames": &config.StringSlice{
		Default: []string{"localhost"},
	},
	"arguments": &config.StringSlice{
		Default: []string{},
	},
	"command": &config.String{
		Default: "/bin/bash",
	},
	"keepalive-ping-timeout": &config.Int{
		Default: 20,
	},
	"max-buffer-size-bytes": &config.Int{
		Default: 512,
	},
	"path-xtermjs": &config.String{
		Default: "/xterm.js",
	},
	"server-addr": &config.String{
		Default: "0.0.0.0",
	},
	"server-port": &config.Int{
		Default: 8376,
	},
	"workdir": &config.String{
		Default: ".",
	},
}
