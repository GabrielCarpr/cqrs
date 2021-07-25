package postgres

import (
	"log"
	"strings"

	"database/sql"
)

// PostgreSQLSchema is an implementation of SchemaAdapter based on PostgreSQL.
type PostgreSQLSchema struct {
	Config Config
}

func (s PostgreSQLSchema) Make() error {
	log.Print("Creating event store")
	db, err := sql.Open("postgres", s.Config.DBDsn())
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS events (
		"offset" SERIAL,
		"owner" VARCHAR(36) NOT NULL,
		"type" VARCHAR(64) NOT NULL,
		"at" TIMESTAMP NOT NULL,
		"version" BIGINT,
		"payload" JSON NOT NULL
	);`)

	return err
}

func (s PostgreSQLSchema) Reset() {
	log.Print("Resetting event store")
	db, err := sql.Open("postgres", s.Config.DBDsn())
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("DELETE FROM events")
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		panic(err)
	}
}
