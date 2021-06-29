package sql

import (
	"strings"

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
