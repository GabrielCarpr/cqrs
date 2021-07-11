package errors

import (
	"context"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
)

// CommandErrorMiddleware blocks internal errors from escaping interfaces
func CommandErrorMiddleware(next bus.CommandHandler) bus.CommandHandler {
	return bus.CmdMiddlewareFunc(func(ctx context.Context, c bus.Command) (bus.CommandResponse, []message.Message) {
		res, msgs := next.Execute(ctx, c)
		switch res.Error.(type) {
		case nil:
		case Error:
			return res, msgs
		}
		res.Error = InternalServerError
		return res, msgs
	})
}

// QueryErrorMiddleware blocks internal errors from escaping query interfaces
func QueryErrorMiddleware(next bus.QueryHandler) bus.QueryHandler {
	return bus.QueryMiddlewareFunc(func(ctx context.Context, q bus.Query, res interface{}) error {
		err := next.Execute(ctx, q, res)
		switch err.(type) {
		case nil:
		case Error:
			return err
		}
		return InternalServerError
	})
}
