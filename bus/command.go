package bus

import (
	"context"

	"github.com/GabrielCarpr/cqrs/bus/message"
)

// Command is a value object instructing the system to change state.
type Command interface {
	message.Message

	// Command returns the commands (unique) name, and must be implemented by every command
	Command() string

	// Valid returns an error if the command is not valid. Must be implemented by every command
	Valid() error

	// Auth returns the list of scopes required for the command to execute.
	// The list of scopes may be dynamic by using data contained within the context,
	// such as user IDs, for protecting user data
	Auth(context.Context) [][]string
}

// CommandType can be embedded in commands to provide sane defaults
// for the Command interface
type CommandType struct {
}

// MessageType implement the message.Message interface
func (c CommandType) MessageType() message.Type {
	return message.Command
}

// Auth implements the Command interface, with public access control
func (c CommandType) Auth(ctx context.Context) [][]string {
	return [][]string{}
}

// CommandResponse originates from a command when it is executed
// synchronously. If async, then the response cannot be provided.
type CommandResponse struct {
	Error error
	ID    string
}

// CommandHandler is a handler for a specific command.
// Command <-> CommandHandler has a 1:1 relationship
type CommandHandler interface {
	// Execute takes a command and executes the stateful logic it requests.
	Execute(context.Context, Command) (CommandResponse, []message.Message)
}
