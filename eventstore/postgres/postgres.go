package postgres

import (
	"bytes"
	"context"
	stdSQL "database/sql"
	"encoding/gob"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/pkg/sql"
	wmMessage "github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/lib/pq"
)

func New(c Config) *PostgresEventStore {
	db := makeDB(c)
	var logger watermill.LoggerAdapter
	publisher, err := sql.NewPublisher(
		db,
		sql.PublisherConfig{
			SchemaAdapter:        PostgreSQLSchema{},
			AutoInitializeSchema: true,
		},
		logger,
	)
	if err != nil {
		panic(err)
	}
	return &PostgresEventStore{db, publisher}
}

func makeDB(c Config) *stdSQL.DB {
	db, err := stdSQL.Open("postgres", c.DBDsn())
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
	db        *stdSQL.DB
	publisher wmMessage.Publisher
}

func (s *PostgresEventStore) Close() error {
	return s.publisher.Close()
}

func (s *PostgresEventStore) Append(ctx context.Context, v bus.ExpectedVersion, events ...bus.Event) error {
	msgs := make([]*wmMessage.Message, 0)
	for _, event := range events {
		msg, err := s.toMessage(ctx, event)
		if err != nil {
			return err
		}
		msgs = append(msgs, msg)
	}

	return s.publisher.Publish("event_store", msgs...)
}

func (s *PostgresEventStore) toMessage(ctx context.Context, event bus.Event) (*wmMessage.Message, error) {
	var payload bytes.Buffer
	enc := gob.NewEncoder(&payload)
	err := enc.Encode(event)
	if err != nil {
		return nil, err
	}

	result := wmMessage.NewMessage(watermill.NewUUID(), payload.Bytes())
	return result, nil
}

func (s *PostgresEventStore) Stream(ctx context.Context, stream bus.Stream, q bus.Select) error {
	defer close(stream)
	var logger watermill.LoggerAdapter
	subscriber, err := sql.NewSubscriber(s.db, sql.SubscriberConfig{
		SchemaAdapter:    PostgreSQLSchema{},
		OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
		InitializeSchema: true,
	}, logger)
	if err != nil {
		return err
	}

	messages, err := subscriber.Subscribe(ctx, "event_store")
	if err != nil {
		return err
	}

	for msg := range messages {
		_, event, err := s.fromMessage(msg)
		if err != nil {
			return err
		}

		stream <- event
	}

	return nil
}

func (s *PostgresEventStore) fromMessage(msg *wmMessage.Message) (context.Context, bus.Event, error) {
	var result bus.Event
	ctx := context.Background()
	dec := gob.NewDecoder(bytes.NewBuffer(msg.Payload))
	err := dec.Decode(&result)
	return ctx, result, err
}

func (s *PostgresEventStore) Subscribe(ctx context.Context, subscribe func(bus.Event) error) error {
	return nil
}
