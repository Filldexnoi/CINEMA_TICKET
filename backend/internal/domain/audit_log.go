package domain

import "time"

type AuditEventType string

const (
	AuditBookingSuccess AuditEventType = "BOOKING_SUCCESS"
	AuditBookingTimeout AuditEventType = "BOOKING_TIMEOUT"
	AuditSeatReleased   AuditEventType = "SEAT_RELEASED"
	AuditSystemError    AuditEventType = "SYSTEM_ERROR"
)

type AuditLog struct {
	ID         string         `bson:"_id,omitempty" json:"id"`
	EventType  AuditEventType `bson:"event_type" json:"event_type"`
	Message    string         `bson:"message" json:"message"`
	BookingID  string         `bson:"booking_id,omitempty" json:"booking_id,omitempty"`
	UserID     string         `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ShowtimeID string         `bson:"showtime_id,omitempty" json:"showtime_id,omitempty"`
	SeatLabels []string       `bson:"seat_labels,omitempty" json:"seat_labels,omitempty"`
	OccurredAt time.Time      `bson:"occurred_at" json:"occurred_at"`
}
