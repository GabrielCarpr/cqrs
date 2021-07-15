package rest

import (
    "github.com/GabrielCarpr/cqrs/bus"
    "github.com/GabrielCarpr/cqrs/ports/rest"
)

//go:generate go run github.com/GabrielCarpr/cqrs/gen gen rest routes.yml

func Rest(b *bus.Bus) *rest.Server {
    return New(b)
}
