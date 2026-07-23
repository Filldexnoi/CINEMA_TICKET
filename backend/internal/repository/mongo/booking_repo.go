package mongo

import (
	"context"
	"errors"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BookingRepo struct{ col *mongo.Collection }

func NewBookingRepo(db *mongo.Database) *BookingRepo {
	return &BookingRepo{col: db.Collection("bookings")}
}

func (r *BookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	_, err := r.col.InsertOne(ctx, booking)
	return err
}

func (r *BookingRepo) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	var b domain.Booking
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&b)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus, paidAt *time.Time) error {
	set := bson.M{"status": status}
	if paidAt != nil {
		set["paid_at"] = *paidAt
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

func (r *BookingRepo) ListFiltered(ctx context.Context, filter ports.BookingFilter) ([]*domain.Booking, error) {
	query := bson.M{}
	if filter.MovieID != "" {
		query["movie_id"] = filter.MovieID
	}
	if filter.ShowtimeDate != "" {
		query["showtime_date"] = filter.ShowtimeDate
	}
	if filter.UserEmail != "" {
		query["user_email"] = bson.M{"$regex": filter.UserEmail, "$options": "i"}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(200)
	cur, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	bookings := []*domain.Booking{}
	if err := cur.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}
