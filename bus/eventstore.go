package bus

import (
	"context"
)

type ExpectedVersion int64

const NoExpectedVersion ExpectedVersion = -1

type Stream = chan<- Event

type EventStore interface {
	Appendable
	Streamable
	Subscribable
}

type Appendable interface {
	Append(context.Context, ExpectedVersion, ...Event)
}

type Streamable interface {
	Stream(context.Context, Stream, Select) error
}

type Subscribable interface {
	Subscribe(context.Context, Stream) error
}

type Select struct {
	StreamID
	From int64
}

type StreamID struct {
	Type string
	ID   string
}
