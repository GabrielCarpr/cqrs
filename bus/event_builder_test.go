package bus

import (
	"context"
	"testing"

	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/stretchr/testify/suite"
)

/*func Events(b bus.EventBuilder) {
	b.Event(&TestEvent{}).Handled(TestEventHandler{}, TestEventProcessManager{})
	b.Handlers(TestEventHandler{}, TestEventProcessManager{}).Listens(&TestEvent{}, &SomeOtherEvent{})

	b.Use(EventLoggingMiddleware)

	b.With(EventLoggingMiddleware).Sync().Event(&TestEvent{}).Handled(TestEventHandler{}, TestEventProcessManager{})

	b.Group(func(b bus.EventBuilder) {
		b.Use(EventLoggingMiddleware)
		b.Sync()

		b.Event(&TestEvent{}).Projected(TestEventHandler{})

		b.Projection(TestEventHandler{}).Listens(&TestEvent{})

		b.Event(&TestEvent{})
	})
}*/

type testEvent struct {
	EventType
}

func (testEvent) Event() string {
	return "test.event"
}

type otherTestEvent struct {
	EventType
}

func (otherTestEvent) Event() string {
	return "test.event.other"
}

type testEventHandler struct {
}

func (testEventHandler) Handle(context.Context, Event) ([]message.Message, error) {
	return []message.Message{}, nil
}

type otherTestEventHandler struct {
}

func (otherTestEventHandler) Handle(context.Context, Event) ([]message.Message, error) {
	return []message.Message{}, nil
}

func testEventMiddleware(h EventHandler) EventHandler {
	return h
}

func TestEventBuilder(t *testing.T) {
	suite.Run(t, new(EventBuilderSuite))
}

type EventBuilderSuite struct {
	suite.Suite

	b EventBuilder
	c *eventContext
}

func (s *EventBuilderSuite) SetupTest() {
	builder := &eventContext{}
	s.b = builder
	s.c = builder
}

func (s *EventBuilderSuite) TestRoutesEventToHandler() {
	s.b.Event(&testEvent{}).Handled(testEventHandler{})

	route := s.c.Route(&testEvent{})
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 1)
	s.Equal(route.handlers[0].handler, testEventHandler{})

	_ = s.c.Route(&otherTestEvent{})
}

func (s *EventBuilderSuite) TestRoutesEventsToHandlers() {
	s.b.Event(&testEvent{}, &otherTestEvent{}).Handled(testEventHandler{}, otherTestEventHandler{})
	s.b.Use(testEventMiddleware)

	route := s.c.Route(&otherTestEvent{})
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})
	s.Len(route.handlers[0].middleware, 1)
	s.Len(route.handlers[1].middleware, 1)

	route = s.c.Route(&testEvent{})
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})
	s.Len(route.handlers[0].middleware, 1)
	s.Len(route.handlers[1].middleware, 1)
}

func (s *EventBuilderSuite) TestRoutesEventToHandlers() {
	s.b.Event(&testEvent{}).Handled(testEventHandler{}, otherTestEventHandler{})

	route := s.c.Route(&testEvent{})
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
}

func (s *EventBuilderSuite) TestAdjacentRoutesIsolated() {
	s.b.Event(&testEvent{}).Handled(testEventHandler{})
	s.b.Event(&otherTestEvent{}).Handled(otherTestEventHandler{})

	route := s.c.Route(&testEvent{})
	s.Len(route.handlers, 1)

	route = s.c.Route(&otherTestEvent{})
	s.Len(route.handlers, 1)
}

func (s *EventBuilderSuite) TestHandlerListens() {
	s.b.Handler(
		testEventHandler{},
		otherTestEventHandler{},
	).Listens(&testEvent{}, &otherTestEvent{})

	route := s.c.Route(&testEvent{})
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})

	route = s.c.Route(&otherTestEvent{})
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})
}

func (s *EventBuilderSuite) TestWithMiddleware() {
	s.b.Event(&otherTestEvent{}).Handled(otherTestEventHandler{})
	s.b.With(testEventMiddleware).Event(&testEvent{}).Handled(testEventHandler{})
	s.b.Use(testEventMiddleware)

	route := s.c.Route(&testEvent{})
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 1)
	s.Len(route.handlers[0].middleware, 2)

	route = s.c.Route(&otherTestEvent{})
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 1)
	s.Len(route.handlers[0].middleware, 1)
}

func (s *EventBuilderSuite) TestGroupMiddleware() {
	s.b.Use(testEventMiddleware)
	s.b.Event(&otherTestEvent{}).Handled(testEventHandler{}, otherTestEventHandler{})
	s.b.Group(func(b EventBuilder) {
		b.With(testEventMiddleware).Handler(testEventHandler{}).Listens(&testEvent{})
		b.Use(testEventMiddleware)
	})
	s.b.Use(testEventMiddleware)

	route := s.c.Route(&otherTestEvent{})
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 2)
	s.Len(route.handlers[0].middleware, 2)

	route = s.c.Route(&testEvent{})
	s.Len(route.handlers, 1)
	s.Len(route.handlers[0].middleware, 4)
}
