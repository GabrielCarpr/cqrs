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

	route, ok := s.c.Route(&testEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 1)
	s.Equal(route.handlers[0].handler, testEventHandler{})

	_, ok = s.c.Route(&otherTestEvent{})
	s.False(ok)
}

func (s *EventBuilderSuite) TestRoutesEventsToHandlers() {
	s.b.Event(&testEvent{}, &otherTestEvent{}).Handled(testEventHandler{}, otherTestEventHandler{})

	route, ok := s.c.Route(&otherTestEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})

	route, ok = s.c.Route(&testEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})
}

func (s *EventBuilderSuite) TestRoutesEventToHandlers() {
	s.b.Event(&testEvent{}).Handled(testEventHandler{}, otherTestEventHandler{})

	route, ok := s.c.Route(&testEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
}

func (s *EventBuilderSuite) TestHandlerListens() {
	s.b.Handler(
		testEventHandler{},
		otherTestEventHandler{},
	).Listens(&testEvent{}, &otherTestEvent{})

	route, ok := s.c.Route(&testEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[0].handler, testEventHandler{})

	route, ok = s.c.Route(&otherTestEvent{})
	s.Require().True(ok)
	s.Equal(route.event.Event(), "test.event.other")
	s.Len(route.handlers, 2)
	s.Equal(route.handlers[1].handler, otherTestEventHandler{})
}
