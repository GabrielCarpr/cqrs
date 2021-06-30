package bus

import (
	"context"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/sarulabs/di/v2"
)

type ctxCtnKeyType string

var ctxCtnKey = ctxCtnKeyType("ctn")

func getCtn(ctx context.Context) di.Container {
	ctn := ctx.Value(ctxCtnKey)
	if ctn == nil {
		panic("Context doesn't contain service container")
	}
	return ctn.(di.Container)
}

// Get retrieves a service from the bus's DI container
func Get(ctx context.Context, key string) interface{} {
	c := getCtn(ctx)
	return c.Get(key)
}

// queryContainerGuard scopes the bus DI container, and injects it into the context
func (b *Bus) queryContainerGuard(ctx context.Context, q Query) (context.Context, Query, error) {
	requestCtn, _ := b.Container.SubContainer()
	ctx = context.WithValue(ctx, ctxCtnKey, requestCtn)
	return ctx, q, nil
}

// queryContainerMiddleware executes the query, then gets the container
// and deletes it, clearing up any resources, after the query has completed
func (b *Bus) queryContainerMiddleware(next QueryHandler) QueryHandler {
	return QueryMiddlewareFunc(func(ctx context.Context, q Query, res interface{}) error {
		ctn := getCtn(ctx)
		defer ctn.Delete()

		return next.Execute(ctx, q, res)
	})
}

// commandContainerGuard scopes the bus DI container, and injects it into the context
func (b *Bus) commandContainerGuard(ctx context.Context, c Command) (context.Context, Command, error) {
	requestCtn, _ := b.Container.SubContainer()
	ctx = context.WithValue(ctx, ctxCtnKey, requestCtn)
	return ctx, c, nil
}

// commandContainerMiddleware executes the query, then gets the container
// and deletes it, clearing up any resources, after the query has completed
func (b *Bus) commandContainerMiddleware(next CommandHandler) CommandHandler {
	return CmdMiddlewareFunc(func(ctx context.Context, c Command) (CommandResponse, []message.Message) {
		ctn := getCtn(ctx)
		defer ctn.Delete()

		return next.Execute(ctx, c)
	})
}
