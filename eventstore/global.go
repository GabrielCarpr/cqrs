package eventstore

import (
	"errors"

	"github.com/GabrielCarpr/cqrs/bus"
)

var (
	// ErrConcurrencyViolation indicates an optimistic locking failure
	ErrConcurrencyViolation = errors.New("cqrs.eventstore: concurrency violation")
	// ErrConsistencyViolation indicates appended events cross a consistency boundary
	ErrConsistencyViolation = errors.New("cqrs.eventstore: consistency violation")
)

// CheckEventsConsistent returns an error if the stream of
// events cross consistency boundaries
func CheckEventsConsistent(events ...bus.Event) error {
	sample := events[0]
	for _, event := range events {
		if event.FromAggregate() != sample.FromAggregate() {
			return ErrConsistencyViolation
		}
		if event.Owned() != sample.Owned() {
			return ErrConsistencyViolation
		}
	}
	return nil
}

// CheckExpectedVersion performs optimistic locking to prevent concurrent
// writes of the same stream
func CheckExpectedVersion(lastEvent bus.Event, v bus.ExpectedVersion) error {
	if v == bus.Any {
		return nil
	}
	if lastEvent != nil && lastEvent.Versioned() != int64(v) {
		return ErrConcurrencyViolation
	}
	if lastEvent == nil && v != bus.ExpectedVersion(0) {
		return ErrConcurrencyViolation
	}
	return nil
}
