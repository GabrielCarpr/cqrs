// +build !unit

package bus_test

import (
	"context"
	"fmt"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/bus/queue/sql"
	"testing"
	"time"

	"github.com/sarulabs/di/v2"
	"github.com/stretchr/testify/assert"
)

type testConfig struct{}

func (testConfig) DBDsn() string {
	return "user=cqrs password=cqrs dbname=cqrs host=db sslmode=disable"
}

type testEvent struct {
	bus.EventType

	Value string
}

func (testEvent) Event() string {
	return "testEvent"
}

type syncTestEventHandler struct {
	handle func(context.Context, bus.Event) ([]message.Message, error)
}

func (h syncTestEventHandler) Handle(ctx context.Context, e bus.Event) ([]message.Message, error) {
	return h.handle(ctx, e)
}

func (syncTestEventHandler) Async() bool {
	return false
}

type asyncTestEventHandler struct {
	handle func(context.Context, bus.Event) ([]message.Message, error)
}

func (h asyncTestEventHandler) Handle(ctx context.Context, e bus.Event) ([]message.Message, error) {
	return h.handle(ctx, e)
}

func (asyncTestEventHandler) Async() bool {
	return true
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
		Name: bus.CommandHandlerName(testCmdHandler{}),
		Build: func(ctn di.Container) (interface{}, error) {
			return testCmdHandler{}, nil
		},
	})

	builder.Add(di.Def{
		Name: bus.QueryHandlerName(testQueryHandler{}),
		Build: func(ctn di.Container) (interface{}, error) {
			return testQueryHandler{}, nil
		},
	})

	return builder
}

func TestBusHandlesEvent(t *testing.T) {
	sql.ResetSQLDB(testConfig{}.DBDsn())
	syncResult := ""
	asyncResult := ""

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	sync := syncTestEventHandler{}
	sync.handle = func(c context.Context, e bus.Event) (msgs []message.Message, err error) {
		event := e.(*testEvent)
		syncResult = event.Value
		return
	}

	async := asyncTestEventHandler{}
	async.handle = func(c context.Context, e bus.Event) (msgs []message.Message, err error) {
		defer cancel()
		event := e.(*testEvent)
		asyncResult = event.Value
		return
	}

	build := setupContainer()
	build.Add(di.Def{
		Name: "event-sync-handler",
		Build: func(ctn di.Container) (interface{}, error) {
			return sync, nil
		},
	})
	build.Add(di.Def{
		Name: "event-async-handler",
		Build: func(ctn di.Container) (interface{}, error) {
			return async, nil
		},
	})
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.ExtendEvents(bus.EventRules{
		&testEvent{}: []string{"event-sync-handler", "event-async-handler"},
	})

	err := b.Publish(context.Background(), &testEvent{Value: "Hello world"})
	assert.NoError(t, err)

	assert.Equal(t, "Hello world", syncResult)
	assert.Empty(t, asyncResult)

	b.Work(ctx)
	assert.Equal(t, "Hello world", asyncResult)
}

func TestBusHandlesCommands(t *testing.T) {
	build := setupContainer()
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.Use(bus.CommandValidationGuard)
	b.ExtendCommands(func(b bus.CmdBuilder) {
		b.Command(stringReturnCmd{}).Handled(testCmdHandler{})
	})

	res, err := b.Dispatch(context.Background(), stringReturnCmd{Return: "hello"}, true)
	assert.NoError(t, err)
	assert.Equal(t, "hello", res.ID)

	res, err = b.Dispatch(context.Background(), stringReturnCmd{}, true)
	assert.Error(t, err)
	assert.Error(t, res.Error)
	assert.EqualError(t, err, "Return must be provided")
}

type queueHandler struct {
	execute func(context.Context, bus.Command) (bus.CommandResponse, []message.Message)
}

func (h queueHandler) Execute(ctx context.Context, c bus.Command) (bus.CommandResponse, []message.Message) {
	return h.execute(ctx, c)
}

func TestBusQueueCommand(t *testing.T) {
	sql.ResetSQLDB(testConfig{}.DBDsn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	result := ""
	build := setupContainer()

	h := queueHandler{}
	h.execute = func(ctx context.Context, c bus.Command) (res bus.CommandResponse, msgs []message.Message) {
		defer cancel()
		cmd := c.(stringReturnCmd)
		result = cmd.Return
		return
	}

	build.Add(di.Def{
		Name: bus.CommandHandlerName(h),
		Build: func(ctn di.Container) (interface{}, error) {
			return h, nil
		},
	})
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.ExtendCommands(func(b bus.CmdBuilder) {
		b.Command(stringReturnCmd{}).Handled(h)
	})

	res, err := b.Dispatch(context.Background(), stringReturnCmd{Return: "hello"}, false)
	assert.NoError(t, err)
	assert.Nil(t, res)

	defer cancel()
	b.Work(ctx)

	assert.Equal(t, "hello", result)
}

func TestBusHandlesQueries(t *testing.T) {
	build := setupContainer()
	b := bus.NewBus(testConfig{}, build, []bus.BoundedContext{})
	b.Use(bus.QueryValidationGuard)
	b.ExtendQueries(func(b bus.QueryBuilder) {
		b.Query(returnQuery{}).Handled(testQueryHandler{})
	})

	var res string
	err := b.Query(context.Background(), returnQuery{Return: "Hii!!"}, &res)
	assert.NoError(t, err)
	assert.Equal(t, "Hii!!", res)

	err = b.Query(context.Background(), returnQuery{}, &res)
	assert.Error(t, err)
	assert.EqualError(t, err, "Query: Return must be provided")
}
