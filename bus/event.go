package bus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GabrielCarpr/cqrs/bus/message"
)

// Event is a routable event indicating something has happened.
// Events are fanned out to both sync and async handlers
type Event interface {
	message.Message

	// OwnedBy tells the event which entity the event originated from
	OwnedBy(fmt.Stringer)

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
func (e *EventType) OwnedBy(id fmt.Stringer) {
	e.Owner = id.String()
}

// EventHandler is a handler for one specific event.
// Each event may have multiple, or 0, EventHandlers.
type EventHandler interface {
	Handle(context.Context, Event) ([]message.Message, error)
	Async() bool
}

// Versionable is an entity that can be versioned with events
type Versionable interface {
	CurrentVersion() int64
	PendingVersion() int64
	Commit() []message.Message
	ApplyChange(bool, ...Event)
}

// Event Queue

// NewEventBuffer returns an owned event queue
func NewEventBuffer(owner fmt.Stringer) EventBuffer {
	return EventBuffer{
		owner: owner,
	}
}

// EventBuffer is embedded in entities to buffer events before
// being released to infrastructure
type EventBuffer struct {
	owner   interface{ String() string }
	events  []Event
	Version int64 `json:"version"`
}

// JSONMarshal implements encoding/json.Marshaler
func (e EventBuffer) JSONMarshal() ([]byte, error) {
	return json.Marshal(e.Version)
}

// Buffer adds events to the buffer queue,
// and sets their owner simutaneously
func (e *EventBuffer) Buffer(isNew bool, events ...Event) {
	for _, event := range events {
		if !isNew {
			e.Version++
			continue
		}
		event.OwnedBy(e.owner)
		e.events = append(e.events, event)
	}
}

// Messages returns the event queue as messages
func (e *EventBuffer) Messages() []message.Message {
	output := make([]message.Message, len(e.events))
	for i, event := range e.events {
		output[i] = event
	}
	return output
}

// Events empties the event queue, returning events
func (e *EventBuffer) Events() []Event {
	return e.events
}

// Flush clears the event queue, without committing
func (e *EventBuffer) Flush() {
	e.events = make([]Event, 0)
}

// CurrentVersion returns the entity's current version,
// the same as the attribute. Required for the interface
func (e *EventBuffer) CurrentVersion() int64 {
	return e.Version
}

// PendingVersion returns the version that the
// entity will get if committed
func (e *EventBuffer) PendingVersion() int64 {
	return e.Version + int64(len(e.events))
}

// Commit releases pending events, and commits the new version to
// the entity. It is assumed that after calling commit, the entity
// with be persisted (with new version), and events published
func (e *EventBuffer) Commit() []message.Message {
	output := e.Messages()
	e.Version = e.PendingVersion()
	e.Flush()
	return output
}

// ForceVersion forces the entity version, useful when not
// doing event sourcing, so the repository can set the stored
// entity version
func (e *EventBuffer) ForceVersion(v int64) {
	e.Version = v
}
