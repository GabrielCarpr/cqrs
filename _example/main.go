package main

import (
	"example/internal/app"
	"flag"
	"log"
	"context"
)

var mode string

func init() {
	flag.StringVar(&mode, "mode", "api", "Mode to run the app in")

	flag.Parse()
}

func main() {
	app := app.Make(context.Background())
    defer app.Delete()
	log.Print("Application building complete")

	switch mode {
	case "api":
		log.Print("Listening for HTTP requests")
		app.Handle()
	case "worker":
		log.Print("Running worker")
		app.Work()
	}

	log.Print("Shutting down")
}
