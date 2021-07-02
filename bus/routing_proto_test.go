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

func (routingCmdHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
	return
}

func TestRouteBasicCommand(t *testing.T) {
	r := bus.NewCommandContext()
	func(b bus.CmdBuilder) {
		b.Command(routingCmd{}).Handled(routingCmdHandler{})
	}(r)

	routes := r.Routes()
	c, ok := routes[routingCmd{}.Command()]
	require.True(t, ok)
	assert.IsType(t, routingCmdHandler{}, c.Handler)
}
