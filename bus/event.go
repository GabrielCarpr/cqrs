package bus

import (
	"context"
	"github.com/GabrielCarpr/cqrs/bus/message"
)

// Event is a routable event indicating something has happened.
// Events are fanned out to both sync and async handlers
type Event interface {
	message.Message

	// OwnedBy tells the event which entity the event originated from
	OwnedBy(interface{ String() string })

	// Event returns the events name. Must be implemented by all events
	Event() string
}

// EventType is a struct designed to be embedded within an event,
// providing some basic behaviours
type EventType struct {
	Owner string
}

// MessageType satisfies the message.Message interface, used for routing
func (e EventType) MessageType() message.Type {
	return message.Event
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

// NewEventQueue returns an owned event queue
func NewEventQueue(owner interface{ String() string }) EventQueue {
	return EventQueue{
		owner: owner,
	}
}

// Event Queue

// EventQueue is embedded in entities to buffer events before
// being released to infrastructure
type EventQueue struct {
	owner     interface{ String() string }
	events    []Event
	GobEncode bool // Unused, purely to make Gob encode the eventqueue and not fail.
}

// Publish adds events to the buffer queue,
// and sets their owner simutaneously
func (e *EventQueue) Publish(events ...Event) {
	for _, event := range events {
		event.OwnedBy(e.owner)
		e.events = append(e.events, event)
	}
}

// Release empties the event queue, returning
func (e *EventQueue) Release() []message.Message {
	output := make([]message.Message, len(e.events))
	for i, event := range e.events {
		output[i] = event
	}
	e.events = make([]Event, 0)
	return output
}

// ReleaseEvents empties the event queue, returning events
func (e *EventQueue) ReleaseEvents() []Event {
	output := make([]Event, len(e.events))
	for i, event := range e.events {
		output[i] = event
	}
	e.events = make([]Event, 0)
	return output
}
