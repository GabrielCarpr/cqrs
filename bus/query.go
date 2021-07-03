package bus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/GabrielCarpr/cqrs/bus/message"
)

// Query is a question that is asked of the application.
// Execution of the query cannot change application state,
// although may still change infrastructure state (such as monitoring)
type Query interface {
	message.Message

	// Query returns the name of the query, and must be implemented
	// by every query
	Query() string

	// Valid returns an error if the query is invalid
	Valid() error

	// Auth returns the scopes required to execute the query.
	// May return dynamic scopes, based on values in the context
	Auth(context.Context) [][]string
}

// QueryType is a utility type that can be embedded within a new Query
type QueryType struct {
}

// MessageType returns the type for use with routing within the bus
func (QueryType) MessageType() message.Type {
	return message.Query
}

// Auth provides a public default to satisfy the query interface
func (QueryType) Auth(context.Context) [][]string {
	return [][]string{}
}

// QueryHandler is an interface for a handler that
// executes a query. Queries have a 1:1 relationship with handlers.
type QueryHandler interface {
	// Execute runs the query, and fills a result provided
	// in the third argument, which must be a pointer.
	Execute(context.Context, Query, interface{}) error
}

// QueryHandlerName returns the name of the handler, used for DI
// and routing
func QueryHandlerName(h QueryHandler) string {
	t := reflect.TypeOf(h)
	return fmt.Sprint(t.PkgPath(), ".", t.Name())
}
