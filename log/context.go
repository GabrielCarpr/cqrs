package log

import (
	"context"
	"github.com/google/uuid"
)

type ctxIDKeyType string

func (k ctxIDKeyType) String() string {
	return string(k)
}

var CtxIDKey = ctxIDKeyType("ID")

// WithID adds a correlation ID to the ctx. If one already exists, it's a no-op
func WithID(ctx context.Context) context.Context {
	existing := GetID(ctx)
	if existing != uuid.Nil {
		return ctx
	}

	id := uuid.New()
	ctx = context.WithValue(ctx, CtxIDKey, id)
	return ctx
}

// GetID returns the UUID from the context, or the UUID if it doesn't exist
func GetID(ctx context.Context) uuid.UUID {
	id := ctx.Value(CtxIDKey)
	if id == nil {
		return uuid.Nil
	}
	return id.(uuid.UUID)
}
