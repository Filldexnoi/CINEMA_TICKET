package mongo

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	dbName := "cinema"
	if idx := strings.LastIndex(uri, "/"); idx != -1 {
		if candidate := strings.SplitN(uri[idx+1:], "?", 2)[0]; candidate != "" {
			dbName = candidate
		}
	}

	return client.Database(dbName), nil
}

type Pinger struct {
	db *mongo.Database
}

func NewPinger(db *mongo.Database) *Pinger {
	return &Pinger{db: db}
}

func (p *Pinger) Ping(ctx context.Context) error {
	return p.db.Client().Ping(ctx, nil)
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if _, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "google_sub", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	seatIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "showtime_id", Value: 1}, {Key: "seat_label", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "showtime_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "lock_expires_at", Value: 1}}},
	}
	if _, err := db.Collection("showtime_seats").Indexes().CreateMany(ctx, seatIndexes); err != nil {
		return err
	}

	if _, err := db.Collection("showtimes").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "movie_id", Value: 1}},
	}); err != nil {
		return err
	}

	bookingIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "movie_id", Value: 1}}},
		{Keys: bson.D{{Key: "showtime_date", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	if _, err := db.Collection("bookings").Indexes().CreateMany(ctx, bookingIndexes); err != nil {
		return err
	}

	if _, err := db.Collection("audit_logs").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "occurred_at", Value: -1}},
	}); err != nil {
		return err
	}

	return nil
}
