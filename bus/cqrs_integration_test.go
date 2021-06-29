// +build !unit

package bus_test

import (
	"context"
	"cqrs/bus"
	"cqrs/bus/message"
	"fmt"
	"testing"

	"github.com/sarulabs/di/v2"
	"github.com/stretchr/testify/assert"
)

type testConfig struct{}

func (testConfig) DBDsn() string {
	return "user=cqrs password=cqrs dbname=cqrs host=db sslmode=disable"
}

type returnQuery struct {
	bus.QueryType

	Return string
}

func (returnQuery) Query() string {
	return "returnQuery"
}

func (q returnQuery) Valid() error {
	switch true {
	case len(q.Return) == 0:
		return fmt.Errorf("Query: Return must be provided")
	}
	return nil
}

type testQueryHandler struct {
}

func (testQueryHandler) Execute(ctx context.Context, q bus.Query, res interface{}) error {
	query := q.(returnQuery)
	*res.(*string) = query.Return
	return nil
}

type testCmdHandler struct {
}

func (testCmdHandler) Execute(ctx context.Context, c bus.Command) (res bus.CommandResponse, _ []message.Message) {
	cmd := c.(stringReturnCmd)
	res = bus.CommandResponse{ID: cmd.Return}
	return
}

type stringReturnCmd struct {
	bus.CommandType

	Return string
}

func (c stringReturnCmd) Valid() error {
	switch true {
	case len(c.Return) == 0:
		return fmt.Errorf("Return must be provided")
	}
	return nil
}

func (stringReturnCmd) Command() string {
	return "string-return-cmd"
}

func setupContainer() *di.Builder {
	builder, _ := di.NewBuilder()

	builder.Add(di.Def{
		Name: "test-cmd-handler",
		Build: func(ctn di.Container) (interface{}, error) {
			return testCmdHandler{}, nil
		},
	})

	builder.Add(di.Def{
		Name: "test-query-handler",
		Build: func(ctn di.Container) (interface{}, error) {
			return testQueryHandler{}, nil
		},
	})

	return builder
}

func TestBusHandlesCommands(t *testing.T) {
	build := setupContainer()
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.Use(bus.CommandValidationGuard)
	b.ExtendCommands(bus.CommandRules{
		stringReturnCmd{}: "test-cmd-handler",
	})

	res, err := b.Dispatch(context.Background(), stringReturnCmd{Return: "hello"}, true)
	assert.NoError(t, err)
	assert.Equal(t, "hello", res.ID)

	res, err = b.Dispatch(context.Background(), stringReturnCmd{}, true)
	assert.Error(t, err)
	assert.Error(t, res.Error)
	assert.EqualError(t, err, "Return must be provided")
}

func TestBusHandlesQueries(t *testing.T) {
	build := setupContainer()
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.Use(bus.QueryValidationGuard)
	b.ExtendQueries(bus.QueryRules{
		returnQuery{}: "test-query-handler",
	})

	var res string
	err := b.Query(context.Background(), returnQuery{Return: "Hii!!"}, &res)
	assert.NoError(t, err)
	assert.Equal(t, "Hii!!", res)

	err = b.Query(context.Background(), returnQuery{}, &res)
	assert.Error(t, err)
	assert.EqualError(t, err, "Query: Return must be provided")
}
