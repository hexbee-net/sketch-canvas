package main

import (
	"context"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	flags "github.com/spf13/pflag"

	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
	"github.com/hexbee-net/sketch-canvas/pkg/server"
)

const (
	defaultWait = time.Second * 15
	defaultPort = 8800
)

func main() {
	log.SetHandler(text.New(os.Stdout))

	var args struct {
		wait    time.Duration
		port    int
		verbose bool
		debug   bool

		storeOptions datastore.RedisOptions
	}

	flags.DurationVarP(&args.wait, "graceful-timeout", "w", defaultWait, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flags.IntVarP(&args.port, "port", "p", defaultPort, "the port number of the canvas server")

	flags.BoolVarP(&args.verbose, "verbose", "v", false, "verbose mode")
	flags.BoolVar(&args.debug, "debug", false, "debug mode")

	flags.StringVarP(&args.storeOptions.Addr, "datastore", "s", "localhost:6379", "the hostname of the redis datastore")
	flags.StringVar(&args.storeOptions.Password, "datastore-password", "", "the password of the redis server")
	flags.IntVar(&args.storeOptions.DB, "datastore-db", 0, "the database to be selected on the redis server")

	flags.Parse()

	log.SetLevel(log.WarnLevel)

	if args.verbose {
		log.SetLevel(log.InfoLevel)
	}

	if args.debug {
		log.SetLevel(log.DebugLevel)
	}

	srv, _ := server.New(args.port, &args.storeOptions)

	ctx := context.WithValue(context.Background(), server.WaitContextKey, args.wait)
	srv.Start(ctx)

	os.Exit(0)
}
