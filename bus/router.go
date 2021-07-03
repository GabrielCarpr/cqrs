package bus

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

// NewMessageRouter returns a new, empty, message router
func NewMessageRouter() MessageRouter {
	return MessageRouter{
		Events:   make(eventRules),
		commands: *NewCommandContext(),
		queries:  *NewQueryContext(),
	}
}

// MessageRouter routes a message to its correct destination.
// TODO: Somehow support command specific middleware - what use case?
// TODO: Change to composition based routing, rather than table based,
// which will allow grouping and middlewares
// TODO: Auto generate documentation of bus
type MessageRouter struct {
	Events        eventRules
	commands      CommandContext
	queries       QueryContext
	commandRoutes commandRouting
	queryRoutes   queryRouting
}

// Extend takes EventRules|CommandRules|QueryRules and extends
// the routers internal routing rules with it
func (r *MessageRouter) Extend(rules EventRules) {
	r.Events = r.Events.Merge(rules)
}

func (r *MessageRouter) ExtendCommands(fn func(b CmdBuilder)) {
	r.commands.Group(fn)
	r.commandRoutes = r.commands.Routes()
}

func (r *MessageRouter) ExtendQueries(fn func(b QueryBuilder)) {
	r.queries.Group(fn)
	r.queryRoutes = r.queries.Routes()
}

// RouteEvent returns all the handlers for an event
func (r MessageRouter) RouteEvent(e Event) []string {
	handlers, ok := r.Events[string(e.Event())]
	if !ok {
		return []string{}
	}
	return handlers
}

// RouteCommand returns the routing record for a command, if it exists
func (r MessageRouter) RouteCommand(cmd Command) (CommandRoute, bool) {
	route, ok := r.commandRoutes[cmd.Command()]
	return route, ok
}

// RouteQuery returns the routing record for a query, if it exists
func (r MessageRouter) RouteQuery(q Query) (QueryRoute, bool) {
	route, ok := r.queryRoutes[q.Query()]
	return route, ok
}
