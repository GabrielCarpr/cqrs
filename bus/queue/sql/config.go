package sql

import "fmt"

type Config struct {
	DBName string
	DBHost string
	DBUser string
	DBPass string
}

func (c Config) DBDsn() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s sslmode=disable",
		c.DBUser,
		c.DBPass,
		c.DBName,
		c.DBHost,
	)
}
