package background

import (
	"context"
	"fmt"
	"os"

	"github.com/GabrielCarpr/cqrs/bus"

	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di/v2"
)

func BuildConfig(appName string) Config {
	return Config{
		AppName: appName,
		DBHost:  os.Getenv("DB_HOST"),
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASS"),
		DBName:  os.Getenv("DB_NAME"),
	}
}

type Config struct {
	AppName string
	DBHost  string
	DBUser  string
	DBPass  string
	DBName  string
}

func (c Config) DBDsn() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=disable",
		c.DBUser, c.DBPass, c.DBName, c.DBHost,
	)
}

func Build(c Config) *Service {
	builder, _ := di.NewBuilder()

	builder.Add(di.Def{
		Name: "db",
		Build: func(ctn di.Container) (interface{}, error) {
			return sqlx.MustConnect("postgres", c.DBDsn()), nil
		},
	})

	builder.Add(di.Def{
		Name: "repository",
		Build: func(ctn di.Container) (interface{}, error) {
			db := ctn.Get("db").(*sqlx.DB)
			return NewRepository(c, db), nil
		},
	})

	builder.Add(di.Def{
		Name: "controller",
		Build: func(ctn di.Container) (interface{}, error) {
			repo := ctn.Get("repository").(*Repository)
			return NewController(repo), nil
		},
	})

	ctn := builder.Build()

	return &Service{ctn: ctn}
}

var _ bus.Plugin = (*Service)(nil)

type Service struct {
	ctn    di.Container
	cancel context.CancelFunc
	b      *bus.Bus
}

func (s *Service) Controller() *Controller {
	return s.ctn.Get("controller").(*Controller)
}

func (s *Service) Register(b *bus.Bus) error {
	s.b = b
	s.Controller().RegisterQueueAction(func(ctx context.Context, cmd bus.Command) error {
		_, err := s.b.Dispatch(ctx, cmd, false)
		return err
	})
	s.b.Use(s.Controller().JobFinishingMiddleware)

	return nil
}

// Run blocks until the context cancels, or the worker exits
func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	return s.Controller().Run(ctx)
}

func (s *Service) RegisterJob(j Job) error {
	repo := s.ctn.Get("repository").(*Repository)
	return repo.Store(j)
}

func (s *Service) Close() error {
	s.cancel()
	s.ctn.Delete()
	return nil
}
