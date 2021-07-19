package app

import (
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/errors"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"
	"github.com/GabrielCarpr/cqrs/bus/queue/sql"
	"github.com/GabrielCarpr/cqrs/ports"
	"example/rest"
	"example/users"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	stdlog "log"
	"os/signal"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	b := bus.Default(ctx, Modules, bus.UseQueue(queue))
	b.Use(
		auth.CommandAuthGuard,
		auth.QueryAuthGuard,
		errors.CommandErrorMiddleware,
		errors.QueryErrorMiddleware,
	)
	b.RegisterContextKey(auth.AuthCtxKey, func(j []byte) interface{} {
		var v auth.Credentials
		json.Unmarshal(j, &v)
		return v
	})
	b.RegisterContextKey(log.CtxIDKey, func(j []byte) interface{} {
		return uuid.MustParse(string(j))
	})

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

	stdlog.Fatal(p.Run(a.ctx))
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
		
			{
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
			},
		}
	},
}
