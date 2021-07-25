package sql

import (
	"bytes"
	"context"
	stdSQL "database/sql"
	"encoding/gob"
	"fmt"
	stdlog "log"
	"time"

	"github.com/GabrielCarpr/cqrs/bus"
	"github.com/GabrielCarpr/cqrs/log"

	_ "github.com/lib/pq"

	"github.com/GabrielCarpr/cqrs/bus/message"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-sql/pkg/sql"
	wmMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

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

func NewSQLQueue(c Config) *SQLQueue {
	db := makeDB(c)
	logger := watermill.NewStdLogger(false, false)
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
	return &SQLQueue{db, logger, publisher}
}

type SQLQueue struct {
	db        *stdSQL.DB
	logger    watermill.LoggerAdapter
	publisher wmMessage.Publisher
}

func (q *SQLQueue) Close() {
	stdlog.Print("Closing queue...")
	q.publisher.Close()
}

func (q *SQLQueue) fromMessage(ctx context.Context, msg message.Message) (*wmMessage.Message, error) {
	var payload bytes.Buffer
	enc := gob.NewEncoder(&payload)
	err := enc.Encode(&msg)
	if err != nil {
		return nil, err
	}

	result := wmMessage.NewMessage(watermill.NewUUID(), payload.Bytes())
	result.Metadata = wmMessage.Metadata(bus.SerializeContext(ctx))
	return result, nil
}

func (q *SQLQueue) toMessage(msg *wmMessage.Message) (context.Context, message.Message, error) {
	var result message.Message
	dec := gob.NewDecoder(bytes.NewBuffer(msg.Payload))
	err := dec.Decode(&result)
	metadata := map[string]string(msg.Metadata)
	return bus.DeserializeContext(context.Background(), metadata), result, err
}

func (q *SQLQueue) Subscribe(topCtx context.Context, fn func(context.Context, message.Message) error) {
	router, err := wmMessage.NewRouter(wmMessage.RouterConfig{}, q.logger)
	if err != nil {
		panic(err)
	}
	poison, err := middleware.PoisonQueue(q.publisher, "failures")
	if err != nil {
		panic(err)
	}
	router.AddMiddleware(
		poison,

		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Second * 2,
			Logger:          q.logger,
			Multiplier:      2,
		}.Middleware,
	)

	subscriber, err := sql.NewSubscriber(
		q.db,
		sql.SubscriberConfig{
			SchemaAdapter:    PostgreSQLSchema{},
			OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
			InitializeSchema: true,
		},
		q.logger,
	)
	if err != nil {
		panic(err)
	}

	router.AddNoPublisherHandler(
		"handle",
		"messages",
		subscriber,
		func(msg *wmMessage.Message) error {
			return q.process(fn, msg)
		},
	)

	if err := router.Run(topCtx); err != nil {
		panic(err)
	}
}

func (q *SQLQueue) process(fn func(context.Context, message.Message) error, msg *wmMessage.Message) (err error) {
	ctx, input, err := q.toMessage(msg)

	defer func() {
		if r := recover(); r != nil {
			msg.Nack()
			err = log.Error(ctx, fmt.Errorf("Panicked running message: %v", r), log.F{})
		}
	}()

	log.Info(ctx, "Received message", log.F{"ID": msg.UUID})
	if err != nil {
		msg.Nack()
		return log.Error(ctx, fmt.Errorf("Failed receiving message: %w", err), log.F{"id": msg.UUID})
	}

	err = fn(ctx, input)
	if err != nil {
		msg.Nack()
		return log.Error(ctx, fmt.Errorf("Failed running message: %w", err), log.F{"id": msg.UUID})
	}

	log.Info(ctx, "Message processed", log.F{"id": msg.UUID})
	msg.Ack()
	return nil
}

func (q *SQLQueue) Publish(ctx context.Context, msgs ...message.Message) error {
	for _, msg := range msgs {
		deliver, err := q.fromMessage(ctx, msg)
		if err != nil {
			return err
		}
		deliver.Metadata = wmMessage.Metadata(bus.SerializeContext(ctx))

		log.Info(ctx, "Publishing message", log.F{"ID": deliver.UUID})
		err = q.publisher.Publish("messages", deliver)
		if err != nil {
			return err
		}
	}
	return nil
}
