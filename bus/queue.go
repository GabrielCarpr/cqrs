package bus

import (
	"context"
	"github.com/GabrielCarpr/cqrs/bus/message"
)

// Queue allows the bus to queue messages for asynchronous execution
type Queue interface {
	// RegisterCtxKey allows serialization of contexts by registering a key and context
	RegisterCtxKey(key interface{ String() string }, fn func([]byte) interface{})

	// Publish publishes a message to the queue
	// blocking until the message has been published
	Publish(context.Context, ...message.Message) error

	// Subscribe registers a callback for inbound messages
	// and runs the queue, blocking
	// TODO: Assess whether this should block
	Subscribe(context.Context, func(context.Context, message.Message) error)

	// Close closes the queue down
	// TODO: Assess necessity and find better way of managing queue lifecycle
	Close()
}
