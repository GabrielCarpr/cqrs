package message

// Type is a specific type of message,
// that allows the bus to discriminated with
// type a Message interface is
type Type string

var (
	// Command is a single request for the application to do something.
	// Commands may be used sync or async
	Command Type = "command"
	// Query is a question asked to the application, should never mutate state
	Query Type = "query"
	// Event is a record of something that has happened
	Event Type = "event"
	// QueuedEvent is an event that has been fanned out to a handler and has been
	// or is ready to be queued
	QueuedEvent Type = "queuedEvent"
)

// Message is a generic message that can be routed to an event or command handler
type Message interface {
	MessageType() Type
}
