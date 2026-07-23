package usecase

import (
	"context"
	"fmt"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/google/uuid"
)

type BookingUsecase struct {
	bookings  ports.BookingRepository
	seats     ports.ShowtimeSeatRepository
	lock      ports.DistributedLock
	pub       ports.EventPublisher
	showtimes ports.ShowtimeRepository
	users     ports.UserRepository
	auditLogs ports.AuditLogRepository
}

func NewBookingUsecase(
	bookings ports.BookingRepository,
	seats ports.ShowtimeSeatRepository,
	lock ports.DistributedLock,
	pub ports.EventPublisher,
	showtimes ports.ShowtimeRepository,
	users ports.UserRepository,
	auditLogs ports.AuditLogRepository,
) *BookingUsecase {
	return &BookingUsecase{
		bookings:  bookings,
		seats:     seats,
		lock:      lock,
		pub:       pub,
		showtimes: showtimes,
		users:     users,
		auditLogs: auditLogs,
	}
}

func (u *BookingUsecase) CreateBooking(ctx context.Context, userID, showtimeID string, seatLabels []string, lockToken string) (*domain.Booking, error) {
	var total float64
	var expiresAt time.Time

	for i, label := range seatLabels {
		held, err := u.lock.IsHeldByToken(ctx, lockKey(showtimeID, label), lockToken)
		if err != nil {
			return nil, err
		}
		if !held {
			return nil, fmt.Errorf("%w: %s", ErrLockExpired, label)
		}

		seat, err := u.seats.FindOne(ctx, showtimeID, label)
		if err != nil {
			return nil, err
		}
		if seat == nil || seat.Status != domain.SeatLocked || seat.LockedBy != userID || seat.LockToken != lockToken {
			return nil, fmt.Errorf("%w: %s", ErrNotSeatOwner, label)
		}
		total += seat.Price
		if i == 0 || (seat.LockExpiresAt != nil && seat.LockExpiresAt.Before(expiresAt)) {
			if seat.LockExpiresAt != nil {
				expiresAt = *seat.LockExpiresAt
			}
		}
	}

	showtime, err := u.showtimes.FindByID(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	booking := &domain.Booking{
		ID:          uuid.NewString(),
		UserID:      userID,
		ShowtimeID:  showtimeID,
		SeatLabels:  seatLabels,
		LockToken:   lockToken,
		TotalAmount: total,
		Status:      domain.BookingPending,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
	}
	if showtime != nil {
		booking.MovieID = showtime.MovieID
		booking.ShowtimeDate = showtime.StartTime.Format("2006-01-02")
	}
	if user != nil {
		booking.UserEmail = user.Email
	}
	if err := u.bookings.Create(ctx, booking); err != nil {
		return nil, err
	}

	for _, label := range seatLabels {
		if err := u.seats.AttachBooking(ctx, showtimeID, label, booking.ID); err != nil {
			return nil, err
		}
	}

	return booking, nil
}

func (u *BookingUsecase) GetBooking(ctx context.Context, id string) (*domain.Booking, error) {
	return u.bookings.FindByID(ctx, id)
}

func (u *BookingUsecase) Pay(ctx context.Context, bookingID, userID string, success bool) (*domain.Booking, error) {
	booking, err := u.bookings.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil || booking.UserID != userID {
		return nil, ErrBookingNotPending
	}
	if booking.Status != domain.BookingPending {
		return nil, ErrBookingNotPending
	}

	if !success {
		u.releaseBookingSeats(ctx, booking)
		if err := u.bookings.UpdateStatus(ctx, bookingID, domain.BookingFailed, nil); err != nil {
			return nil, err
		}
		booking.Status = domain.BookingFailed
		return booking, nil
	}

	for _, label := range booking.SeatLabels {
		held, err := u.lock.IsHeldByToken(ctx, lockKey(booking.ShowtimeID, label), booking.LockToken)
		if err != nil {
			return nil, err
		}
		if !held {
			u.bookings.UpdateStatus(ctx, bookingID, domain.BookingFailed, nil)
			return nil, fmt.Errorf("%w: %s", ErrLockExpired, label)
		}
	}

	for _, label := range booking.SeatLabels {
		if _, err := u.seats.MarkBooked(ctx, booking.ShowtimeID, label, booking.LockToken, booking.ID); err != nil {
			return nil, err
		}
		u.lock.ReleaseIfMatch(ctx, lockKey(booking.ShowtimeID, label), booking.LockToken)
	}

	paidAt := time.Now().UTC()
	if err := u.bookings.UpdateStatus(ctx, bookingID, domain.BookingConfirmed, &paidAt); err != nil {
		return nil, err
	}
	booking.Status = domain.BookingConfirmed
	booking.PaidAt = &paidAt

	_ = u.pub.Publish(ctx, domain.SeatEvent{
		EventID:    uuid.NewString(),
		EventType:  domain.EventBookingConfirmed,
		ShowtimeID: booking.ShowtimeID,
		SeatLabels: booking.SeatLabels,
		UserID:     userID,
		BookingID:  booking.ID,
		OccurredAt: time.Now().UTC(),
	})

	recordAudit(ctx, u.auditLogs, &domain.AuditLog{
		EventType:  domain.AuditBookingSuccess,
		Message:    fmt.Sprintf("booking %s confirmed for user %s (seats %v)", booking.ID, userID, booking.SeatLabels),
		BookingID:  booking.ID,
		UserID:     userID,
		ShowtimeID: booking.ShowtimeID,
		SeatLabels: booking.SeatLabels,
	})

	return booking, nil
}

func (u *BookingUsecase) releaseBookingSeats(ctx context.Context, booking *domain.Booking) {
	for _, label := range booking.SeatLabels {
		u.lock.ReleaseIfMatch(ctx, lockKey(booking.ShowtimeID, label), booking.LockToken)
		u.seats.MarkAvailable(ctx, booking.ShowtimeID, label)
	}
	_ = u.pub.Publish(ctx, domain.SeatEvent{
		EventID:    uuid.NewString(),
		EventType:  domain.EventSeatReleased,
		ShowtimeID: booking.ShowtimeID,
		SeatLabels: booking.SeatLabels,
		OccurredAt: time.Now().UTC(),
	})
}
