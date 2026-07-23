package domain

import "time"

type EventType string

const (
	EventSeatLocked       EventType = "seat.locked"
	EventSeatReleased     EventType = "seat.released"
	EventBookingConfirmed EventType = "booking.confirmed"
	EventBookingExpired   EventType = "booking.expired"
)

type SeatEvent struct {
	EventID    string    `json:"event_id"`
	EventType  EventType `json:"event_type"`
	ShowtimeID string    `json:"showtime_id"`
	SeatLabels []string  `json:"seat_labels"`
	UserID     string    `json:"user_id,omitempty"`
	BookingID  string    `json:"booking_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}
