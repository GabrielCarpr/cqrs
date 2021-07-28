package app

import (
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"
	"github.com/GabrielCarpr/cqrs/bus/queue/sql"
	"github.com/GabrielCarpr/cqrs/ports"
	pgEventStore "github.com/GabrielCarpr/cqrs/eventstore/postgres"
	"example/rest"
	"example/users"
	"context"
	//"fmt"
	"github.com/google/uuid"
	stdlog "log"
	"os/signal"
	"os"

	//"github.com/golang-migrate/migrate/v4"
	//_ "github.com/golang-migrate/migrate/v4/database/postgres"
	//_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Takes the event store from the DI container and uses it in the bus.
// Necessary was we want to use the event store in repositories too
func useContainerEventStore(name string) bus.Config {
	return func(b *bus.Bus) error {
		s := b.Get(name).(bus.EventStore)
		conf := bus.UseEventStore(s)
		return conf(b)
	}
}

var Modules = []bus.Module{
	users.Users{},
	Main,
}

// Make Builds and returns the app
func Make(ctx context.Context) *App {
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	queue := sql.NewSQLQueue(sql.Config{
		DBUser: config.Values.DBUser,
		DBPass: config.Values.DBPass,
		DBHost: config.Values.DBHost,
		DBName: config.Values.DBName,
	})

	// Bus setup
	b := bus.Default(ctx, Modules, bus.UseQueue(queue), useContainerEventStore("event-store"))
	b.Use(
		auth.CommandAuthGuard,
		auth.QueryAuthGuard,
	)
	bus.RegisterContextKey(auth.AuthCtxKey, auth.Credentials{})
	bus.RegisterContextKey(log.CtxIDKey, uuid.New())

	app := App{Bus: b, ctx: ctx}
	return &app
}

type App struct {
	Bus    *bus.Bus
	ctx context.Context
}

func (a *App) Handle() {
	restServer := rest.Rest(a.Bus, config.Values)
	p := ports.Ports{restServer}

	err := p.Run(a.ctx)
	if err != nil {
		stdlog.Fatal(err)
	}
}

func (a *App) Work() {
	a.Bus.Work()
}

func (a *App) Delete() {
	a.Bus.Close()
}

// Main is the main module, which provides some infrastructure
// dependencies for other modules to use
var Main = bus.FuncModule{
	ServicesFunc: func () []bus.Def {
		return []bus.Def{
			{
				Name: "db",
				Build: func(ctn di.Container) (interface{}, error) {
					return sqlx.MustConnect("postgres", config.Values.DBDsn()), nil
				},
				Close: func(obj interface{}) error {
					db := obj.(*sqlx.DB)
					return db.Close()
				},
			},
		
			{
				Name: "gorm",
				Build: func(ctn di.Container) (interface{}, error) {
					return gorm.Open(postgres.Open(config.Values.DBDsn()), &gorm.Config{})
				},
				Close: func(obj interface{}) error {
					db := obj.(*gorm.DB)
					d, _ := db.DB()
					return d.Close()
				},
			},
		
			/*{
				Name: "migrator",
				Build: func(ctn di.Container) (interface{}, error) {
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
					return m, err
				},
				Close: func(obj interface{}) error {
					m := obj.(*migrate.Migrate)
					err1, err2 := m.Close()
					if err1 != nil {
						return err1
					}
					return err2
				},
				Unshared: true,
			},*/

			{
				Name: "event-store",
				Build: func(ctn di.Container) (interface{}, error) {
					c := pgEventStore.Config{
						DBName: config.Values.DBName,
						DBUser: config.Values.DBUser,
						DBHost: config.Values.DBHost,
						DBPass: config.Values.DBPass,
					}
					store := pgEventStore.New(c)
					return store, nil
				},
				Close: func(obj interface{}) error {
					s := obj.(bus.EventStore)
					return s.Close()
				},
			},
		}
	},
}
