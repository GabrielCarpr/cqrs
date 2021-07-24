package bus

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/google/uuid"
)

func init() {
	gob.Register(uuid.New())
}

type Metadata map[string]string

// Event is a routable event indicating something has happened.
// Events are fanned out to both sync and async handlers
type Event interface {
	message.Message

	// Owned is the ID of the entity that owns/produced the event
	Owned() fmt.Stringer

	// OwnedBy tells the event which entity the event originated from
	OwnedBy(fmt.Stringer)

	// ForAggregate is the type of entity that owns/produces the event
	ForAggregate(string)

	// FromAggregate is the type of entity that produced the event
	FromAggregate() string

	// PublishedAt sets the time the event was published
	PublishedAt(time.Time)

	// WasPublishedAt returns the time the event was published
	WasPublishedAt() time.Time

	// IsVersion is the version of the event on the entity
	IsVersion(int64)

	// Versioned returns the version of the entity the event
	Versioned() int64

	// Event returns the events name. Must be implemented by all events
	Event() string
}

// EventType is a struct designed to be embedded within an event,
// providing some basic behaviours
type EventType struct {
	Owner fmt.Stringer `json:"owner"`

	At time.Time `json:"at"`

	Version int64 `json:"version"`

	Aggregate string `json:"aggregate"`

	Metadata Metadata `json:"metadata"`
}

// MessageType satisfies the message.Message interface, used for routing
func (e EventType) MessageType() message.Type {
	return message.Event
}

// OwnedBy is the owning entity of the event
func (e *EventType) OwnedBy(id fmt.Stringer) {
	e.Owner = id
}

// Owned implements the Event interface Owned
func (e EventType) Owned() fmt.Stringer {
	return e.Owner
}

// PublishedAt implements Event interface PublishedAt
func (e *EventType) PublishedAt(t time.Time) {
	e.At = t
}

// WasPublishedAt implements Event interface WasPublishedAt
func (e EventType) WasPublishedAt() time.Time {
	return e.At
}

// IsVersion implements the Event interface IsVersion
func (e *EventType) IsVersion(v int64) {
	e.Version = v
}

// Versioned implements the Event interface Versioned
func (e EventType) Versioned() int64 {
	return e.Version
}

// ForAggregate implements the Event interface ForAggregate
func (e *EventType) ForAggregate(t string) {
	e.Aggregate = t
}

// FromAggregate implements the EventInterface FromAggregate
func (e EventType) FromAggregate() string {
	return e.Aggregate
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
func NewEventBuffer(owner fmt.Stringer, t string) EventBuffer {
	return EventBuffer{
		owner: owner,
		Type:  t,
	}
}

// EventBuffer is embedded in entities to buffer events before
// being released to infrastructure
type EventBuffer struct {
	owner   fmt.Stringer
	events  []Event
	Version int64 `json:"version"`
	Type    string
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
		event.IsVersion(e.PendingVersion() + 1)
		event.PublishedAt(time.Now())
		event.ForAggregate(e.Type)
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
func (e *EventBuffer) Commit() []Event {
	output := e.Events()
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
