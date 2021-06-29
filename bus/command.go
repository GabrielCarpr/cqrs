package bus

import (
	"context"
	"cqrs/bus/message"
)

// Command is a value object instructing the system to change state.
type Command interface {
	message.Message
	Command() string
	Valid() error
	Auth(context.Context) [][]string
}

type CommandType struct {
}

func (c CommandType) MessageType() string {
	return "command"
}

func (c CommandType) Auth(ctx context.Context) [][]string {
	return [][]string{}
}

// CommandResponse originates from a command when it's executed synchronously.
type CommandResponse struct {
	Error error
	ID    string
}

// CommandHandler is a handler for a specific command.
// Command <-> CommandHandler has a 1:1 relationship
type CommandHandler interface {
	Execute(context.Context, Command) (CommandResponse, []message.Message)
}
