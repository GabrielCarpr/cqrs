package bus

import "context"

// Plugin is an extension of the bus
type Plugin interface {
	// Close shuts the plugin down
	Close() error

	// Register registers the bus with a plugin
	Register(*Bus) error

	// Run runs the plugin, blocking
	Run(context.Context) error
}
