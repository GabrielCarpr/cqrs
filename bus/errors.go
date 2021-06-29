package bus

import (
	"errors"
	"fmt"
)

var (
	InvalidQueryResultError = errors.New("Query result must be a pointer")
)

// NoCommandHandler is an error returned when a command's handler cannot be found
type NoCommandHandler struct {
	Cmd Command
}

func (e NoCommandHandler) Error() string {
	return fmt.Sprintf("No command handler for command: %s", e.Cmd.Command())
}

type NoQueryHandler struct {
	Query Query
}

func (e NoQueryHandler) Error() string {
	return fmt.Sprintf("No query handler for query: %s", e.Query.Query())
}
