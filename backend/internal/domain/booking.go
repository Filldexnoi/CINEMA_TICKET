package domain

import "time"

type BookingStatus string

const (
	BookingPending   BookingStatus = "PENDING"
	BookingConfirmed BookingStatus = "CONFIRMED"
	BookingExpired   BookingStatus = "EXPIRED"
	BookingFailed    BookingStatus = "FAILED"
)

type Booking struct {
	ID          string        `bson:"_id,omitempty" json:"id"`
	UserID      string        `bson:"user_id" json:"user_id"`
	ShowtimeID  string        `bson:"showtime_id" json:"showtime_id"`
	SeatLabels  []string      `bson:"seat_labels" json:"seat_labels"`
	LockToken   string        `bson:"lock_token" json:"-"`
	TotalAmount float64       `bson:"total_amount" json:"total_amount"`
	Status      BookingStatus `bson:"status" json:"status"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	ExpiresAt   time.Time     `bson:"expires_at" json:"expires_at"`
	PaidAt      *time.Time    `bson:"paid_at,omitempty" json:"paid_at,omitempty"`

	MovieID      string `bson:"movie_id" json:"movie_id"`
	ShowtimeDate string `bson:"showtime_date" json:"showtime_date"`
	UserEmail    string `bson:"user_email" json:"user_email"`
}
