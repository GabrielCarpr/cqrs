package bus

import (
	// Bus package - a CQRS message bus
	_ "github.com/gabrielcarpr/cqrs/bus"
	// Background package - extension of bus for running long running processes in the background
	_ "github.com/gabrielcarpr/cqrs/background"
	// Log package - a basic global logger
	_ "github.com/gabrielcarpr/cqrs/log"
)
