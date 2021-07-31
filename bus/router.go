package bus

// NewMessageRouter returns a new, empty, message router
func NewMessageRouter() MessageRouter {
	return MessageRouter{
		events:   *NewEventContext(),
		commands: *NewCommandContext(),
		queries:  *NewQueryContext(),
	}
}

// MessageRouter routes a message to its correct destination.
type MessageRouter struct {
	events        eventContext
	commands      CommandContext
	queries       QueryContext
	eventRoutes   map[string]eventRoute
	commandRoutes commandRouting
	queryRoutes   queryRouting
}

// Extend takes EventRules|CommandRules|QueryRules and extends
// the routers internal routing rules with it
func (r *MessageRouter) ExtendEvents(fn func(b EventBuilder)) {
	r.events.Group(fn)
	r.eventRoutes = r.events.Render()
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
func (r MessageRouter) RouteEvent(e Event) eventRoute {
	route, ok := r.eventRoutes[e.Event()]
	if !ok {
		return r.events.Route(e)
	}
	return route
}

func (r MessageRouter) EventHandlerRoute(e Event, h EventHandler) (eventHandlerRoute, bool) {
	return r.events.HandlerRoute(e, h)
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
