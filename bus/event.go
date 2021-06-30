package bus

import (
	"context"
	"github.com/gabrielcarpr/cqrs/bus/message"
)

// Event is a routable event
type Event interface {
	message.Message
	OwnedBy(interface{ String() string })
	Event() string
}

// EventType is a record of something that has happened.
// Once routed, the event is fanned out to multiple
// handlers
type EventType struct {
	Owner string
}

func (e EventType) MessageType() string {
	return "event"
}

// OwnedBy is the owning entity of the event
func (e *EventType) OwnedBy(id interface{ String() string }) {
	e.Owner = id.String()
}

// EventHandler is a handler for one specific event.
// Each event may have multiple, or 0, EventHandlers.
type EventHandler interface {
	Handle(context.Context, Event) ([]message.Message, error)
	Async() bool
}

func NewEventQueue(owner interface{ String() string }) EventQueue {
	return EventQueue{
		owner: owner,
	}
}

// Event Queue

type EventQueue struct {
	owner     interface{ String() string }
	events    []message.Message
	GobEncode bool // Unused, purely to make Gob encode the eventqueue and not fail.
}

func (e *EventQueue) Publish(events ...Event) {
	for _, event := range events {
		event.OwnedBy(e.owner)
		e.events = append(e.events, event)
	}
}

func (e *EventQueue) Release() []message.Message {
	output := make([]message.Message, len(e.events))
	for i, event := range e.events {
		output[i] = event
	}
	e.events = make([]message.Message, 0)
	return output
}
