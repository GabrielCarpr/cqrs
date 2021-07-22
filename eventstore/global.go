package eventstore

import "errors"

var (
	ErrConcurrencyViolation = errors.New("cqrs.eventstore: concurrency violation")
	ErrConsistencyViolation = errors.New("cqrs.eventstore: consistency violation")
)
