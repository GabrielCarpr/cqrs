package bus

import (
	"fmt"
)

// Example module config - v1

/*type Module struct {
}

func (Module) Commands(b CmdBuilder) {
	b.Command(CreateUser{}).Handler(b.Name(CreateUserHandler{}))

	b.CMiddleware(func(b CmdBuilder) {
		b.Command(UpdateUser{}).Handler(b.Name(UpdateUserHandler{}))
		b.Command(RegisterUser{}).Handler(b.Name(RegisterUserHandler{}))
	})
}

func (Module) Queries(b QueryBuilder) {
	b.Query(TopUsers{}).Handler(b.Name(CreateUserHandler{}))

	b.QGroup(bus.Group{Middleware: []bus.Middleware{AccessControl{}}}, func(b QueryBuilder) {
		b.Query(LastActiveUsers{}).Handler(b.Name(LastActiveUsersHandler{}))
	})
}

func (Module) Events(b EventBuilder) {
	b.Event(Registration{}).Subscribed(
		RegistrationEmail{},
		MarketingAddition{},
	)

	b.EGroup(bus.Group{
		Middleware: []b.Middleware{bus.Logging},
	}, func(b EventBuilder) {
		b.Event(Registration{}).Subscribed(
			MarketingAddition{},
		)
	})
}

// Example routing config - v2

type Module2 struct {
}

func (Module2) Commands(b CmdBuilder) {
	b.Use(bus.LoggingMiddleware, bus.RecoveryMiddleware)

	b.With(bus.LoggingMiddleware).Command(Register{}).Handled(RegisterHandler{})
	b.With(
		bus.LoggingMiddleware,
		bus.RetryMiddleware(5, 30),
	).Command(SendEmail{}).Handled(SendEmailHandler{})

	b.Group(bus.Group{
		Middleware: b.Middleware(bus.LoggingMiddleware),
	}, func(b CmdBuilder) {
		b.Command(SendEmail{}).Handled(SendEmailHandler{})
	})
}

func (Module2) Queries(b QueryBuilder) {
	b.Use(bus.LoggingMiddleware, bus.Recovery)

	b.With(bus.LoggingMiddleware).Command(Users{}).Handled(UsersHandler{})
}*/

func NewCommandContext() *CommandContext {
	return &CommandContext{commands: Routes{}}
}

type CommandContext struct {
	middlewares []CommandMiddleware
	commands    Routes
	contexts    []*CommandContext
}

func (c CommandContext) Route(cmd Command) (Route, bool) {
	r, ok := c.commands[cmd.Command()]
	if ok {
		return Route{
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
	return Route{}, false
}

func (c CommandContext) Routes() Routing {
	commands := c.flatten()
	result := make(Routing)

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

func (c *CommandContext) Command(cmd Command) Record {
	r := &RoutingRecord{Command: cmd}
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

func (c *CommandContext) With(middlewares ...CommandMiddleware) CmdRegister {
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

type CmdRegister interface {
	Command(Command) Record
}

type CmdBuilder interface {
	CmdRegister

	Use(middlewares ...CommandMiddleware)

	With(middlewares ...CommandMiddleware) CmdRegister

	Group(func(CmdBuilder))
}

type RoutingRecord struct {
	Command Command
	Handler CommandHandler
}

type Routes map[string]*RoutingRecord

type Route struct {
	Command    Command
	Middleware []CommandMiddleware
	Handler    CommandHandler
}

type Routing map[string]Route

func (r Routing) Merge(new Routing) {
	for cmd, record := range new {
		r[cmd] = record
	}
}

func (r *RoutingRecord) Handled(h CommandHandler) {
	r.Handler = h
}

type Record interface {
	Handled(CommandHandler)
}
