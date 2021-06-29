package bus

import (
	"context"
	"cqrs/bus/message"
	"cqrs/log"
)

/*
 * Guards
**/

type CommandGuard = func(context.Context, Command) (context.Context, Command, error)

func CommandValidationGuard(ctx context.Context, c Command) (context.Context, Command, error) {
	err := c.Valid()
	if err != nil {
		return ctx, c, err
	}

	return ctx, c, nil
}

type QueryGuard = func(context.Context, Query) (context.Context, Query, error)

func QueryValidationGuard(ctx context.Context, q Query) (context.Context, Query, error) {
	err := q.Valid()
	if err != nil {
		return ctx, q, err
	}
	return ctx, q, nil
}

/*
 * Command Middleware
 */

type BaseCommandMiddleware struct {
	ExecuteMethod func(context.Context, Command) (CommandResponse, []message.Message)
}

func (m BaseCommandMiddleware) Execute(ctx context.Context, c Command) (CommandResponse, []message.Message) {
	return m.ExecuteMethod(ctx, c)
}

func CmdFunc(fn func(context.Context, Command) (CommandResponse, []message.Message)) CommandHandler {
	handler := struct{ BaseCommandMiddleware }{}
	handler.ExecuteMethod = fn
	return handler
}

type CommandMiddleware = func(CommandHandler) CommandHandler

func CommandLoggingMiddleware(next CommandHandler) CommandHandler {
	return CmdFunc(func(ctx context.Context, c Command) (res CommandResponse, msgs []message.Message) {
		log.Info(ctx, "Executing command", log.F{"command": c.Command()})

		res, msgs = next.Execute(ctx, c)

		log.Info(ctx, "Finished executing command", log.F{"command": c.Command()})
		return
	})
}

/*
 * Query middleware
 */

type BaseQueryMiddleware struct {
	ExecuteMethod func(context.Context, Query, interface{}) error
}

func (m BaseQueryMiddleware) Execute(ctx context.Context, q Query, result interface{}) error {
	return m.ExecuteMethod(ctx, q, result)
}

func QueryFunc(fn func(context.Context, Query, interface{}) error) QueryHandler {
	handler := struct{ BaseQueryMiddleware }{}
	handler.ExecuteMethod = fn
	return handler
}

type QueryMiddleware = func(QueryHandler) QueryHandler

func QueryLoggingMiddleware(next QueryHandler) QueryHandler {
	return QueryFunc(func(ctx context.Context, q Query, res interface{}) (err error) {
		log.Info(ctx, "Executing query", log.F{"query": q.Query()})

		err = next.Execute(ctx, q, res)

		log.Info(ctx, "Finished executing query", log.F{"query": q.Query()})
		return
	})
}
