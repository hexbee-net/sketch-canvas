package main

import (
	"context"
	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/hexbee-net/sketch-canvas/pkg/server"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetHandler(text.New(os.Stdout))

	var args struct {
		wait         time.Duration
		port         int
		verbose      bool
		superVerbose bool

		storeOptions datastore.Options
	}
	flag.DurationVarP(&args.wait, "graceful-timeout", "w", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.IntVarP(&args.port, "port", "p", 8800, "the port number of the canvas server")

	flag.BoolVarP(&args.verbose, "verbose", "v", false, "verbose mode")
	flag.BoolVar(&args.superVerbose, "vv", false, "super-verbose mode")

	flag.StringVarP(&args.storeOptions.Addr, "datastore", "s", "localhost:6379", "the hostname of the redis datastore")
	flag.StringVar(&args.storeOptions.Password, "datastore-password", "", "the password of the redis server")
	flag.IntVar(&args.storeOptions.DB, "datastore-db", 0, "the database to be selected on the redis server")

	flag.Parse()

	log.SetLevel(log.WarnLevel)
	if args.verbose {
		log.SetLevel(log.InfoLevel)
	}
	if args.superVerbose {
		log.SetLevel(log.DebugLevel)
	}

	srv := server.New(args.port, &args.storeOptions)

	ctx := context.WithValue(context.Background(), "wait", args.wait)
	srv.Start(ctx)

	os.Exit(0)
}
