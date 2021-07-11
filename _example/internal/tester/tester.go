package tester

import (
	"example/internal/app"
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/bus"

	"github.com/golang-migrate/migrate/v4"
	"context"
)

var (
	doubles = []bus.Def{}
)

func testModule(defs []bus.Def) bus.Module {
	return bus.FuncModule{
		ServicesFunc: func() []bus.Def{
			return append(doubles, defs...)
		},
	}
}

type Integration struct {
	bus *bus.Bus
	Doubles []bus.Def
}

func (i *Integration) SetupTest() {
	config.Values = GetTestConfig()
	i.bus = bus.Default(context.Background(), append(app.Modules, testModule(i.Doubles)))
	i.migrate()
}

func (i *Integration) TearDownTest() {
	i.bus.Close()
	i.Doubles = []bus.Def{}
}

func (i *Integration) Bus() *bus.Bus {
	return i.bus
}

func (i *Integration) Get(svc interface{}) interface{} {
	return i.bus.Get(svc)
}

func (i *Integration) migrate() {
	migrator := i.bus.Get("migrator").(*migrate.Migrate)
	err := migrator.Drop()
	if err != nil {
		panic(err)
	}

	migrator = i.bus.Get("migrator").(*migrate.Migrate)
	err = migrator.Up()
	if err != nil {
		panic(err)
	}
}

func GetTestConfig() config.Config {
	c := config.NewConfig()
	c.Environment = "development"
	c.AppName =     "example"
	c.DBHost =      "db"
	c.DBPort =      "5432"
	c.DBName =      "cqrs"
	c.DBUser =      "cqrs"
	c.DBPass =      "cqrs"
	c.Secret =      "secret"
	c.CORSOrigin =  "*"
	c.AppURL =      "localhost:8000"
	return c
}
