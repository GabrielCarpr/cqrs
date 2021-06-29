// +build !unit

package queue_test

import (
	"cqrs/bus"
	"cqrs/bus/message"
	"cqrs/bus/queue/sql"
	"cqrs/log"
	"encoding/gob"
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

type TestConfig struct {
}

func (TestConfig) DBDsn() string {
	return "user=cqrs password=cqrs dbname=cqrs host=db sslmode=disable"
}

func TestQueueIntegrationTest(t *testing.T) {
	suite.Run(t, &QueueIntegrationTest{queue: sql.NewSQLQueue(TestConfig{})})
}

type QueueIntegrationTest struct {
	suite.Suite

	queue bus.Queue
}

func (s *QueueIntegrationTest) SetupTest() {
	s.queue.RegisterCtxKey(log.CtxIDKey, func(b []byte) interface{} {
		return uuid.MustParse(string(b))
	})
	gob.Register(testCmd{})
}

func (s *QueueIntegrationTest) TearDownTest() {
	s.queue.Close()
}

func (s QueueIntegrationTest) TestPublishesAndSubscribes() {
	returned := make(chan string)
	idReturn := make(chan uuid.UUID)
	cmd := testCmd{Result: "test"}

	result := ""
	idResult := uuid.Nil
	topCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s.queue.Subscribe(topCtx, func(ctx context.Context, msg message.Message) error {
			c := msg.(testCmd)
			id := log.GetID(ctx)
			idReturn <- id
			returned <- c.Result
			return nil
		})
	}()

	ctx := log.WithID(context.Background())
	err := s.queue.Publish(ctx, cmd)
	s.NoError(err)

	received := 0
	for {
		select {
		case r := <-returned:
			result = r
			received++
		case r := <-idReturn:
			idResult = r
			received++
		case <-time.After(time.Millisecond * 2700):
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
