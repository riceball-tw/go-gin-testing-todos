package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	slogchannel "github.com/samber/slog-channel"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoLogBufferSize    = 1000
	mongoLogInsertTimeout = 2 * time.Second
	defaultMongoURI       = "mongodb://localhost:27017"
	defaultMongoLogDB     = "todo_db"
	defaultMongoLogColl   = "logs"
)

type mongoLogWorker struct {
	collection *mongo.Collection
	records    chan *slog.Record
	done       chan struct{}
}

func newMongoLogHandler() slog.Handler {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(envOrDefault("MONGO_URI", defaultMongoURI)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create mongodb log client: %v\n", err)
		return nil
	}

	collection := client.
		Database(envOrDefault("MONGO_LOG_DB", defaultMongoLogDB)).
		Collection(envOrDefault("MONGO_LOG_COLLECTION", defaultMongoLogColl))

	worker := &mongoLogWorker{
		collection: collection,
		records:    make(chan *slog.Record, mongoLogBufferSize),
		done:       make(chan struct{}),
	}
	go worker.run()

	handler := slogchannel.Option{
		Level:    slog.LevelInfo,
		Channel:  worker.records,
		Blocking: false,
	}.NewChannelHandler()

	return handler
}

func (w *mongoLogWorker) run() {
	defer close(w.done)

	for record := range w.records {
		ctx, cancel := context.WithTimeout(context.Background(), mongoLogInsertTimeout)
		_, err := w.collection.InsertOne(ctx, slogRecordDocument(record))
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write log to mongodb: %v\n", err)
		}
	}
}

func (w *mongoLogWorker) shutdown() {
	close(w.records)
	<-w.done
}

func slogRecordDocument(record *slog.Record) bson.M {
	doc := bson.M{
		"time":        record.Time,
		"level":       record.Level.String(),
		"level_value": int(record.Level),
		"message":     record.Message,
	}

	attrs := bson.M{}
	record.Attrs(func(attr slog.Attr) bool {
		addSlogAttr(attrs, attr)
		return true
	})

	if len(attrs) > 0 {
		doc["attrs"] = attrs
	}

	return doc
}

func addSlogAttr(attrs bson.M, attr slog.Attr) {
	if attr.Key == "" {
		return
	}

	attrs[attr.Key] = slogValue(attr.Value)
}

func slogValue(value slog.Value) any {
	value = value.Resolve()

	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindBool:
		return value.Bool()
	case slog.KindDuration:
		return value.Duration()
	case slog.KindTime:
		return value.Time()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindUint64:
		return value.Uint64()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindGroup:
		group := bson.M{}
		for _, attr := range value.Group() {
			addSlogAttr(group, attr)
		}
		return group
	case slog.KindAny:
		return bsonSafeValue(value.Any())
	default:
		return value.String()
	}
}

func bsonSafeValue(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case error:
		return v.Error()
	default:
		return v
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
