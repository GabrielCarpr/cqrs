package bus

import (
	"fmt"
	"reflect"
)

// CommandRules is a map that determines routing for a command.
// The key is the command name, and the value is the DI handler name.
type CommandRules map[Command]string

type commandRules map[string]string

func (r commandRules) Merge(rules CommandRules) commandRules {
	for cmd, handler := range rules {
		r[cmd.Command()] = handler
	}
	return r
}

// EventRules is a map that determines routing for an event.
// The key is the event name, and the value is a list of DI handler names.
type EventRules map[Event][]string

type eventRules map[string][]string

func (r eventRules) Merge(rules EventRules) eventRules {
	for event, handlers := range rules {
		existing, _ := r[event.Event()]
		r[event.Event()] = r.deduplicate(existing, handlers...)
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
type QueryRules map[Query]string

type queryRules map[string]string

func (r queryRules) Merge(rules QueryRules) queryRules {
	for query, handler := range rules {
		r[query.Query()] = handler
	}
	return r
}

func NewMessageRouter() MessageRouter {
	return MessageRouter{
		Events:   make(eventRules),
		Commands: make(commandRules),
		Queries:  make(queryRules),
	}
}

// MessageRouter routes a command to the correct destination
type MessageRouter struct {
	Events   eventRules
	Commands commandRules
	Queries  queryRules
}

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
		handler, ok := r.Commands[m.Command()]
		if !ok {
			return []string{}
		}
		return []string{handler}
	case Event:
		handlers, ok := r.Events[m.Event()]
		if !ok {
			return []string{}
		}
		return handlers
	case Query:
		handler, ok := r.Queries[m.Query()]
		if !ok {
			return []string{}
		}
		return []string{handler}
	default:
		panic(fmt.Sprintf("Tried to route non command or event: %s", reflect.TypeOf(m)))
	}
}
