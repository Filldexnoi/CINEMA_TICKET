package usecase

import (
	"context"
	"fmt"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/google/uuid"
)

type SeatUsecase struct {
	seats     ports.ShowtimeSeatRepository
	bookings  ports.BookingRepository
	lock      ports.DistributedLock
	publisher ports.EventPublisher
	auditLogs ports.AuditLogRepository
	lockTTL   time.Duration
}

func NewSeatUsecase(
	seats ports.ShowtimeSeatRepository,
	bookings ports.BookingRepository,
	lock ports.DistributedLock,
	publisher ports.EventPublisher,
	lockTTL time.Duration,
	auditLogs ports.AuditLogRepository,
) *SeatUsecase {
	return &SeatUsecase{seats: seats, bookings: bookings, lock: lock, publisher: publisher, lockTTL: lockTTL, auditLogs: auditLogs}
}

func lockKey(showtimeID, seatLabel string) string {
	return fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatLabel)
}

func (u *SeatUsecase) GetSeatMapSnapshot(ctx context.Context, showtimeID string) ([]*domain.ShowtimeSeat, error) {
	seats, err := u.seats.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, s := range seats {
		if s.IsStaleLock(now) {
			if err := u.expireLockedSeat(ctx, showtimeID, s.SeatLabel, s.BookingID); err != nil {
				return nil, err
			}
			s.Status = domain.SeatAvailable
			s.LockedBy = ""
			s.LockExpiresAt = nil
			s.BookingID = ""
		}
	}
	return seats, nil
}

func (u *SeatUsecase) ExpireSeat(ctx context.Context, showtimeID, seatLabel string) error {
	seat, err := u.seats.FindOne(ctx, showtimeID, seatLabel)
	if err != nil {
		return err
	}
	if seat == nil || seat.Status != domain.SeatLocked {
		return nil
	}
	return u.expireLockedSeat(ctx, showtimeID, seatLabel, seat.BookingID)
}

func (u *SeatUsecase) expireLockedSeat(ctx context.Context, showtimeID, seatLabel, bookingID string) error {
	if err := u.seats.MarkAvailable(ctx, showtimeID, seatLabel); err != nil {
		return err
	}

	if bookingID != "" {
		booking, err := u.bookings.FindByID(ctx, bookingID)
		if err != nil {
			return err
		}
		if booking != nil && booking.Status == domain.BookingPending {
			if err := u.bookings.UpdateStatus(ctx, bookingID, domain.BookingExpired, nil); err != nil {
				return err
			}
			u.publish(ctx, domain.EventBookingExpired, showtimeID, []string{seatLabel}, booking.UserID, bookingID)
			recordAudit(ctx, u.auditLogs, &domain.AuditLog{
				EventType:  domain.AuditBookingTimeout,
				Message:    fmt.Sprintf("booking %s timed out unpaid (seat %s)", bookingID, seatLabel),
				BookingID:  bookingID,
				UserID:     booking.UserID,
				ShowtimeID: showtimeID,
				SeatLabels: []string{seatLabel},
			})
			return nil
		}
	}

	u.publishReleased(ctx, showtimeID, []string{seatLabel})
	return nil
}

func (u *SeatUsecase) LockSeats(ctx context.Context, showtimeID string, seatLabels []string, userID string) (token string, err error) {
	token = userID + ":" + uuid.NewString()
	now := time.Now().UTC()
	expiresAt := now.Add(u.lockTTL)

	var acquired []string
	rollback := func() {
		for _, seat := range acquired {
			u.lock.ReleaseIfMatch(ctx, lockKey(showtimeID, seat), token)
			u.seats.MarkAvailable(ctx, showtimeID, seat)
		}
		if len(acquired) > 0 {
			u.publishReleased(ctx, showtimeID, acquired)
		}
	}

	for _, seat := range seatLabels {
		ok, lockErr := u.lock.Acquire(ctx, lockKey(showtimeID, seat), token, u.lockTTL)
		if lockErr != nil {
			rollback()
			recordAudit(ctx, u.auditLogs, &domain.AuditLog{
				EventType:  domain.AuditSystemError,
				Message:    fmt.Sprintf("redis lock acquire failed for seat %s: %v", seat, lockErr),
				UserID:     userID,
				ShowtimeID: showtimeID,
				SeatLabels: []string{seat},
			})
			return "", lockErr
		}
		if !ok {
			rollback()
			return "", fmt.Errorf("%w: %s", ErrSeatUnavailable, seat)
		}

		marked, markErr := u.seats.MarkLocked(ctx, showtimeID, seat, userID, token, expiresAt, now)
		if markErr != nil {
			u.lock.ReleaseIfMatch(ctx, lockKey(showtimeID, seat), token)
			rollback()
			return "", markErr
		}
		if !marked {
			u.lock.ReleaseIfMatch(ctx, lockKey(showtimeID, seat), token)
			rollback()
			return "", fmt.Errorf("%w: %s", ErrSeatUnavailable, seat)
		}

		acquired = append(acquired, seat)
	}

	u.publish(ctx, domain.EventSeatLocked, showtimeID, seatLabels, userID, "")
	return token, nil
}

func (u *SeatUsecase) UnlockSeats(ctx context.Context, showtimeID string, seatLabels []string, token string) error {
	var released []string
	for _, seat := range seatLabels {
		ok, err := u.lock.ReleaseIfMatch(ctx, lockKey(showtimeID, seat), token)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := u.seats.MarkAvailable(ctx, showtimeID, seat); err != nil {
			return err
		}
		released = append(released, seat)
	}
	if len(released) > 0 {
		u.publishReleased(ctx, showtimeID, released)
	}
	return nil
}

func (u *SeatUsecase) publish(ctx context.Context, eventType domain.EventType, showtimeID string, seatLabels []string, userID, bookingID string) {
	_ = u.publisher.Publish(ctx, domain.SeatEvent{
		EventID:    uuid.NewString(),
		EventType:  eventType,
		ShowtimeID: showtimeID,
		SeatLabels: seatLabels,
		UserID:     userID,
		BookingID:  bookingID,
		OccurredAt: time.Now().UTC(),
	})
}

func (u *SeatUsecase) publishReleased(ctx context.Context, showtimeID string, seatLabels []string) {
	u.publish(ctx, domain.EventSeatReleased, showtimeID, seatLabels, "", "")
	recordAudit(ctx, u.auditLogs, &domain.AuditLog{
		EventType:  domain.AuditSeatReleased,
		Message:    fmt.Sprintf("seat(s) %v released back to available", seatLabels),
		ShowtimeID: showtimeID,
		SeatLabels: seatLabels,
	})
}
