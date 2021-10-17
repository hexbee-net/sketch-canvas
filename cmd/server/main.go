package main

import (
	"flag"
	"github.com/hexbee-net/sketch-canvas/pkg/server"
	"os"
	"time"
)

func main() {
	var wait time.Duration
	var port int

	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.IntVar(&port, "port", 8800, "the port number of the canvas server")
	flag.Parse()

	srv := server.New(port)
	srv.Start(wait)

	os.Exit(0)
}
