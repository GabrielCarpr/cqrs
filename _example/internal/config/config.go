package config

import (
	"fmt"
	"os"
)

var Values Config

func init() {
	Values = NewConfig()
}

// NewConfig returns a Config object
func NewConfig() Config {
	return Config{
		AppName:     "example",
		Environment: requiredS("ENVIRONMENT"),
		DBHost:      requiredS("DB_HOST"),
		DBPort:      requiredS("DB_PORT"),
		DBUser:      requiredS("DB_USER"),
		DBName:      requiredS("DB_NAME"),
		DBPass:      requiredS("DB_PASS"),
		Secret:      requiredS("SECRET_KEY"),
		CORSOrigin:  defaultS("CORS_ORIGIN", ""),
		AppURL:      defaultS("APP_URL", "http://localhost:8080"),
	}
}

// Config is application config
type Config struct {
	Environment string

	AppName string

	DBHost string
	DBPort string
	DBUser string
	DBName string
	DBPass string

	Secret     string
	CORSOrigin string

	AppURL string
}

func (c Config) DBDsn() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=disable",
		c.DBUser, c.DBPass, c.DBName, c.DBHost,
	)
}

func defaultS(key string, dflt string) string {
	value := os.Getenv(key)
	if value == "" {
		return dflt
	}
	return value
}

func requiredS(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Config value %s is required", key))
	}
	return value
}
