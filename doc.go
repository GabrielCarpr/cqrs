package bus

import (
	// Bus package - a CQRS message bus
	_ "github.com/GabrielCarpr/cqrs/bus"
	// Background package - extension of bus for running long running processes in the background
	_ "github.com/GabrielCarpr/cqrs/background"
	// Log package - a basic global logger
	_ "github.com/GabrielCarpr/cqrs/log"
	// Auth package - access control and authorization adapters
	_ "github.com/GabrielCarpr/cqrs/auth"
	// Error package - error types for use library wide
	_ "github.com/GabrielCarpr/cqrs/errors"
)
