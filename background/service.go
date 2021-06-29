package background

import (
	"fmt"
	"os"

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

	return &Service{ctn}
}

type Service struct {
	ctn di.Container
}

func (s *Service) Controller() *Controller {
	return s.ctn.Get("controller").(*Controller)
}

func (s *Service) AttachRouter(qa queueAction) {
	s.Controller().RegisterQueueAction(qa)
}

func (s *Service) Work(block bool) {
	if block {
		s.Controller().Block(0)
	} else {
		c := make(chan bool)
		s.Controller().Run(c)
	}
}

func (s *Service) RegisterJob(j Job) error {
	repo := s.ctn.Get("repository").(*Repository)
	return repo.Store(j)
}

func (s *Service) Delete() {
	s.ctn.Delete()
}
