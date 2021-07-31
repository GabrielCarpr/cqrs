package postgres

import "fmt"

type Config struct {
	DBName string
	DBPass string
	DBHost string
	DBUser string
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
