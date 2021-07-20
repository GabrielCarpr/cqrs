package errors

import (
	"context"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/log"
)

// CommandErrorMiddleware blocks internal errors from escaping interfaces
func CommandErrorMiddleware(next bus.CommandHandler) bus.CommandHandler {
	return bus.CmdMiddlewareFunc(func(ctx context.Context, c bus.Command) (bus.CommandResponse, []message.Message) {
		res, msgs := next.Execute(ctx, c)
		if res.Error == nil {
			return res, msgs
		}
		if _, ok := res.Error.(Error); ok {
			return res, msgs
		}

		log.Error(ctx, res.Error, log.F{})
		res.Error = InternalServerError
		return res, msgs
	})
}

// QueryErrorMiddleware blocks internal errors from escaping query interfaces
func QueryErrorMiddleware(next bus.QueryHandler) bus.QueryHandler {
	return bus.QueryMiddlewareFunc(func(ctx context.Context, q bus.Query, res interface{}) error {
		err := next.Execute(ctx, q, res)
		if err == nil {
			return err
		}
		if err, ok := err.(Error); ok {
			return err
		}

		log.Error(ctx, err, log.F{})
		return InternalServerError
	})
}
