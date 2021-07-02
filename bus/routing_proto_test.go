package bus_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/stretchr/testify/require"
)

type routingCmd struct {
	bus.CommandType
}

func (routingCmd) Command() string {
	return "routingCmd"
}

func (routingCmd) Valid() error {
	return nil
}

type routingCmdHandler struct {
}

func routingCmdMiddleware(next bus.CommandHandler) bus.CommandHandler {
	return bus.CmdMiddlewareFunc(func(ctx context.Context, c bus.Command) (bus.CommandResponse, []message.Message) {
		return next.Execute(ctx, c)
	})
}

func (routingCmdHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
	return
}

func TestRouteBasicCommand(t *testing.T) {
	r := bus.NewCommandContext()
	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.IsType(t, routingCmdHandler{}, c.Handler)
}

func TestRouteBasicCommandWithGlobalMiddleware(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Use(routingCmdMiddleware)

		b.Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 1)
	assert.IsType(t, routingCmdHandler{}, c.Handler)
}

func TestRouteBasicCommandGlobalMiddlewareDeclarative(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})

		b.Use(routingCmdMiddleware)
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 1)
}

func TestRouteMultipleMiddleware(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Use(routingCmdMiddleware, routingCmdMiddleware)

		b.Command(routingCmd{}).Handled(routingCmdHandler{})

		b.Use(routingCmdMiddleware)
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 3)
}
