package mongo

import (
	"context"
	"time"

	"cinema-ticket/backend/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ShowtimeSeatRepo struct{ col *mongo.Collection }

func NewShowtimeSeatRepo(db *mongo.Database) *ShowtimeSeatRepo {
	return &ShowtimeSeatRepo{col: db.Collection("showtime_seats")}
}

func (r *ShowtimeSeatRepo) Insert(ctx context.Context, seat *domain.ShowtimeSeat) error {
	_, err := r.col.InsertOne(ctx, seat)
	return err
}

func (r *ShowtimeSeatRepo) ListByShowtime(ctx context.Context, showtimeID string) ([]*domain.ShowtimeSeat, error) {
	cur, err := r.col.Find(ctx, bson.M{"showtime_id": showtimeID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var seats []*domain.ShowtimeSeat
	if err := cur.All(ctx, &seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *ShowtimeSeatRepo) FindOne(ctx context.Context, showtimeID, seatLabel string) (*domain.ShowtimeSeat, error) {
	var s domain.ShowtimeSeat
	err := r.col.FindOne(ctx, bson.M{"showtime_id": showtimeID, "seat_label": seatLabel}).Decode(&s)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShowtimeSeatRepo) MarkLocked(ctx context.Context, showtimeID, seatLabel, userID, lockToken string, expiresAt, now time.Time) (bool, error) {
	filter := bson.M{
		"showtime_id": showtimeID,
		"seat_label":  seatLabel,
		"$or": []bson.M{
			{"status": domain.SeatAvailable},
			{"status": domain.SeatLocked, "lock_expires_at": bson.M{"$lt": now}},
		},
	}
	update := bson.M{"$set": bson.M{
		"status":          domain.SeatLocked,
		"locked_by":       userID,
		"lock_token":      lockToken,
		"lock_expires_at": expiresAt,
		"updated_at":      now,
	}}
	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.ModifiedCount == 1, nil
}

func (r *ShowtimeSeatRepo) MarkAvailable(ctx context.Context, showtimeID, seatLabel string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"showtime_id": showtimeID, "seat_label": seatLabel},
		bson.M{
			"$set":   bson.M{"status": domain.SeatAvailable, "updated_at": time.Now().UTC()},
			"$unset": bson.M{"locked_by": "", "lock_token": "", "lock_expires_at": "", "booking_id": ""},
		},
	)
	return err
}

func (r *ShowtimeSeatRepo) AttachBooking(ctx context.Context, showtimeID, seatLabel, bookingID string) error {
	_, err := r.col.UpdateOne(ctx,
		bson.M{"showtime_id": showtimeID, "seat_label": seatLabel},
		bson.M{"$set": bson.M{"booking_id": bookingID, "updated_at": time.Now().UTC()}},
	)
	return err
}

func (r *ShowtimeSeatRepo) MarkBooked(ctx context.Context, showtimeID, seatLabel, lockToken, bookingID string) (bool, error) {
	filter := bson.M{
		"showtime_id": showtimeID,
		"seat_label":  seatLabel,
		"status":      domain.SeatLocked,
		"lock_token":  lockToken,
	}
	update := bson.M{
		"$set":   bson.M{"status": domain.SeatBooked, "booking_id": bookingID, "updated_at": time.Now().UTC()},
		"$unset": bson.M{"locked_by": "", "lock_token": "", "lock_expires_at": ""},
	}
	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.ModifiedCount == 1, nil
}

func (r *ShowtimeSeatRepo) FindExpiredLocked(ctx context.Context, now time.Time) ([]*domain.ShowtimeSeat, error) {
	cur, err := r.col.Find(ctx, bson.M{
		"status":          domain.SeatLocked,
		"lock_expires_at": bson.M{"$lt": now},
	})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var seats []*domain.ShowtimeSeat
	if err := cur.All(ctx, &seats); err != nil {
		return nil, err
	}
	return seats, nil
}
