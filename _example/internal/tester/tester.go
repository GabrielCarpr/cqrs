package tester

import (
	"example/internal/app"
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/bus"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"fmt"

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
	i.migrate()
	i.bus = bus.Default(context.Background(), append(app.Modules, testModule(i.Doubles)))
}

func (i *Integration) TearDownTest() {
	i.CloseBus()
	i.Doubles = []bus.Def{}
}

func (i *Integration) CloseBus() {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	i.bus.Close()
}

func (i *Integration) Bus() *bus.Bus {
	return i.bus
}

func (i *Integration) Get(svc interface{}) interface{} {
	return i.bus.Get(svc)
}

func (i *Integration) migrate() {
	i.CloseBus()
	migs := config.Values.Migrations
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migs),
		fmt.Sprintf(
			"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			config.Values.DBUser,
			config.Values.DBPass,
			config.Values.DBHost,
			config.Values.DBPort,
			config.Values.DBName,
		),
	)
	if err != nil {
		panic(err)
	}
	defer m.Close()

	err = m.Drop()
	if err != nil {
		panic(err)
	}
	err = m.Up()
	if err != nil {
		//panic(err)
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
