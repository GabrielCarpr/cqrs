package bus

import (
	"context"
	"cqrs/bus/message"
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

func Get(ctx context.Context, key string) interface{} {
	c := getCtn(ctx)
	return c.Get(key)
}

func (b *Bus) queryContainerGuard(ctx context.Context, q Query) (context.Context, Query, error) {
	requestCtn, _ := b.Container.SubContainer()
	ctx = context.WithValue(ctx, ctxCtnKey, requestCtn)
	return ctx, q, nil
}

func (b *Bus) queryContainerMiddleware(next QueryHandler) QueryHandler {
	return QueryFunc(func(ctx context.Context, q Query, res interface{}) error {
		err := next.Execute(ctx, q, res)

		ctn := getCtn(ctx)
		ctn.Delete()
		return err
	})
}

func (b *Bus) commandContainerGuard(ctx context.Context, c Command) (context.Context, Command, error) {
	requestCtn, _ := b.Container.SubContainer()
	ctx = context.WithValue(ctx, ctxCtnKey, requestCtn)
	return ctx, c, nil
}

func (b *Bus) commandContainerMiddleware(next CommandHandler) CommandHandler {
	return CmdFunc(func(ctx context.Context, c Command) (CommandResponse, []message.Message) {
		res, msgs := next.Execute(ctx, c)

		ctn := getCtn(ctx)
		ctn.Delete()
		return res, msgs
	})
}
