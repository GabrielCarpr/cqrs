package bus

// Interfaces

type EventBuilder interface {
	eventBuilderRouteStarter
	eventBuilderMiddleware
}

type eventBuilderRouteStarter interface {
	eventBuilderHandler
	eventBuilderEvent
}

type eventBuilderEvent interface {
	Event(...Event) eventBuilderHandled
}

type eventBuilderHandled interface {
	Handled(...EventHandler)
}

type eventBuilderHandler interface {
	Handler(...EventHandler) eventBuilderListener
}

type eventBuilderListener interface {
	Listens(...Event)
}

type eventBuilderMiddleware interface {
	Use(...EventMiddleware)
	With(...EventMiddleware) eventBuilderRouteStarter
	Group(func(EventBuilder))
}

// Implementation

type eventRoute struct {
	event    Event
	handlers []eventHandlerRoute
}

func (e eventRoute) Merge(r eventRoute) eventRoute {
	if e.event.Event() != r.event.Event() {
		panic("tried to merge different events")
	}

	e.handlers = append(e.handlers, r.handlers...)
	return e
}

type eventHandlerRoute struct {
	handler    EventHandler
	middleware []EventMiddleware
}

var _ EventBuilder = (*eventContext)(nil)

func NewEventContext() *eventContext {
	return &eventContext{}
}

type eventContext struct {
	events     []Event
	handlers   []EventHandler
	middleware []EventMiddleware

	contexts []*eventContext
}

func (c *eventContext) Route(e Event) eventRoute {
	r := eventRoute{event: e}

	if c.handlesEvent(e.Event()) {
		for _, handler := range c.handlers {
			r.handlers = append(r.handlers, eventHandlerRoute{
				handler: handler,
			})
		}
	}

	for _, ctx := range c.contexts {
		r = r.Merge(ctx.Route(e))
	}

	for i := range r.handlers {
		r.handlers[i].middleware = append(r.handlers[i].middleware, c.middleware...)
	}

	return r
}

func (c *eventContext) HandlerRoute(e Event, h EventHandler) (eventHandlerRoute, bool) {
	route := c.Route(e)

	for _, handler := range route.handlers {
		if EventHandlerName(handler.handler) == EventHandlerName(h) {
			return handler, true
		}
	}

	return eventHandlerRoute{}, false
}

func (c *eventContext) Render() map[string]eventRoute {
	events := c.Events()
	result := make(map[string]eventRoute)

	for _, event := range events {
		result[event.Event()] = c.Route(event)
	}

	return result
}

func (c *eventContext) Events() []Event {
	events := make(map[string]Event, 0)
	for _, event := range c.events {
		events[event.Event()] = event
	}

	for _, ctx := range c.contexts {
		res := ctx.Events()
		for _, event := range res {
			events[event.Event()] = event
		}
	}

	result := make([]Event, 0)
	for _, event := range events {
		result = append(result, event)
	}

	return result
}

func (c *eventContext) handlesEvent(name string) bool {
	for _, event := range c.events {
		if event.Event() == name {
			return true
		}
	}
	return false
}

func (e *eventContext) Event(events ...Event) eventBuilderHandled {
	c := new(eventContext)
	c.events = events
	e.contexts = append(e.contexts, c)
	return c
}

func (e *eventContext) Handler(handlers ...EventHandler) eventBuilderListener {
	c := new(eventContext)
	c.handlers = handlers
	e.contexts = append(e.contexts, c)
	return c
}

func (e *eventContext) Handled(handlers ...EventHandler) {
	e.handlers = append(e.handlers, handlers...)
}

func (e *eventContext) Listens(events ...Event) {
	e.events = append(e.events, events...)
}

func (e *eventContext) Use(mw ...EventMiddleware) {
	e.middleware = append(e.middleware, mw...)
}

func (e *eventContext) With(mw ...EventMiddleware) eventBuilderRouteStarter {
	c := new(eventContext)
	c.middleware = mw
	e.contexts = append(e.contexts, c)
	return c
}

func (e *eventContext) Group(fn func(EventBuilder)) {
	c := new(eventContext)
	fn(c)
	e.contexts = append(e.contexts, c)
}
