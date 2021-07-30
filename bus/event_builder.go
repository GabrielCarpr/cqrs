package bus

// Interfaces

type EventBuilder interface {
	eventBuilderRouteStarter
	//eventBuilderMiddleware
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
	Use(EventMiddleware)
	With(...EventMiddleware) eventBuilderRouteStarter
	Group(func(EventBuilder))
}

// Implementation

type eventRoute struct {
	event    Event
	handlers []eventHandlerRoute
}

type eventHandlerRoute struct {
	handler    EventHandler
	middleware []EventMiddleware
}

var _ EventBuilder = (*eventContext)(nil)

type eventContext struct {
	events   []Event
	handlers []EventHandler
}

func (c *eventContext) Route(e Event) (eventRoute, bool) {
	for _, event := range c.events {
		if event.Event() == e.Event() {
			var handlers []eventHandlerRoute
			for _, handler := range c.handlers {
				handlers = append(handlers, eventHandlerRoute{
					handler: handler,
				})
			}
			return eventRoute{
				event:    event,
				handlers: handlers,
			}, true
		}
	}
	return eventRoute{}, false
}

func (e *eventContext) Event(events ...Event) eventBuilderHandled {
	e.events = append(e.events, events...)
	return e
}

func (e *eventContext) Handler(handlers ...EventHandler) eventBuilderListener {
	e.handlers = append(e.handlers, handlers...)
	return e
}

func (e *eventContext) Handled(handlers ...EventHandler) {
	e.Handler(handlers...)
}

func (e *eventContext) Listens(events ...Event) {
	e.Event(events...)
}
