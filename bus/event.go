package bus

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/google/uuid"
)

func init() {
	gob.Register(uuid.New())
}

type Metadata map[string]string

func (m Metadata) Merge(n Metadata) {
	for key, val := range n {
		m[key] = val
	}
}

// Event is a routable event indicating something has happened.
// Events are fanned out to both sync and async handlers
type Event interface {
	message.Message

	// Owned is the ID of the entity that owns/produced the event
	Owned() string

	// OwnedBy tells the event which entity the event originated from
	OwnedBy(string)

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

	WithMetadata(Metadata)

	HasMetadata() Metadata

	// Event returns the events name. Must be implemented by all events
	Event() string
}

// EventType is a struct designed to be embedded within an event,
// providing some basic behaviours
type EventType struct {
	Owner string `json:"owner"`

	At time.Time `json:"at"`

	Version int64 `json:"version"`

	Aggregate string `json:"aggregate"`

	Metadata Metadata `json:"metadata"`
}

// MessageType satisfies the message.Message interface, used for routing
func (e EventType) MessageType() message.Type {
	return message.Event
}

func (e *EventType) WithMetadata(m Metadata) {
	if e.Metadata == nil {
		e.Metadata = m
		return
	}

	e.Metadata.Merge(m)
}

func (e *EventType) HasMetadata() Metadata {
	return e.Metadata
}

// OwnedBy is the owning entity of the event
func (e *EventType) OwnedBy(id string) {
	e.Owner = id
}

// Owned implements the Event interface Owned
func (e EventType) Owned() string {
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
	Version int64  `json:"-"`
	Type    string `json:"-"`
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
		event.OwnedBy(e.owner.String())
		event.IsVersion(e.PendingVersion() + 1)
		event.PublishedAt(time.Now())
		event.ForAggregate(e.Type)
		e.events = append(e.events, event)
	}
}

// Messages returns the event queue as messages
func (e *EventBuffer) Messages(ctx context.Context) []message.Message {
	events := e.applyMetadata(ctx, e.events...)
	output := make([]message.Message, len(events))
	for i, event := range events {
		output[i] = event
	}
	return output
}

// Events empties the event queue, returning events
func (e *EventBuffer) Events(ctx context.Context) []Event {
	return e.applyMetadata(ctx, e.events...)
}

func (e *EventBuffer) applyMetadata(ctx context.Context, events ...Event) []Event {
	result := make([]Event, len(events))
	for i, ev := range events {
		m := Metadata(SerializeContext(ctx))
		ev.WithMetadata(m)
		result[i] = ev
	}
	return result
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

// Commit commits the new version to
// the entity. It is assumed that before calling commit, the entity
// has been persisted (with new version), and events published
func (e *EventBuffer) Commit() {
	e.Version = e.PendingVersion()
	e.Flush()
}

// ForceVersion forces the entity version, useful when not
// doing event sourcing, so the repository can set the stored
// entity version
func (e *EventBuffer) ForceVersion(v int64) {
	e.Version = v
}

func EventHandlerName(h EventHandler) string {
	t := reflect.TypeOf(h)
	return fmt.Sprint(t.PkgPath(), ".", t.Name())
}
