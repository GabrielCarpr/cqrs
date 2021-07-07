package auth

import (
	"context"
	"github.com/GabrielCarpr/cqrs/bus"
)

// CommandAuthGuard is a command guard for ensuring an executor
// has authorisation to execute a command
func CommandAuthGuard(ctx context.Context, c bus.Command) (context.Context, bus.Command, error) {
	err := Enforce(ctx, c.Auth(ctx)...)
	return ctx, c, err
}

// QueryAuthGuard is a query guard for ensuring an executor
// can run a query
func QueryAuthGuard(ctx context.Context, q bus.Query) (context.Context, bus.Query, error) {
	err := Enforce(ctx, q.Auth(ctx)...)
	return ctx, q, err
}
