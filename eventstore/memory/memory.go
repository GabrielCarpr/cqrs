package memory

import (
	"context"
	"sync"
	"time"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/eventstore"
)

type MemoryEventStore struct {
	events []bus.Event

	mx sync.Mutex

	offset int
}

func (s *MemoryEventStore) Append(ctx context.Context, v bus.ExpectedVersion, events ...bus.Event) error {
	sample := events[0]
	for _, event := range events {
		if event.FromAggregate() != sample.FromAggregate() {
			return eventstore.ErrConsistencyViolation
		}
		if event.Owned() != sample.Owned() {
			return eventstore.ErrConsistencyViolation
		}
	}

	last, ok := s.lastEventFor(bus.StreamID{Type: events[0].FromAggregate(), ID: events[0].Owned().String()})
	if ok && last.Versioned() != int64(v) {
		return eventstore.ErrConcurrencyViolation
	}
	if !ok && v != bus.ExpectedVersion(0) {
		return eventstore.ErrConcurrencyViolation
	}

	s.events = append(s.events, events...)
	return nil
}

func (s *MemoryEventStore) lastEventFor(id bus.StreamID) (bus.Event, bool) {
	for i := len(s.events) - 1; i >= 0; i-- {
		event := s.events[i]
		if id.Type != "" && event.FromAggregate() != id.Type {
			continue
		}
		if id.ID != "" && event.Owned().String() != id.ID {
			continue
		}
		return event, true
	}
	return nil, false
}

func (s *MemoryEventStore) Stream(ctx context.Context, stream bus.Stream, q bus.Select) error {
	for _, event := range s.events {
		if q.Type != "" && event.FromAggregate() != q.Type {
			continue
		}
		if q.ID != "" && event.Owned().String() != q.ID {
			continue
		}
		if event.Versioned() < q.From {
			continue
		}
		stream <- event
	}

	close(stream)
	return nil
}

func (s *MemoryEventStore) Subscribe(ctx context.Context, q bus.Select, subscription func(bus.Event) error) error {
	errors := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(time.Millisecond * 10):
			err := func() error {
				s.mx.Lock()
				defer s.mx.Unlock()
				if len(s.events) == s.offset {
					return nil // Up to date
				}
				res := subscription(s.events[s.offset])
				if res != nil {
					errors++
					if errors > 4 {
						return res // There were 5 errors and retrying didn't work
					}
					return nil // There was an error but we can retry
				}
				s.offset++
				return nil // Success
			}()
			if err != nil {
				return err // The subscription failed
			}
		}
	}
}
