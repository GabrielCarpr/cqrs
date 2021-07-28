package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
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
	return &PostgresEventStore{db: db, closing: make(chan struct{})}
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
	db      *sql.DB
	closed  bool
	closing chan struct{}
	wg      sync.WaitGroup
}

func (s *PostgresEventStore) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true

	close(s.closing)
	s.wg.Wait()

	return s.db.Close()
}

func (s *PostgresEventStore) Append(ctx context.Context, v bus.ExpectedVersion, events ...bus.Event) error {
	s.wg.Add(1)
	defer s.wg.Done()

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
		switch {
		case err != nil && strings.Contains(err.Error(), "unique constraint"):
			tx.Rollback()
			return eventstore.ErrConcurrencyViolation
		case err == nil:
			continue
		default:
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (s *PostgresEventStore) lastEvent(tx *sql.Tx, id bus.StreamID) (bus.Event, error) {
	row := tx.QueryRow("SELECT payload FROM events WHERE owner = $1 and type = $2 ORDER BY version DESC LIMIT 1 FOR UPDATE", id.ID, id.Type)
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
	if s.closed {
		return errors.New("store is closed")
	}
	s.wg.Add(1)
	defer s.wg.Done()
	defer close(stream)

	query, args := s.buildStreamQuery(q)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var data []byte
	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

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

func (s *PostgresEventStore) buildStreamQuery(q bus.Select) (string, []interface{}) {
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

	return strings.Join(query, " "), args
}

func (s *PostgresEventStore) Subscribe(ctx context.Context, subscribe func(bus.Event) error) (err error) {
	if s.closed {
		return errors.New("store is closed")
	}
	s.wg.Add(1)
	defer s.wg.Done()

	backoff := 0
	for {
		select {
		case <-s.closing:
			log.Info(ctx, "subscriber closing", log.F{})
			return nil
		case <-ctx.Done():
			log.Info(ctx, "subscribe context finished", log.F{"reason": ctx.Err().Error()})
			return nil
		case <-time.After(time.Millisecond * time.Duration(backoff)):
			break
		}

		log.Info(ctx, "checking for new events", log.F{})
		err, processed := s.run(ctx, subscribe)
		switch {
		case err != nil:
			backoff = 100
		case !processed:
			backoff = 1000
		default:
			backoff = 0
		}
	}
}

func (s *PostgresEventStore) run(ctx context.Context, subscribe func(bus.Event) error) (error, bool) {
	var tx *sql.Tx
	tx, err := s.db.Begin()
	if err != nil {
		return log.Error(ctx, "could not open transaction", log.F{"error": err.Error()}), false
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
	err = tx.QueryRow(claim, Now().Add(-time.Minute)).Scan(&offset)
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return log.Error(ctx, "Error claiming event", log.F{"error": err.Error()}), false
	}
	if err != nil && err == sql.ErrNoRows {
		tx.Rollback()
		return nil, false
	}

	var data []byte
	err = tx.QueryRow(`SELECT payload FROM events WHERE "offset" = $1`, offset).Scan(&data)
	if err != nil {
		tx.Rollback()
		return log.Error(ctx, "could not get event payload", log.F{"error": err.Error()}), false
	}

	var msg message.Message
	msg, err = bus.DeserializeMessage(data)
	if err != nil {
		tx.Rollback()
		return log.Error(ctx, "could not deserialize event message", log.F{"error": err.Error()}), false
	}

	err = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = log.Error(ctx, "panicked subscribing", log.F{"panic": fmt.Sprint(r)})
			}
		}()
		err = subscribe(msg.(bus.Event))
		if err != nil {
			return log.Error(ctx, "error when subscribing event", log.F{"error": err.Error()})
		}
		if s.closed || ctx.Err() != nil {
			return log.Error(ctx, "event store closed mid subscribe, discarding", log.F{"event": msg.(bus.Event).Event()})
		}
		return nil
	}()
	if err != nil {
		tx.Rollback()
		return err, false
	}

	_, err = tx.Exec(`UPDATE events SET acked_at = $1 WHERE "offset" = $2`, Now(), offset)
	if err != nil {
		tx.Rollback()
		return log.Error(ctx, "error marking event as acked", log.F{"error": err.Error()}), false
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return log.Error(ctx, "error commiting subscribed event", log.F{"error": err.Error()}), false
	}

	log.Info(ctx, "processed event subscription", log.F{"event": msg.(bus.Event).Event()})
	return nil, true
}
