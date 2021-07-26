// +build !unitsd

package eventstore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/eventstore"
	"github.com/GabrielCarpr/cqrs/eventstore/memory"
	"github.com/GabrielCarpr/cqrs/eventstore/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

type TestEvent struct {
	bus.EventType

	Name string
	Age  int
}

func Buffer(id uuid.UUID) bus.EventBuffer {
	return bus.NewEventBuffer(id, "testEntity")
}

func (TestEvent) Event() string {
	return "test.event"
}

func TestMemoryEventStore(t *testing.T) {
	s := &EventStoreBlackboxTest{factory: func() bus.EventStore {
		return &memory.MemoryEventStore{}
	}}
	suite.Run(t, s)
}

func TestPostgresEventStore(t *testing.T) {
	c := postgres.Config{
		DBName: "cqrs",
		DBPass: "cqrs",
		DBHost: "db",
		DBUser: "cqrs",
	}
	s := &EventStoreBlackboxTest{
		factory: func() bus.EventStore {
			return postgres.New(c)
		},
	}
	s.setupHook = func() error {
		schema := postgres.PostgreSQLSchema{Config: c}
		schema.Reset()
		return nil
	}
	suite.Run(t, s)
}

/**
Test Suite
*/

type EventStoreBlackboxTest struct {
	suite.Suite

	factory   func() bus.EventStore
	setupHook func() error

	entity      uuid.UUID
	buffer      bus.EventBuffer
	otherBuffer bus.EventBuffer
	store       bus.EventStore
}

func (s *EventStoreBlackboxTest) SetupTest() {
	s.store = s.factory()
	s.entity = uuid.New()
	s.buffer = Buffer(s.entity)
	s.otherBuffer = Buffer(uuid.New())
	bus.RegisterMessage(&TestEvent{})

	if s.setupHook != nil {
		err := s.setupHook()
		if err != nil {
			panic(err)
		}
	}
}

func (s *EventStoreBlackboxTest) TearDownTest() {
	s.store.Close()
}

func (s EventStoreBlackboxTest) TestAppendsEventsAndStreams() {
	e := &TestEvent{Name: "Gabriel", Age: 24}
	s.buffer.Buffer(true, e)
	evs := s.buffer.Events(context.Background())
	err := s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.buffer.CurrentVersion()),
		evs...,
	)
	s.NoError(err)

	e2 := &TestEvent{Name: "Gabriel", Age: 24}
	s.otherBuffer.Buffer(true, e2)
	evs2 := s.otherBuffer.Events(context.Background())
	err = s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.otherBuffer.CurrentVersion()),
		evs2...,
	)
	s.NoError(err)

	query := bus.Select{
		StreamID: bus.StreamID{Type: "testEntity", ID: s.entity.String()},
		From:     0,
	}
	stream := make(chan bus.Event)
	results := make([]bus.Event, 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return s.store.Stream(ctx, stream, query)
	})

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case ev, ok := <-stream:
			if !ok {
				break loop
			}
			results = append(results, ev)
		}
	}

	err = group.Wait()
	s.Require().NoError(err)
	s.Len(results, 1)
}

func (s EventStoreBlackboxTest) TestAppendsEventAndStreamsFrom() {
	events := make([]bus.Event, 10)
	for i := 0; i < 10; i++ {
		events[i] = &TestEvent{Name: "Gabriel", Age: 20 + i}
	}
	s.buffer.Buffer(true, events...)

	err := s.store.Append(context.Background(), bus.ExpectedVersion(s.buffer.Version), s.buffer.Events(context.Background())...)
	s.NoError(err)

	var result []bus.Event
	stream := make(chan bus.Event)
	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		return s.store.Stream(ctx, stream, bus.Select{
			StreamID: bus.StreamID{ID: s.entity.String(),
				Type: "testEntity"},
			From: 6,
		})
	})

	for event := range stream {
		result = append(result, event)
	}

	err = group.Wait()
	s.NoError(err)
	s.Len(result, 5)
}

func (s EventStoreBlackboxTest) TestStreamsAll() {
	e := &TestEvent{Name: "Gabriel", Age: 24}
	s.buffer.Buffer(true, e)
	evs := s.buffer.Events(context.Background())
	err := s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.buffer.CurrentVersion()),
		evs...,
	)
	s.NoError(err)

	e2 := &TestEvent{Name: "Gabriel", Age: 24}
	s.otherBuffer.Buffer(true, e2)
	evs2 := s.otherBuffer.Events(context.Background())
	err = s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.otherBuffer.CurrentVersion()),
		evs2...,
	)
	s.NoError(err)

	results := make([]bus.Event, 0)
	stream := make(chan bus.Event)
	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		return s.store.Stream(ctx, stream, bus.Select{})
	})

	for event := range stream {
		results = append(results, event)
	}

	err = group.Wait()
	s.NoError(err)
	s.Len(results, 2)
}

func (s EventStoreBlackboxTest) TestOptimisticLocking() {
	e := &TestEvent{Name: "Gabriel", Age: 24}
	s.buffer.Buffer(true, e)
	err := s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.buffer.CurrentVersion()),
		s.buffer.Events(context.Background())...,
	)
	s.NoError(err)

	// A concurrent append will now attempt to store
	s.buffer.ForceVersion(0)
	e = &TestEvent{Name: "Gabriel", Age: 27}
	s.buffer.Buffer(true, e)
	err = s.store.Append(
		context.Background(),
		bus.ExpectedVersion(s.buffer.CurrentVersion()),
		s.buffer.Events(context.Background())...,
	)
	s.Require().Error(err)
	s.EqualError(err, eventstore.ErrConcurrencyViolation.Error())
}

func (s EventStoreBlackboxTest) TestEnforcesSameStreamAppends() {
	e := &TestEvent{Name: "Gabriel", Age: 23}
	e.ForAggregate("lol")
	e.OwnedBy(uuid.New().String())
	e2 := &TestEvent{Name: "Giddian", Age: 99}
	e2.ForAggregate("yomp")
	e2.OwnedBy(uuid.New().String())

	err := s.store.Append(context.Background(), bus.ExpectedVersion(0), e, e2)
	s.Require().Error(err)
	s.EqualError(err, eventstore.ErrConsistencyViolation.Error())
}

func (s EventStoreBlackboxTest) TestSubscribesAll() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	for i := 0; i < 5; i++ {
		event := &TestEvent{Name: "Gabriel", Age: 25 + i}
		s.buffer.Buffer(true, event)
		err := s.store.Append(ctx, bus.ExpectedVersion(s.buffer.Version), s.buffer.Events(context.Background())...)
		s.buffer.Commit()
		s.Require().NoError(err)
	}

	var result []bus.Event
	err := s.store.Subscribe(ctx, func(e bus.Event) error {
		result = append(result, e)
		return nil
	})
	s.Require().NoError(err)

	s.Require().Len(result, 5)
	s.Equal("Gabriel", result[0].(*TestEvent).Name)
	s.Equal(25, result[0].(*TestEvent).Age)
	s.Equal(int64(1), result[0].Versioned())
}

func (s EventStoreBlackboxTest) TestNoEventsDoesntCallBack() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*20)
	defer cancel()

	err := s.store.Subscribe(ctx, func(e bus.Event) error {
		s.FailNow("Called back")
		cancel()
		return nil
	})
	s.Require().NoError(err)
}

func (s EventStoreBlackboxTest) TestSubscribesConcurrentlyOnceOnly() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	for i := 0; i < 100; i++ {
		event := &TestEvent{Name: "Gabriel", Age: 25 + i}
		s.buffer.Buffer(true, event)
		err := s.store.Append(ctx, bus.ExpectedVersion(s.buffer.Version), s.buffer.Events(context.Background())...)
		s.buffer.Commit()
		s.Require().NoError(err, "on append #%d", i)
	}

	start := make(chan struct{})
	results := make(chan bus.Event, 100)
	group, ctx := errgroup.WithContext(ctx)
	for i := 0; i < 25; i++ {
		group.Go(func() error {
			<-start
			return s.store.Subscribe(ctx, func(e bus.Event) error {
				results <- e
				return nil
			})
		})
	}

	close(start)
	err := group.Wait()
	s.Require().NoError(err)

	end := make([]bus.Event, 0)
	for event := range results {
		end = append(end, event)
		if len(results) == 0 {
			close(results)
			break
		}
	}

	ages := make(map[int]struct{})
	versions := make(map[int64]struct{})
	s.Require().Len(end, 100)
	for _, event := range end {
		if _, ok := ages[event.(*TestEvent).Age]; ok {
			s.FailNow("Age already delivered")
		}
		if _, ok := versions[event.Versioned()]; ok {
			s.FailNow("Version already delivered")
		}
		ages[event.(*TestEvent).Age] = struct{}{}
		versions[event.Versioned()] = struct{}{}
	}
}

func (s EventStoreBlackboxTest) TestSubscribeErrorNacks() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*70)
	defer cancel()

	for i := 0; i < 3; i++ {
		event := &TestEvent{Name: "Gabriel", Age: 25 + i}
		s.buffer.Buffer(true, event)
		err := s.store.Append(ctx, bus.ExpectedVersion(s.buffer.Version), s.buffer.Events(context.Background())...)
		s.buffer.Commit()
		s.Require().NoError(err, "on append #%d", i)
	}

	received := 0
	err := s.store.Subscribe(ctx, func(e bus.Event) error {
		s.Equal(25, e.(*TestEvent).Age)
		received++
		if received >= 3 {
			cancel()
		}
		return errors.New("test error")
	})
	s.Require().NoError(err)

	s.Require().GreaterOrEqual(received, 3)
}

func (s EventStoreBlackboxTest) TestConcurrentAppends() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	s.buffer.Buffer(true, &TestEvent{Name: "Gabriel"})
	event := s.buffer.Events(context.Background())[0]

	start := make(chan struct{})
	group, ctx := errgroup.WithContext(ctx)
	errors := 0
	for i := 0; i < 100; i++ {
		i := i
		group.Go(func() error {
			<-start
			err := s.store.Append(ctx, bus.ExpectedVersion(0), event)
			if err != nil {
				s.Require().EqualError(err, eventstore.ErrConcurrencyViolation.Error(), "Error'd on: %d", i)
				errors++
			}
			return nil
		})
	}

	close(start)
	group.Wait()

	s.Equal(99, errors)
}

func (s EventStoreBlackboxTest) TestSubscribesInOrder() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*2000)
	defer cancel()

	for i := 0; i < 500; i++ {
		s.buffer.Buffer(true, &TestEvent{Name: "Gabriel", Age: 25 + i})
		event := s.buffer.Events(ctx)[0]
		err := s.store.Append(ctx, bus.ExpectedVersion(s.buffer.Version), event)
		s.buffer.Commit()
		s.Require().NoError(err)
	}

	results := []int{}
	s.store.Subscribe(ctx, func(e bus.Event) error {
		age := e.(*TestEvent).Age
		results = append(results, age)
		if age == 524 {
			cancel()
		}
		return nil
	})

	starting := 25
	for _, age := range results {
		s.Require().Equal(starting, age)
		starting++
	}
}

func (s EventStoreBlackboxTest) TestCloseDoesntLoseEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*600)
	defer cancel()

	for i := 0; i < 3; i++ {
		s.buffer.Buffer(true, &TestEvent{Name: "Gabriel", Age: 25 + i})
		event := s.buffer.Events(ctx)[0]
		err := s.store.Append(ctx, bus.ExpectedVersion(s.buffer.Version), event)
		s.buffer.Commit()
		s.Require().NoError(err)
	}

	results := make(map[int]int)
	s.store.Subscribe(ctx, func(e bus.Event) error {
		s.store.Close()
		return nil
	})

	if opener, ok := s.store.(interface{ Open() }); ok {
		opener.Open()
	} else {
		s.store = s.factory()
	}
	err := s.store.Subscribe(ctx, func(e bus.Event) error {
		results[e.(*TestEvent).Age] += 1
		return nil
	})
	s.NoError(err)

	s.GreaterOrEqual(results[25], 1)
	s.GreaterOrEqual(results[26], 1)
	s.GreaterOrEqual(results[27], 1)
}

// Add another test to test postgres not acking, and timing out reserved
