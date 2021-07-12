package bus

import (
	"context"
	"fmt"
	"reflect"

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
	requestCtn, _ := b.container.SubContainer()
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
	requestCtn, _ := b.container.SubContainer()
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

type Def struct {
	Build    func(ctn di.Container) (interface{}, error)
	Close    func(obj interface{}) error
	Name     interface{}
	Scope    string
	Tags     []di.Tag
	Unshared bool
}

func (d Def) name() string {
	switch v := d.Name.(type) {
	case string:
		return v
	case CommandHandler:
		return CommandHandlerName(v)
	case QueryHandler:
		return QueryHandlerName(v)
	default:
		t := reflect.TypeOf(v)
		return fmt.Sprint(t.PkgPath(), ".", t.Name())
	}
}

func (d Def) diDef() di.Def {
	return di.Def{
		Build:    d.Build,
		Close:    d.Close,
		Name:     d.name(),
		Scope:    d.Scope,
		Tags:     d.Tags,
		Unshared: d.Unshared,
	}
}

// BoundedContext represents the integration between the main app and a BC.
type Module interface {
	EventRules() EventRules
	Commands(CmdBuilder)
	Queries(QueryBuilder)

	Services() []Def
}

type FuncModule struct {
	EventsFunc   func() EventRules
	CommandsFunc func(CmdBuilder)
	QueriesFunc  func(QueryBuilder)
	ServicesFunc func() []Def

	Defs []Def
}

func (m FuncModule) Commands(b CmdBuilder) {
	if m.CommandsFunc != nil {
		m.CommandsFunc(b)
	}
}

func (m FuncModule) Queries(b QueryBuilder) {
	if m.QueriesFunc != nil {
		m.QueriesFunc(b)
	}
}

func (m FuncModule) Services() []Def {
	if m.ServicesFunc != nil {
		return append(m.ServicesFunc(), m.Defs...)
	}
	return m.Defs
}

func (m FuncModule) EventRules() EventRules {
	if m.EventsFunc != nil {
		return m.EventsFunc()
	}
	return EventRules{}
}
