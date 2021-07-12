package bus

import "context"

// Plugin is an extension of the bus
type Plugin interface {
	// Close shuts the plugin down
	Close() error

	// Work runs the worker in a goroutine, blocking
	Work(context.Context) error

	// Middleware returns any injected middlewares from the plugin
	Middleware() []interface{}

	// Attach gives the plugin a way to dispatch commands
	Attach(func(context.Context, Command) error)
}
