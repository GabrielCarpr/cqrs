package bus

import (
	"context"
	"cqrs/bus/message"
)

// Query is a query object
type Query interface {
	message.Message
	Query() string
	Valid() error
	Auth(context.Context) [][]string
}

type QueryType struct {
}

func (QueryType) MessageType() string {
	return "query"
}

func (QueryType) Auth(context.Context) [][]string {
	return [][]string{}
}

type QueryHandler interface {
	Execute(context.Context, Query, interface{}) error
}
