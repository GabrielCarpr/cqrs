package make

import (
	"log"
	"strings"
)

func Make(args ...string) {
	command := args[2]
	switch strings.ToLower(command) {
	case "command":
	case "query":
	case "test":
	default:
		log.Fatalf("%s is not a valid type", command)
	}
}
