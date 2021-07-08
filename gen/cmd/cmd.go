package cmd

import (
	"log"
	"os"

	"github.com/GabrielCarpr/cqrs/gen/gen"
	initCmd "github.com/GabrielCarpr/cqrs/gen/init"
	"github.com/GabrielCarpr/cqrs/gen/make"
)

func Execute() {
	if len(os.Args) == 1 {
		log.Fatal("Must provide a command")
	}

	switch os.Args[1] {
	case "gen":
		gen.Graphql()
	case "make":
		make.Make(os.Args[2:]...)
	case "init":
		initCmd.Init(os.Args[2:]...)
	default:
		log.Fatalf("%s is not a valid command", os.Args[1])
	}
}
