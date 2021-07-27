package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/bus/message"
	"github.com/GabrielCarpr/cqrs/eventstore"
	"github.com/GabrielCarpr/cqrs/log"
	_ "github.com/lib/pq"
)

var Now = time.Now

func New(c Config) *PostgresEventStore {
	db := makeDB(c)
	schema := PostgreSQLSchema{c}
	err := schema.Make()
	if err != nil {
		panic(err)
	}
	return &PostgresEventStore{db: db}
}

func makeDB(c Config) *sql.DB {
	db, err := sql.Open("postgres", c.DBDsn())
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

type PostgresEventStore struct {
	db     *sql.DB
	closed bool
}

func (s *PostgresEventStore) Close() error {
	s.closed = true
	return s.db.Close()
}

func (s *PostgresEventStore) Append(ctx context.Context, v bus.ExpectedVersion, events ...bus.Event) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	last, err := s.lastEvent(tx, bus.StreamID{ID: events[0].Owned(), Type: events[0].FromAggregate()})
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := eventstore.CheckExpectedVersion(last, v); err != nil {
		tx.Rollback()
		return err
	}
	if err := eventstore.CheckEventsConsistent(events...); err != nil {
		tx.Rollback()
		return err
	}

	for _, event := range events {
		data, err := bus.SerializeMessage(event, bus.Json)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec(`INSERT INTO events (owner, type, at, version, payload)
			VALUES ($1, $2, $3, $4, $5)`,
			event.Owned(),
			event.FromAggregate(),
			event.WasPublishedAt(),
			event.Versioned(),
			data,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (s *PostgresEventStore) lastEvent(tx *sql.Tx, id bus.StreamID) (bus.Event, error) {
	row := tx.QueryRow("SELECT payload FROM events WHERE owner = $1 and type = $2 ORDER BY version DESC LIMIT 1", id.ID, id.Type)
	var data []byte
	err := row.Scan(&data)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}

	msg, err := bus.DeserializeMessage(data)
	if err != nil {
		return nil, err
	}
	return msg.(bus.Event), nil
}

func (s *PostgresEventStore) Stream(ctx context.Context, stream bus.Stream, q bus.Select) error {
	defer close(stream)

	query := []string{}
	args := []interface{}{}
	argNum := 1
	query = append(query, "SELECT payload FROM events")
	if q.StreamID.ID != "" || q.StreamID.Type != "" || q.From != 0 {
		query = append(query, "WHERE")
	}
	if q.StreamID.ID != "" {
		query = append(query, fmt.Sprintf("owner = $%v", argNum))
		argNum++
		args = append(args, q.StreamID.ID)
	}
	if q.StreamID.Type != "" {
		if argNum != 1 {
			query = append(query, "AND")
		}
		query = append(query, fmt.Sprintf("type = $%v", argNum))
		argNum++
		args = append(args, q.StreamID.Type)
	}
	if q.From != 0 {
		if argNum != 1 {
			query = append(query, "AND")
		}
		query = append(query, fmt.Sprintf("version >= $%v", argNum))
		argNum++
		args = append(args, q.From)
	}
	query = append(query, "ORDER BY version ASC")
	rows, err := s.db.QueryContext(ctx, strings.Join(query, " "), args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var data []byte
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return err
		}

		event, err := bus.DeserializeMessage(data)
		if err != nil {
			return err
		}
		stream <- event.(bus.Event)
	}
	err = rows.Err()

	return err
}

func (s *PostgresEventStore) Subscribe(ctx context.Context, subscribe func(bus.Event) error) (err error) {
	backoff := 0
	for {
		if ctx.Err() != nil || s.closed {
			log.Info(ctx, "ending postgres event store subscribe loop", log.F{})
			break
		}
		<-time.After(time.Millisecond * time.Duration(backoff))
		log.Info(ctx, "checking for new events", log.F{})

		var tx *sql.Tx
		tx, err = s.db.Begin()
		if err != nil {
			log.Error(ctx, "could not open transaction", log.F{"error": err.Error()})
			backoff = 1000
			continue
		}

		claim := `
		UPDATE events
		SET reserved_at = NOW()
		WHERE "offset" = (
			SELECT "offset"
			FROM events
			WHERE (reserved_at IS NULL
				OR (reserved_at < $1 AND acked_at IS NULL))
			ORDER BY "offset" ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING "offset"`

		var offset int
		err1 := tx.QueryRow(claim, Now().Add(-time.Minute)).Scan(&offset)
		if err1 != nil && err1 != sql.ErrNoRows {
			err = err1
			tx.Rollback()
			log.Warn(ctx, "Error claiming event", log.F{"error": err1.Error()})
			continue
		}
		if err1 != nil && err1 == sql.ErrNoRows {
			tx.Rollback()
			backoff = 1000
			continue
		}

		var data []byte
		err = tx.QueryRow(`SELECT payload FROM events WHERE "offset" = $1`, offset).Scan(&data)
		if err != nil {
			tx.Rollback()
			log.Warn(ctx, "could not get event payload", log.F{"error": err.Error()})
			continue
		}

		var msg message.Message
		msg, err = bus.DeserializeMessage(data)
		if err != nil {
			tx.Rollback()
			log.Warn(ctx, "could not deserialize event message", log.F{"error": err.Error()})
			continue
		}

		err = subscribe(msg.(bus.Event))
		if err != nil {
			tx.Rollback()
			log.Info(ctx, "error when subscribing event", log.F{"error": err.Error()})
			continue
		}
		if s.closed {
			tx.Rollback()
			log.Info(ctx, "event store closed mid subscribe, discarding", log.F{"event": msg.(bus.Event).Event()})
		}

		_, err = tx.Exec(`UPDATE events SET acked_at = $1 WHERE "offset" = $2`, Now(), offset)
		if err != nil {
			tx.Rollback()
			log.Warn(ctx, "error marking event as acked", log.F{"error": err.Error()})
			continue
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			log.Warn(ctx, "error commiting subscribed event", log.F{"error": err.Error()})
			continue
		}
		log.Info(ctx, "processed event subscription", log.F{"event": msg.(bus.Event).Event()})
		backoff = 0
	}

	return nil
}
