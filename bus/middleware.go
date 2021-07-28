package bus

import (
	"context"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/errors"
	"github.com/GabrielCarpr/cqrs/log"
)

/*
* TODO: Add new middleware
* - Panic recovery
* - Auth/access control guard

/*
 * Guards
 * TODO: Add tests
**/

// CommandGuard allows runtime composition of code to test commands
// before being routed to a handler. Intended for use with validation and access control.
// Will always run before a command is queued
type CommandGuard = func(context.Context, Command) (context.Context, Command, error)

// CommandValidationGuard checks a command is valid before being executed, and returns an error if not
func CommandValidationGuard(ctx context.Context, c Command) (context.Context, Command, error) {
	err := c.Valid()
	if err != nil {
		return ctx, c, err
	}

	return ctx, c, nil
}

// QueryGuard allows runtime composition of code to test queries before being
// routed to a handler. Intended for use with validation and access control.
type QueryGuard = func(context.Context, Query) (context.Context, Query, error)

// QueryValidationGuard ensures a query is valid before being routed to a handler
func QueryValidationGuard(ctx context.Context, q Query) (context.Context, Query, error) {
	err := q.Valid()
	if err != nil {
		return ctx, q, err
	}
	return ctx, q, nil
}

/*
 * Command Middleware
 * TODO: Add tests
 */

type baseCommandMiddleware struct {
	ExecuteMethod func(context.Context, Command) (CommandResponse, []message.Message)
}

func (m baseCommandMiddleware) Execute(ctx context.Context, c Command) (CommandResponse, []message.Message) {
	return m.ExecuteMethod(ctx, c)
}

// CmdMiddlewareFunc is used for creating middleware with a function
func CmdMiddlewareFunc(fn func(context.Context, Command) (CommandResponse, []message.Message)) CommandHandler {
	handler := struct{ baseCommandMiddleware }{}
	handler.ExecuteMethod = fn
	return handler
}

// CommandMiddleware allows access to a command before and after it's executed by a handler
type CommandMiddleware = func(CommandHandler) CommandHandler

// CommandLoggingMiddleware logs before and after a command is executed by it's handler
func CommandLoggingMiddleware(next CommandHandler) CommandHandler {
	return CmdMiddlewareFunc(func(ctx context.Context, c Command) (res CommandResponse, msgs []message.Message) {
		log.Info(ctx, "Executing command", log.F{"command": string(c.Command())})
		defer log.Info(ctx, "Finished executing command", log.F{"command": string(c.Command())})

		return next.Execute(ctx, c)
	})
}

/*
 * Query middleware
 * TODO: Add tests
 */

type baseQueryMiddleware struct {
	ExecuteMethod func(context.Context, Query, interface{}) error
}

func (m baseQueryMiddleware) Execute(ctx context.Context, q Query, result interface{}) error {
	return m.ExecuteMethod(ctx, q, result)
}

// QueryMiddlewareFunc allows construction of a QueryMiddleware
func QueryMiddlewareFunc(fn func(context.Context, Query, interface{}) error) QueryHandler {
	handler := struct{ baseQueryMiddleware }{}
	handler.ExecuteMethod = fn
	return handler
}

// QueryMiddleware allows access to a query before and after it's executed
type QueryMiddleware = func(QueryHandler) QueryHandler

// QueryLoggingMiddleware logs the query before and after it's executed
func QueryLoggingMiddleware(next QueryHandler) QueryHandler {
	return QueryMiddlewareFunc(func(ctx context.Context, q Query, res interface{}) (err error) {
		log.Info(ctx, "Executing query", log.F{"query": q.Query()})
		defer log.Info(ctx, "Finished executing query", log.F{"query": q.Query()})

		return next.Execute(ctx, q, res)
	})
}

func CommandErrorMiddleware(next CommandHandler) CommandHandler {
	return CmdMiddlewareFunc(func(ctx context.Context, c Command) (CommandResponse, []message.Message) {
		res, msgs := next.Execute(ctx, c)
		if res.Error == nil {
			return res, msgs
		}
		if _, ok := res.Error.(errors.Error); ok {
			return res, msgs
		}

		log.Error(ctx, res.Error, log.F{})
		res.Error = errors.InternalServerError
		return res, msgs
	})
}

// QueryErrorMiddleware blocks internal errors from escaping query interfaces
func QueryErrorMiddleware(next QueryHandler) QueryHandler {
	return QueryMiddlewareFunc(func(ctx context.Context, q Query, res interface{}) error {
		err := next.Execute(ctx, q, res)
		if err == nil {
			return err
		}
		if err, ok := err.(errors.Error); ok {
			return err
		}

		log.Error(ctx, err, log.F{})
		return errors.InternalServerError
	})
}
