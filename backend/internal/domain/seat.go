package domain

import "time"

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type ShowtimeSeat struct {
	ID            string     `bson:"_id,omitempty" json:"id"`
	ShowtimeID    string     `bson:"showtime_id" json:"showtime_id"`
	SeatLabel     string     `bson:"seat_label" json:"seat_label"`
	Row           int        `bson:"row" json:"row"`
	Col           int        `bson:"col" json:"col"`
	Status        SeatStatus `bson:"status" json:"status"`
	LockedBy      string     `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
	LockToken     string     `bson:"lock_token,omitempty" json:"-"`
	LockExpiresAt *time.Time `bson:"lock_expires_at,omitempty" json:"lock_expires_at,omitempty"`
	BookingID     string     `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	Price         float64    `bson:"price" json:"price"`
	UpdatedAt     time.Time  `bson:"updated_at" json:"updated_at"`
}

func (s *ShowtimeSeat) IsStaleLock(now time.Time) bool {
	return s.Status == SeatLocked && s.LockExpiresAt != nil && s.LockExpiresAt.Before(now)
}
