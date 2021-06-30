package bus

import (
	"fmt"
	"reflect"
)

// CommandRules is a map that determines routing for a command.
// The key is the command name, and the value is the DI handler name.
// TODO: Change into routing composition
type CommandRules map[Command]string

type commandRules map[string]string

func (r commandRules) Merge(rules CommandRules) commandRules {
	for cmd, handler := range rules {
		r[string(cmd.Command())] = handler
	}
	return r
}

// EventRules is a map that determines routing for an event.
// The key is the event name, and the value is a list of DI handler names.
// TODO: Change into routing composition
type EventRules map[Event][]string

type eventRules map[string][]string

func (r eventRules) Merge(rules EventRules) eventRules {
	for event, handlers := range rules {
		existing, _ := r[string(event.Event())]
		r[string(event.Event())] = r.deduplicate(existing, handlers...)
	}
	return r
}

func (eventRules) deduplicate(existing []string, handlers ...string) []string {
	merged := make([]string, len(existing))
	copy(merged, existing)
one:
	for _, h := range handlers {
		for _, e := range existing {
			if e == h {
				break one
			}
		}
		merged = append(merged, h)
	}
	return merged
}

// QueryRules is a map of queries and query handlers
// TODO: Change into routing composition
type QueryRules map[Query]string

type queryRules map[string]string

func (r queryRules) Merge(rules QueryRules) queryRules {
	for query, handler := range rules {
		r[query.Query()] = handler
	}
	return r
}

// NewMessageRouter returns a new, empty, message router
func NewMessageRouter() MessageRouter {
	return MessageRouter{
		Events:   make(eventRules),
		Commands: make(commandRules),
		Queries:  make(queryRules),
	}
}

// MessageRouter routes a message to its correct destination.
// TODO: Somehow support command specific middleware - what use case?
// TODO: Change to composition based routing, rather than table based,
// which will allow grouping and middlewares
type MessageRouter struct {
	Events   eventRules
	Commands commandRules
	Queries  queryRules
}

// Extend takes EventRules|CommandRules|QueryRules and extends
// the routers internal routing rules with it
func (r *MessageRouter) Extend(rules interface{}) {
	switch v := rules.(type) {
	case EventRules:
		r.Events = r.Events.Merge(v)
		return
	case CommandRules:
		r.Commands = r.Commands.Merge(v)
		return
	case QueryRules:
		r.Queries = r.Queries.Merge(v)
		return
	}
	panic(fmt.Sprintf("Tried to extend MessageRouter with non-rules: %s", reflect.TypeOf(rules)))
}

// Route takes a command or event, and returns it's handlers
func (r MessageRouter) Route(m interface{}) []string {
	switch m := m.(type) {
	case Command:
		handler, ok := r.Commands[string(m.Command())]
		if !ok {
			return []string{}
		}
		return []string{handler}
	case Event:
		handlers, ok := r.Events[string(m.Event())]
		if !ok {
			return []string{}
		}
		return handlers
	case Query:
		handler, ok := r.Queries[string(m.Query())]
		if !ok {
			return []string{}
		}
		return []string{handler}
	default:
		panic(fmt.Sprintf("Tried to route non command or event: %s", reflect.TypeOf(m)))
	}
}
