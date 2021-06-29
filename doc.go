package bus

import (
	// Bus package - a CQRS message bus
	_ "cqrs/bus"
	// Background package - extension of bus for running long running processes in the background
	_ "cqrs/background"
	// Log package - a basic global logger
	_ "cqrs/log"
)
