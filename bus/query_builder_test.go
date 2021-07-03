package bus_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/stretchr/testify/require"
)

type routingQuery struct {
	bus.QueryType
}

func (routingQuery) Query() string {
	return "routingQuery"
}

func (routingQuery) Valid() error {
	return nil
}

type routingQuery2 struct {
	bus.QueryType
}

func (routingQuery2) Query() string {
	return "routingQuery2"
}

func (routingQuery2) Valid() error {
	return nil
}

type routingQueryHandler struct {
}

func routingQueryMiddleware(next bus.QueryHandler) bus.QueryHandler {
	return bus.QueryMiddlewareFunc(func(ctx context.Context, q bus.Query, res interface{}) error {
		return next.Execute(ctx, q, res)
	})
}

func (routingQueryHandler) Execute(ctx context.Context, q bus.Query, res interface{}) error {
	return nil
}

func TestRouteBasicQuery(t *testing.T) {
	r := bus.NewQueryContext()
	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.IsType(t, routingQueryHandler{}, c.Handler)
}

func TestRouteBasicQueryWithGlobalMiddleware(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Use(routingQueryMiddleware)

		b.Query(routingQuery{}).Handled(routingQueryHandler{})
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 1)
	assert.IsType(t, routingQueryHandler{}, c.Handler)
}

func TestRouteBasicQueryGlobalMiddlewareDeclarative(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})

		b.Use(routingQueryMiddleware)
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 1)
}

func TestRouteMultipleMiddleware(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Use(routingQueryMiddleware, routingQueryMiddleware)

		b.Query(routingQuery{}).Handled(routingQueryHandler{})

		b.Use(routingQueryMiddleware)
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 3)
}

func TestRouteInGroup(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery2{}).Handled(routingQueryHandler{})

		b.Group(func(b bus.QueryBuilder) {
			b.Query(routingQuery{}).Handled(routingQueryHandler{})

			b.Use(routingQueryMiddleware)
		})
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.IsType(t, routingQueryHandler{}, c.Handler)
	assert.IsType(t, routingQuery{}, c.Query)
	assert.Len(t, c.Middleware, 1)

	c2, ok := r.Route(routingQuery2{})
	require.True(t, ok)
	assert.IsType(t, routingQueryHandler{}, c2.Handler)
	assert.IsType(t, routingQuery2{}, c2.Query)
	assert.Len(t, c2.Middleware, 0)
}

func TestRouteGroupNoQuery(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Group(func(b bus.QueryBuilder) {
			b.Query(routingQuery{}).Handled(routingQueryHandler{})
		})
	}(r)

	_, ok := r.Route(routingQuery2{})
	require.False(t, ok)
}

func TestRouteQueryWith(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})

		b.With(routingQueryMiddleware).Query(routingQuery2{}).Handled(routingQueryHandler{})
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	assert.Len(t, c.Middleware, 0)

	c2, ok := r.Route(routingQuery2{})
	require.True(t, ok)
	assert.Len(t, c2.Middleware, 1)
	assert.IsType(t, routingQuery2{}, c2.Query)
	assert.IsType(t, routingQueryHandler{}, c2.Handler)
}

func TestPanicsDuplicateQuerys(t *testing.T) {
	r := bus.NewQueryContext()
	panicked := false

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})
		b.Query(routingQuery{}).Handled(routingQueryHandler{})
	}(r)

	require.True(t, panicked, "Did not panic")
}

func TestEmptyBuilder(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {

	}(r)
}

func TestAppliesEachContextMiddleware(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Use(routingQueryMiddleware)

		b.Group(func(b bus.QueryBuilder) {
			b.Use(routingQueryMiddleware)

			b.Group(func(b bus.QueryBuilder) {
				b.Use(routingQueryMiddleware)
				b.Query(routingQuery{}).Handled(routingQueryHandler{})
			})
		})
	}(r)

	c, ok := r.Route(routingQuery{})
	require.True(t, ok)
	require.Len(t, c.Middleware, 3)
}

func TestCreatesRoutingTable(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Use(routingQueryMiddleware)

		b.Query(routingQuery{}).Handled(routingQueryHandler{})

		b.Group(func(b bus.QueryBuilder) {
			b.Use(routingQueryMiddleware)
			b.Query(routingQuery2{}).Handled(routingQueryHandler{})
		})
	}(r)

	routes := r.Routes()
	require.Len(t, routes, 2)
	assert.Len(t, routes[routingQuery{}.Query()].Middleware, 1)
	assert.Len(t, routes[routingQuery2{}.Query()].Middleware, 2)
}

func TestCannotTakeMultipleQuerys(t *testing.T) {
	r := bus.NewQueryContext()
	panicked := false

	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})

		b.Query(routingQuery{}).Handled(routingQueryHandler{})
	}(r)

	require.True(t, panicked)
}

func TestSelfTestMultipleQuerys(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{}).Handled(routingQueryHandler{})
		b.With().Query(routingQuery{}).Handled(routingQueryHandler{})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}

func TestSelfTestMultipleQuerysSiblings(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.With().Query(routingQuery{}).Handled(routingQueryHandler{})
		b.Group(func(b bus.QueryBuilder) {
			b.Query(routingQuery{}).Handled(routingQueryHandler{})
		})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}

func TestSelfTestQueryNoHandler(t *testing.T) {
	r := bus.NewQueryContext()

	func(b bus.QueryBuilder) {
		b.Query(routingQuery{})
	}(r)

	err := r.SelfTest()
	require.Error(t, err)
}
