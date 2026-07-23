package ports

import (
	"context"
	"time"

	"cinema-ticket/backend/internal/domain"
)

type UserRepository interface {
	FindByGoogleSub(ctx context.Context, googleSub string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	Upsert(ctx context.Context, user *domain.User) (*domain.User, error)
}

type MovieRepository interface {
	List(ctx context.Context) ([]*domain.Movie, error)
	FindByID(ctx context.Context, id string) (*domain.Movie, error)
	Upsert(ctx context.Context, movie *domain.Movie) error
}

type CinemaRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Cinema, error)
	Upsert(ctx context.Context, cinema *domain.Cinema) error
}

type ShowtimeRepository interface {
	ListByMovie(ctx context.Context, movieID string) ([]*domain.Showtime, error)
	FindByID(ctx context.Context, id string) (*domain.Showtime, error)
	Upsert(ctx context.Context, showtime *domain.Showtime) error
}

type ShowtimeSeatRepository interface {
	Insert(ctx context.Context, seat *domain.ShowtimeSeat) error
	ListByShowtime(ctx context.Context, showtimeID string) ([]*domain.ShowtimeSeat, error)
	FindOne(ctx context.Context, showtimeID, seatLabel string) (*domain.ShowtimeSeat, error)

	MarkLocked(ctx context.Context, showtimeID, seatLabel, userID, lockToken string, expiresAt, now time.Time) (bool, error)

	MarkAvailable(ctx context.Context, showtimeID, seatLabel string) error

	MarkBooked(ctx context.Context, showtimeID, seatLabel, lockToken, bookingID string) (bool, error)

	AttachBooking(ctx context.Context, showtimeID, seatLabel, bookingID string) error

	FindExpiredLocked(ctx context.Context, now time.Time) ([]*domain.ShowtimeSeat, error)
}

type BookingFilter struct {
	MovieID      string
	ShowtimeDate string
	UserEmail    string
}

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus, paidAt *time.Time) error

	ListFiltered(ctx context.Context, filter BookingFilter) ([]*domain.Booking, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error)
}
