package bus_test

import (
	"errors"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEvent struct {
	bus.EventType
	Payload string
}

func (TestEvent) Event() string {
	return "event.test"
}

func TestEventBufferCommits(t *testing.T) {
	owner := uuid.New()
	queue := bus.NewEventBuffer(owner, "lol")

	e := &TestEvent{Payload: "Hi"}
	e2 := &TestEvent{Payload: "Bye"}
	queue.Buffer(true, e, e2)

	events := queue.Commit()
	assert.Len(t, events, 2)
	assert.Len(t, queue.Events(), 0)
	event := events[0].(*TestEvent)
	assert.Equal(t, owner.String(), event.Owner.String())
}

func TestEventBufferVersions(t *testing.T) {
	queue := bus.NewEventBuffer(uuid.New(), "lol")
	require.Equal(t, int64(0), queue.Version)

	e := &TestEvent{Payload: "Hi"}
	queue.Buffer(false, e)
	require.Equal(t, int64(1), queue.Version)
	require.Equal(t, int64(1), queue.PendingVersion())

	queue.Buffer(true, e)
	require.Equal(t, int64(1), queue.Version)
	require.Equal(t, int64(2), queue.PendingVersion())

	msgs := queue.Commit()
	require.Len(t, msgs, 1)
	require.Equal(t, int64(2), queue.Version)
	require.Equal(t, int64(2), queue.PendingVersion())
	require.Equal(t, msgs[0], e)
}

/*
Integration test
*/

type testNameChanged struct {
	bus.EventType

	Name string
}

func (testNameChanged) Event() string {
	return "test.name.changed"
}

func newTestEntity() testEntity {
	ID := uuid.New()
	e := testEntity{
		ID:          ID,
		EventBuffer: bus.NewEventBuffer(ID, "testEntity"),
	}
	return e
}

type testEntity struct {
	ID   uuid.UUID
	Name string

	bus.EventBuffer
}

func (e *testEntity) ApplyChange(new bool, events ...bus.Event) {
	for _, event := range events {
		switch event := event.(type) {
		case *testNameChanged:
			e.Name = event.Name
		}
		e.Buffer(new, event)
	}
}

func (e testEntity) GiveName(name string) (testEntity, error) {
	if e.Name != "" {
		return e, errors.New("Already has a name")
	}

	e.ApplyChange(true, &testNameChanged{Name: name})

	return e, nil
}

func TestApplyChange(t *testing.T) {
	entity := newTestEntity()

	entity, err := entity.GiveName("Gabriel")
	require.Nil(t, err)
	require.Equal(t, "Gabriel", entity.Name)
	require.Equal(t, int64(0), entity.Version)
	require.Equal(t, int64(1), entity.PendingVersion())

	events := entity.Events()
	entity.Commit()
	require.Equal(t, "test.name.changed", events[0].Event())
	require.Equal(t, int64(1), events[0].Versioned())
	require.False(t, events[0].WasPublishedAt().IsZero())
	require.Equal(t, "testEntity", events[0].FromAggregate())
	require.Equal(t, entity.ID.String(), events[0].Owned().String())

	require.Equal(t, int64(1), entity.CurrentVersion())
}

func TestLoadFromEvents(t *testing.T) {
	events := []bus.Event{&testNameChanged{Name: "Harry Potter"}}
	entity := newTestEntity()

	entity.ApplyChange(false, events...)

	require.Equal(t, int64(1), entity.Version)
	require.Equal(t, int64(1), entity.PendingVersion())
}

func TestLoadFromData(t *testing.T) {
	entity := newTestEntity()
	entity.Name = "Harry Potter"
	entity.ForceVersion(3)

	require.Equal(t, int64(3), entity.Version)
	require.Equal(t, "Harry Potter", entity.Name)
	require.Equal(t, int64(3), entity.PendingVersion())
}
