package bus

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidQueryResult indicates a programming error where the query result
	// wasn't passed as a pointer to be filled
	ErrInvalidQueryResult = errors.New("Query result must be a pointer")
)

// NoCommandHandler is an error returned when a command's handler cannot be found
type NoCommandHandler struct {
	Cmd Command
}

func (e NoCommandHandler) Error() string {
	return fmt.Sprintf("No command handler for command: %s", e.Cmd.Command())
}

// NoQueryHandler is an error returned when a query's handler cannot be found
type NoQueryHandler struct {
	Query Query
}

func (e NoQueryHandler) Error() string {
	return fmt.Sprintf("No query handler for query: %s", e.Query.Query())
}
