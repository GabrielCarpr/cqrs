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

type routingCmd2 struct {
	bus.CommandType
}

func (routingCmd2) Command() string {
	return "routingCmd2"
}

func (routingCmd2) Valid() error {
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

func TestRouteInGroup(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd2{}).Handled(routingCmdHandler{})

		b.Group(func(b bus.CmdBuilder) {
			b.Command(routingCmd{}).Handled(routingCmdHandler{})

			b.Use(routingCmdMiddleware)
		})
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.IsType(t, routingCmdHandler{}, c.Handler)
	assert.IsType(t, routingCmd{}, c.Command)
	assert.Len(t, c.Middleware, 1)

	c2, ok := r.Route(routingCmd2{})
	require.True(t, ok)
	assert.IsType(t, routingCmdHandler{}, c2.Handler)
	assert.IsType(t, routingCmd2{}, c2.Command)
	assert.Len(t, c2.Middleware, 0)
}

func TestRouteGroupNoCmd(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Group(func(b bus.CmdBuilder) {
			b.Command(routingCmd{}).Handled(routingCmdHandler{})
		})
	}(r)

	_, ok := r.Route(routingCmd2{})
	require.False(t, ok)
}

func TestRouteCmdWith(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})

		b.With(routingCmdMiddleware).Command(routingCmd2{}).Handled(routingCmdHandler{})
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 0)

	c2, ok := r.Route(routingCmd2{})
	require.True(t, ok)
	assert.Len(t, c2.Middleware, 1)
	assert.IsType(t, routingCmd2{}, c2.Command)
	assert.IsType(t, routingCmdHandler{}, c2.Handler)
}

func TestPanicsDuplicateCommands(t *testing.T) {
	r := bus.NewCommandContext()
	panicked := false

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})
		b.Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	require.True(t, panicked, "Did not panic")
}

func TestEmptyBuilder(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {

	}(r)
}

func TestAppliesEachContextMiddleware(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Use(routingCmdMiddleware)

		b.Group(func(b bus.CmdBuilder) {
			b.Use(routingCmdMiddleware)

			b.Group(func(b bus.CmdBuilder) {
				b.Use(routingCmdMiddleware)
				b.Command(routingCmd{}).Handled(routingCmdHandler{})
			})
		})
	}(r)

	c, ok := r.Route(routingCmd{})
	require.True(t, ok)
	require.Len(t, c.Middleware, 3)
}

func TestCreatesRoutingTable(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Use(routingCmdMiddleware)

		b.Command(routingCmd{}).Handled(routingCmdHandler{})

		b.Group(func(b bus.CmdBuilder) {
			b.Use(routingCmdMiddleware)
			b.Command(routingCmd2{}).Handled(routingCmdHandler{})
		})
	}(r)

	routes := r.Routes()
	require.Len(t, routes, 2)
	assert.Len(t, routes[routingCmd{}.Command()].Middleware, 1)
	assert.Len(t, routes[routingCmd2{}.Command()].Middleware, 2)
}

func TestCannotTakeMultipleCommands(t *testing.T) {
	r := bus.NewCommandContext()
	panicked := false

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})

		b.Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	require.True(t, panicked)
}

func TestSelfTestMultipleCommands(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})
		b.With().Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}

func TestSelfTestMultipleCommandsSiblings(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.With().Command(routingCmd{}).Handled(routingCmdHandler{})
		b.Group(func(b bus.CmdBuilder) {
			b.Command(routingCmd{}).Handled(routingCmdHandler{})
		})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}

func TestSelfTestCommandNoHandler(t *testing.T) {
	r := bus.NewCommandContext()

	func(b bus.CmdBuilder) {
		b.Command(routingCmd{})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}
