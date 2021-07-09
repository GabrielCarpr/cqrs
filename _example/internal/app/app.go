package app

import (
	"example/internal/config"
	"github.com/GabrielCarpr/cqrs/auth"
	"github.com/GabrielCarpr/cqrs/background"
	"github.com/GabrielCarpr/cqrs/errors"
	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"
	"github.com/GabrielCarpr/cqrs/bus/queue/sql"
	"example/users"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	stdlog "log"
	"net/http"
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

var modules = []bus.Module{
	users.Users{},
	Main,
}

// Make Builds and returns the app
func Make() *App {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	queue := sql.NewSQLQueue(sql.Config{
		DBUser: config.Values.DBUser,
		DBPass: config.Values.DBPass,
		DBHost: config.Values.DBHost,
		DBName: config.Values.DBName,
	})

	// Bus setup
	b := bus.Default(ctx, modules, bus.UseQueue(queue))
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

	// Background setup
	bgConf := background.BuildConfig("example")
	bg := background.Build(bgConf)
	bg.AttachRouter(func(ctx context.Context, c bus.Command) error {
		_, err := b.Dispatch(ctx, c, false)
		return err
	})
	b.Use(bg.Controller().JobFinishingMiddleware)
	b.RegisterDeletion(func() {
		bg.Delete()
	})
	b.RegisterWork(func() {
		bg.Work(false)
	})

	app := App{Bus: b}
	return &app
}

type App struct {
	Bus    *bus.Bus
}

func (a *App) Handle() {
	mux := http.NewServeMux()
	stdlog.Fatal(http.ListenAndServe(":80", mux))
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
					m, err := migrate.New(
						"file:///var/migrations",
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
			},
		}
	},
}
