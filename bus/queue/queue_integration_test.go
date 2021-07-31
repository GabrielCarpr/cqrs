// +build !unit

package queue_test

import (
	"encoding/gob"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/bus/queue/sql"
	"github.com/GabrielCarpr/cqrs/log"
	"time"

	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type testCmd struct {
	bus.CommandType

	Result string
}

func (testCmd) Command() string {
	return "testcmd"
}

func (testCmd) Valid() error {
	return nil
}

var TestConfig = sql.Config{
	DBName: "cqrs",
	DBHost: "db",
	DBUser: "cqrs",
	DBPass: "cqrs",
}

func TestSQLQueueIntegrationTest(t *testing.T) {
	suite.Run(t, &QueueIntegrationTest{queue: sql.NewSQLQueue(TestConfig)})
}

type QueueIntegrationTest struct {
	suite.Suite

	queue bus.Queue
}

func (s *QueueIntegrationTest) SetupTest() {
	sql.ResetSQLDB(TestConfig.DBDsn())
	bus.RegisterContextKey(log.CtxIDKey, uuid.Nil)
	gob.Register(testCmd{})
}

func (s *QueueIntegrationTest) TearDownTest() {
	sql.ResetSQLDB(TestConfig.DBDsn())
}

func (s QueueIntegrationTest) TestPublishesAndSubscribes() {
	returned := make(chan string)
	idReturn := make(chan uuid.UUID)
	cmd := testCmd{Result: "test"}

	result := ""
	idResult := uuid.Nil
	topCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx := log.WithID(context.Background())
	err := s.queue.Publish(ctx, cmd)
	s.NoError(err)

	go func() {
		s.queue.Subscribe(topCtx, func(ctx context.Context, msg message.Message) error {
			defer cancel()
			c := msg.(testCmd)
			id := log.GetID(ctx)
			idReturn <- id
			returned <- c.Result
			return nil
		})
	}()

	received := 0
	for {
		select {
		case r := <-returned:
			result = r
			received++
		case r := <-idReturn:
			idResult = r
			received++
		case <-time.After(time.Millisecond * 3500):
			cancel()
			s.T().Error("Timed out")
			return
		}
		if received >= 2 {
			break
		}
	}

	s.Equal("test", result)
	s.NotEqual(idResult, uuid.Nil)
}
