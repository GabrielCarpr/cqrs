package bus_test

import (
	"github.com/gabrielcarpr/cqrs/bus"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
	bus.EventType
	Payload string
}

func (TestEvent) Event() string {
	return "event.test"
}

func TestEventQueue(t *testing.T) {
	owner := uuid.New()
	queue := bus.NewEventQueue(owner)

	e := &TestEvent{Payload: "Hi"}
	e2 := &TestEvent{Payload: "Bye"}
	queue.Publish(e, e2)

	events := queue.Release()
	assert.Len(t, events, 2)
	events2 := queue.Release()
	assert.Len(t, events2, 0)
	event := events[0].(*TestEvent)
	assert.Equal(t, owner.String(), event.Owner)
}
