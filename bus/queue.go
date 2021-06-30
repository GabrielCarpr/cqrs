package bus

import (
	"context"
	"github.com/gabrielcarpr/cqrs/bus/message"
)

type Queue interface {
	// RegisterCtxKey allows serialization of contexts
	RegisterCtxKey(key interface{ String() string }, fn func([]byte) interface{})

	// Publish publishes a message to the queue
	Publish(context.Context, ...message.Message) error

	// Subscribe registers a callback for inbound messages
	Subscribe(context.Context, func(context.Context, message.Message) error)

	// Close closes the queue down
	Close()
}
