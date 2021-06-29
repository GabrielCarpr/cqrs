package message

// Message is a generic message that can be routed to an event or command handler
type Message interface {
	MessageType() string
}
