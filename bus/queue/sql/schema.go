package sql

import (
	"log"
	"strings"

	stdSQL "database/sql"

	wmSql "github.com/ThreeDotsLabs/watermill-sql/pkg/sql"
)

// PostgreSQLSchema is an implementation of SchemaAdapter based on PostgreSQL.
type PostgreSQLSchema struct {
	wmSql.DefaultPostgreSQLSchema
}

func (s PostgreSQLSchema) SchemaInitializingQueries(topic string) []string {
	createMessagesTable := strings.Join([]string{
		`CREATE TABLE IF NOT EXISTS ` + s.MessagesTable(topic) + ` (`,
		`"offset" SERIAL,`,
		`"uuid" VARCHAR(36) NOT NULL,`,
		`"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,`,
		`"payload" BYTEA DEFAULT NULL,`,
		`"metadata" JSON DEFAULT NULL`,
		`);`,
	}, "\n")

	return []string{createMessagesTable}
}

func ResetSQLDB(dsn string) {
	log.Print("Resetting SQL queue database")
	db, err := stdSQL.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("DELETE FROM watermill_messages")
	if err != nil {
		panic(err)
	}
}
