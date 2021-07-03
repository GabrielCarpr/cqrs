package bus

import (
	"fmt"
)

/*
Command Context
*/

// NewCommandContext creates an initialized CommandContext
func NewCommandContext() *CommandContext {
	return &CommandContext{commands: commandRoutes{}}
}

// CommandContext is a context that command routes are built in
type CommandContext struct {
	middlewares []CommandMiddleware
	commands    commandRoutes
	contexts    []*CommandContext
}

// Route takes a command and returns it's execution route
func (c CommandContext) Route(cmd Command) (commandRoute, bool) {
	r, ok := c.commands[cmd.Command()]
	if ok {
		return commandRoute{
			Command:    r.Command,
			Handler:    r.Handler,
			Middleware: c.middlewares,
		}, true
	}

	for _, ctx := range c.contexts {
		r, ok := ctx.Route(cmd)
		if ok {
			r.Middleware = append(r.Middleware, c.middlewares...)
			return r, true
		}
	}
	return commandRoute{}, false
}

// Routes buildings a routing table, a mapping between routable message and
// route entry
func (c CommandContext) Routes() commandRouting {
	commands := c.flatten()
	result := make(commandRouting)

	for _, cmd := range commands {
		route, ok := c.Route(cmd)
		if !ok {
			panic(fmt.Sprint("Could not compute route ", cmd.Command()))
		}
		result[cmd.Command()] = route
	}

	return result
}

func (c CommandContext) flatten() []Command {
	cmds := make([]Command, 0)

	for _, r := range c.commands {
		cmds = append(cmds, r.Command)
	}

	for _, ctx := range c.contexts {
		cmds = append(cmds, ctx.flatten()...)
	}
	return cmds
}

func (c *CommandContext) Command(cmd Command) commandRecord {
	r := &commandRoutingRecord{Command: cmd}
	_, exists := c.commands[cmd.Command()]
	if exists {
		panic(fmt.Sprint("Cannot register command twice: ", cmd.Command()))
	}
	c.commands[cmd.Command()] = r
	return r
}

func (c *CommandContext) Use(middlewares ...CommandMiddleware) {
	c.middlewares = append(c.middlewares, middlewares...)
}

func (c *CommandContext) Group(fn func(CmdBuilder)) {
	subContext := NewCommandContext()
	fn(subContext)
	c.contexts = append(c.contexts, subContext)
}

func (c *CommandContext) With(middlewares ...CommandMiddleware) CmdReceiver {
	subContext := NewCommandContext()
	subContext.middlewares = middlewares
	c.contexts = append(c.contexts, subContext)
	return subContext
}

func (c CommandContext) SelfTest() error {
	_, err := c.detectMultipleCommands(map[string]struct{}{})
	if err != nil {
		return err
	}

	err = c.detectHandlerlessCommands()
	if err != nil {
		return err
	}

	return nil
}

func (c CommandContext) detectHandlerlessCommands() error {
	for _, r := range c.commands {
		if r.Handler == nil {
			return fmt.Errorf("Command %s has no handler", r.Command.Command())
		}
	}

	for _, ctx := range c.contexts {
		err := ctx.detectHandlerlessCommands()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c CommandContext) detectMultipleCommands(cmds map[string]struct{}) (map[string]struct{}, error) {
	for cmd := range c.commands {
		_, exists := cmds[cmd]
		if exists {
			return cmds, fmt.Errorf("Command registered twice: %s", cmd)
		}
		cmds[cmd] = struct{}{}
	}

	for _, ctx := range c.contexts {
		cmds, err := ctx.detectMultipleCommands(cmds)
		if err != nil {
			return cmds, err
		}
	}
	return cmds, nil
}

/*
Commands
*/

// CmdReceiver allows registration of a command and handler
type CmdReceiver interface {
	Command(Command) commandRecord
}

// CmdBuilder allows building of command routing patterns
type CmdBuilder interface {
	CmdReceiver

	Use(middlewares ...CommandMiddleware)

	With(middlewares ...CommandMiddleware) CmdReceiver

	Group(func(CmdBuilder))
}

type commandRoutingRecord struct {
	Command Command
	Handler CommandHandler
}

func (r *commandRoutingRecord) Handled(h CommandHandler) {
	r.Handler = h
}

type commandRoutes map[string]*commandRoutingRecord

type commandRoute struct {
	Command    Command
	Middleware []CommandMiddleware
	Handler    CommandHandler
}

type commandRouting map[string]commandRoute

type commandRecord interface {
	Handled(CommandHandler)
}

/*
Query Context
*/

// NewQueryContext creates an initialized QueryContext
func NewQueryContext() *QueryContext {
	return &QueryContext{queries: queryRoutes{}}
}

// QueryContext is a context that command routes are built in
type QueryContext struct {
	middlewares []QueryMiddleware
	queries     queryRoutes
	contexts    []*QueryContext
}

// Route takes a command and returns it's execution route
func (c QueryContext) Route(q Query) (queryRoute, bool) {
	r, ok := c.queries[q.Query()]
	if ok {
		return queryRoute{
			Query:      r.Query,
			Handler:    r.Handler,
			Middleware: c.middlewares,
		}, true
	}

	for _, ctx := range c.contexts {
		r, ok := ctx.Route(q)
		if ok {
			r.Middleware = append(r.Middleware, c.middlewares...)
			return r, true
		}
	}
	return queryRoute{}, false
}

// Routes buildings a routing table, a mapping between routable message and
// route entry
func (c QueryContext) Routes() queryRouting {
	queries := c.flatten()
	result := make(queryRouting)

	for _, q := range queries {
		route, ok := c.Route(q)
		if !ok {
			panic(fmt.Sprint("Could not compute route ", q.Query()))
		}
		result[q.Query()] = route
	}

	return result
}

func (c QueryContext) flatten() []Query {
	qs := make([]Query, 0)

	for _, r := range c.queries {
		qs = append(qs, r.Query)
	}

	for _, ctx := range c.contexts {
		qs = append(qs, ctx.flatten()...)
	}
	return qs
}

func (c *QueryContext) Query(q Query) queryRecord {
	r := &queryRoutingRecord{Query: q}
	_, exists := c.queries[q.Query()]
	if exists {
		panic(fmt.Sprint("Cannot register command twice: ", q.Query()))
	}
	c.queries[q.Query()] = r
	return r
}

func (c *QueryContext) Use(middlewares ...QueryMiddleware) {
	c.middlewares = append(c.middlewares, middlewares...)
}

func (c *QueryContext) Group(fn func(QueryBuilder)) {
	subContext := NewQueryContext()
	fn(subContext)
	c.contexts = append(c.contexts, subContext)
}

func (c *QueryContext) With(middlewares ...QueryMiddleware) QueryReceiver {
	subContext := NewQueryContext()
	subContext.middlewares = middlewares
	c.contexts = append(c.contexts, subContext)
	return subContext
}

func (c QueryContext) SelfTest() error {
	_, err := c.detectMultipleQueries(map[string]struct{}{})
	if err != nil {
		return err
	}

	err = c.detectHandlerlessQueries()
	if err != nil {
		return err
	}

	return nil
}

func (c QueryContext) detectHandlerlessQueries() error {
	for _, r := range c.queries {
		if r.Handler == nil {
			return fmt.Errorf("Command %s has no handler", r.Query.Query())
		}
	}

	for _, ctx := range c.contexts {
		err := ctx.detectHandlerlessQueries()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c QueryContext) detectMultipleQueries(qs map[string]struct{}) (map[string]struct{}, error) {
	for q := range c.queries {
		_, exists := qs[q]
		if exists {
			return qs, fmt.Errorf("Query registered twice: %s", q)
		}
		qs[q] = struct{}{}
	}

	for _, ctx := range c.contexts {
		qs, err := ctx.detectMultipleQueries(qs)
		if err != nil {
			return qs, err
		}
	}
	return qs, nil
}

/*
Query
*/

type QueryReceiver interface {
	Query(Query) queryRecord
}

type QueryBuilder interface {
	QueryReceiver

	Use(middlewares ...QueryMiddleware)

	With(middlewares ...QueryMiddleware) QueryReceiver

	Group(func(QueryBuilder))
}

type queryRoutingRecord struct {
	Query   Query
	Handler QueryHandler
}

func (r *queryRoutingRecord) Handled(h QueryHandler) {
	r.Handler = h
}

type queryRoutes map[string]*queryRoutingRecord

type queryRoute struct {
	Query      Query
	Middleware []QueryMiddleware
	Handler    QueryHandler
}

type queryRouting map[string]queryRoute

type queryRecord interface {
	Handled(QueryHandler)
}
