package bus

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
	return Route{}, false
}

func (c *CommandContext) Command(cmd Command) Record {
	r := &RoutingRecord{Command: cmd}
	c.commands[cmd.Command()] = r
	return r
}

func (c *CommandContext) Use(middlewares ...CommandMiddleware) {
	c.middlewares = append(c.middlewares, middlewares...)
}

type CmdBuilder interface {
	Use(middlewares ...CommandMiddleware)

	//With(middlewares ...CommandMiddleware) CmdBuilder

	Command(Command) Record

	//Group(func(CmdBuilder))
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

func (r *RoutingRecord) Handled(h CommandHandler) {
	r.Handler = h
}

type Record interface {
	Handled(CommandHandler)
}
