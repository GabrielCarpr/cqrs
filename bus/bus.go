package bus

import (
	"context"
	"encoding/gob"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"reflect"

	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/log"

	"github.com/sarulabs/di/v2"
)

var Instance *Bus

// NewBus returns a new configured bus.
func NewBus(ctx context.Context, bcs []Module, configs ...Config) *Bus {
	if Instance != nil {
		return Instance
	}

	builder, _ := di.NewBuilder()
	for _, bc := range bcs {
		for _, def := range bc.Services() {
			builder.Add(def.diDef())
		}
	}
	c := builder.Build()
	gob.Register(queuedEvent{})

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	b := &Bus{
		routes:    NewMessageRouter(),
		Container: c,
		queue:     nil,
		ctx:       ctx,
		ctxCancel: cancel,
		workers:   make([]func(), 0),
		deletions: make([]func(), 0),
	}

	for _, conf := range configs {
		err := conf(b)
		if err != nil {
			panic(err)
		}
	}

	for _, bc := range bcs {
		b.ExtendEvents(bc.EventRules())
		b.routes.ExtendCommands(bc.Commands)
		b.routes.ExtendQueries(bc.Queries)
	}

	b.Use(
		b.queryContainerMiddleware,
		b.commandContainerMiddleware,
		b.queryContainerGuard,
		b.commandContainerGuard,
	)
	Instance = b
	return b
}

// Bus is the main dependency. It is the entry point for all
// messages and routes them to the correct place either synchronously
// or asynchronously
type Bus struct {
	routes    MessageRouter
	Container di.Container
	queue     Queue
	ctx       context.Context
	ctxCancel context.CancelFunc

	workers   []func()
	deletions []func()

	commandGuards     []CommandGuard
	queryGuards       []QueryGuard
	commandMiddleware []CommandMiddleware
	queryMiddleware   []QueryMiddleware
}

// Close deletes all the container resources.
func (b *Bus) Close() {
	log.Info(b.ctx, "Closing bus", log.F{})
	defer b.ctxCancel()
	for _, del := range b.deletions {
		del()
	}
	if b.queue != nil {
		b.queue.Close()
	}
	b.Container.Delete()
	Instance = nil
}

// Work runs the bus in subscribe mode, to be ran as on a worker
// node, or in the background on an API server
func (b *Bus) Work() {
	for _, work := range b.workers {
		work()
	}

	// The queue blocks. The ctx will signal cancellation, and then it will unblock
	b.queue.Subscribe(b.ctx, func(ctx context.Context, msg message.Message) error {
		return b.routeFromQueue(ctx, msg)
	})

	b.Close()
}

func (b *Bus) Get(key string) interface{} {
	return b.Container.UnscopedGet(key)
}

// RegisterDeletion allows a plugin to register a function to
// clean itself up.
func (b *Bus) RegisterDeletion(fn func()) {
	b.deletions = append(b.deletions, fn)
}

// RegisterWork allows plugins to register a function for themselves
// that the bus should call when in worker mode
func (b *Bus) RegisterWork(fn func()) {
	b.workers = append(b.workers, fn)
}

// ExtendEvents extends the Bus EventRules
func (b *Bus) ExtendEvents(rules ...EventRules) *Bus {
	for _, rule := range rules {
		b.routes.Extend(rule)
		for event := range rule {
			stdlog.Printf("Registered event with gob: %s", event.Event())
			gob.Register(event)
			gob.Register(&event)
		}
	}
	return b
}

func (b *Bus) ExtendCommands(fn func(CmdBuilder)) {
	b.routes.ExtendCommands(fn)
}

func (b *Bus) ExtendQueries(fn func(QueryBuilder)) {
	b.routes.ExtendQueries(fn)
}

// RegisterContextKey registers a context key interpretation value for serialization
func (b *Bus) RegisterContextKey(key interface{ String() string }, fn func(j []byte) interface{}) {
	b.queue.RegisterCtxKey(key, fn)
}

// Use registers middleware and guards. Accepts a union of command/query guards and middleware.
func (b *Bus) Use(ms ...interface{}) {
	for _, m := range ms {
		switch v := m.(type) {
		case CommandGuard:
			b.commandGuards = append(b.commandGuards, v)
			break
		case QueryGuard:
			b.queryGuards = append(b.queryGuards, v)
			break
		case CommandMiddleware:
			b.commandMiddleware = append(b.commandMiddleware, v)
			break
		case QueryMiddleware:
			b.queryMiddleware = append(b.queryMiddleware, v)
			break
		default:
			panic(fmt.Sprint("Not a valid middleware, is ", reflect.TypeOf(v)))
		}
	}
}

func (b *Bus) routeFromQueue(ctx context.Context, msg message.Message) error {
	var err error
	var msgs []message.Message
	switch v := msg.(type) {
	case Command:
		_, err = b.Dispatch(ctx, v, true)
		break
	case queuedEvent:
		msgs, err = b.handleEvent(ctx, v, false)
		break
	}
	if err != nil {
		return err
	}
	err = b.route(ctx, msgs...)
	if err != nil {
		return err
	}
	return nil
}

// Routes a group of events. Will always favour async. Designed
// to be used with the return of events and commands
func (b *Bus) route(ctx context.Context, messages ...message.Message) error {
	for _, message := range messages {
		var err error
		switch v := message.(type) {
		case Command:
			_, err = b.Dispatch(ctx, v, false)
			break
		case Event:
			err = b.Publish(ctx, v)
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Dispatch runs a command, either synchronously or asynchronously
func (b *Bus) Dispatch(ctx context.Context, cmd Command, sync bool) (*CommandResponse, error) {
	ctx, cmd, err := b.runCmdGuards(ctx, cmd)
	if err != nil {
		return &CommandResponse{Error: err}, err
	}

	route, ok := b.routes.RouteCommand(cmd)
	if !ok {
		return &CommandResponse{}, NoCommandHandler{cmd}
	}
	handlerName := CommandHandlerName(route.Handler)

	if !sync {
		log.Info(ctx, "Publishing command", log.F{"command": cmd.Command()})
		err = b.queue.Publish(ctx, cmd)
		return nil, err
	}

	handler := Get(ctx, handlerName).(CommandHandler)
	for _, mw := range b.commandMiddleware {
		handler = mw(handler)
	}

	log.Info(ctx, "Routed command", log.F{"command": cmd.Command(), "handler": handlerName})

	response, messages := handler.Execute(ctx, cmd)
	err = b.route(ctx, messages...)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (b *Bus) runCmdGuards(ctx context.Context, cmd Command) (context.Context, Command, error) {
	var err error
	for _, guard := range b.commandGuards {
		ctx, cmd, err = guard(ctx, cmd)
		if err != nil {
			return ctx, cmd, err
		}
	}
	return ctx, cmd, err
}

// Publish distributes one or more events to the system
func (b *Bus) Publish(ctx context.Context, events ...Event) error {
	var queueables []queuedEvent
	for _, event := range events {
		handlerNames := b.routes.RouteEvent(event)
		for _, name := range handlerNames {
			queueables = append(queueables, queuedEvent{event, name})
		}
	}

	for _, queueable := range queueables {
		messages, err := b.handleEvent(ctx, queueable, true)
		if err != nil {
			return err
		}
		err = b.route(ctx, messages...)
		if err != nil {
			return err
		}
	}

	return nil
}

// Query routes and handles a query
func (b *Bus) Query(ctx context.Context, query Query, result interface{}) error {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidQueryResult
	}

	ctx, query, err := b.runQueryGuards(ctx, query)
	if err != nil {
		return err
	}

	route, exists := b.routes.RouteQuery(query)
	if !exists {
		return NoQueryHandler{query}
	}
	handlerName := QueryHandlerName(route.Handler)

	handler := Get(ctx, handlerName).(QueryHandler)

	for _, mw := range b.queryMiddleware {
		handler = mw(handler)
	}
	return handler.Execute(ctx, query, result)
}

func (b *Bus) runQueryGuards(ctx context.Context, q Query) (context.Context, Query, error) {
	var err error
	for _, guard := range b.queryGuards {
		ctx, q, err = guard(ctx, q)
		if err != nil {
			return ctx, q, err
		}
	}
	return ctx, q, err
}

// handleEvent handles a queued event
func (b *Bus) handleEvent(ctx context.Context, e queuedEvent, allowAsync bool) ([]message.Message, error) {
	handler := b.Container.Get(e.Handler).(EventHandler)
	if allowAsync && handler.Async() {
		log.Info(ctx, "Queuing event", log.F{"event": e.Event.Event(), "handler": reflect.TypeOf(e.Handler).String()})
		err := b.queue.Publish(ctx, e)
		return []message.Message{}, err
	}
	log.Info(ctx, "Handling event", log.F{"event": e.Event.Event(), "handler": reflect.TypeOf(handler).String()})
	return handler.Handle(ctx, e.Event)
}

type queuedEvent struct {
	Event   Event
	Handler string
}

func (queuedEvent) MessageType() message.Type {
	return message.QueuedEvent
}
