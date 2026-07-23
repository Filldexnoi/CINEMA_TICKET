package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"cinema-ticket/backend/internal/domain"
)

func TestLockSeats_SecondUserBlockedWhenSeatAlreadyLocked(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatAvailable, Price: 150})
	uc := NewSeatUsecase(seats, newFakeBookingRepo(), newFakeLock(), &fakePublisher{}, 5*time.Minute, &fakeAuditLogRepo{})

	tokenA, err := uc.LockSeats(ctx, "st1", []string{"A1"}, "user-a")
	if err != nil {
		t.Fatalf("user A should have locked the seat, got error: %v", err)
	}
	if tokenA == "" {
		t.Fatal("expected a non-empty lock token")
	}

	if _, err := uc.LockSeats(ctx, "st1", []string{"A1"}, "user-b"); !errors.Is(err, ErrSeatUnavailable) {
		t.Fatalf("expected ErrSeatUnavailable for user B, got: %v", err)
	}

	seat, _ := seats.FindOne(ctx, "st1", "A1")
	if seat.Status != domain.SeatLocked || seat.LockedBy != "user-a" {
		t.Fatalf("expected seat to still be locked by user-a, got status=%s lockedBy=%s", seat.Status, seat.LockedBy)
	}
}

func TestLockSeats_RollsBackAllOnPartialFailure(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatAvailable, Price: 150})
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A2", Status: domain.SeatBooked, Price: 150})
	uc := NewSeatUsecase(seats, newFakeBookingRepo(), newFakeLock(), &fakePublisher{}, 5*time.Minute, &fakeAuditLogRepo{})

	if _, err := uc.LockSeats(ctx, "st1", []string{"A1", "A2"}, "user-a"); !errors.Is(err, ErrSeatUnavailable) {
		t.Fatalf("expected ErrSeatUnavailable, got: %v", err)
	}

	seatA1, _ := seats.FindOne(ctx, "st1", "A1")
	if seatA1.Status != domain.SeatAvailable {
		t.Fatalf("expected A1 rolled back to AVAILABLE since A2 failed, got %s", seatA1.Status)
	}
}

func TestExpireSeat_ReleasesLockBackToAvailable(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	past := time.Now().Add(-1 * time.Minute)
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok", LockExpiresAt: &past, Price: 150})
	audit := &fakeAuditLogRepo{}
	uc := NewSeatUsecase(seats, newFakeBookingRepo(), newFakeLock(), &fakePublisher{}, 5*time.Minute, audit)

	if err := uc.ExpireSeat(ctx, "st1", "A1"); err != nil {
		t.Fatalf("ExpireSeat failed: %v", err)
	}

	seat, _ := seats.FindOne(ctx, "st1", "A1")
	if seat.Status != domain.SeatAvailable {
		t.Fatalf("expected seat AVAILABLE after expiry, got %s", seat.Status)
	}
	if len(audit.logs) != 1 || audit.logs[0].EventType != domain.AuditSeatReleased {
		t.Fatalf("expected one SEAT_RELEASED audit entry, got %+v", audit.logs)
	}
}

func TestExpireSeat_ExpiresThePendingBookingToo(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	past := time.Now().Add(-1 * time.Minute)
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok", LockExpiresAt: &past, BookingID: "bk1", Price: 150})
	bookings := newFakeBookingRepo()
	bookings.Create(ctx, &domain.Booking{ID: "bk1", UserID: "user-a", ShowtimeID: "st1", SeatLabels: []string{"A1"}, Status: domain.BookingPending})
	audit := &fakeAuditLogRepo{}
	uc := NewSeatUsecase(seats, bookings, newFakeLock(), &fakePublisher{}, 5*time.Minute, audit)

	if err := uc.ExpireSeat(ctx, "st1", "A1"); err != nil {
		t.Fatalf("ExpireSeat failed: %v", err)
	}

	booking, _ := bookings.FindByID(ctx, "bk1")
	if booking.Status != domain.BookingExpired {
		t.Fatalf("expected booking EXPIRED, got %s", booking.Status)
	}
	if len(audit.logs) != 1 || audit.logs[0].EventType != domain.AuditBookingTimeout {
		t.Fatalf("expected one BOOKING_TIMEOUT audit entry, got %+v", audit.logs)
	}
}

func TestLockSeats_RecordsSystemErrorAuditOnGenuineLockFailure(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatAvailable, Price: 150})
	lock := newFakeLock()
	lock.acquireErr = errors.New("redis: connection refused")
	audit := &fakeAuditLogRepo{}
	uc := NewSeatUsecase(seats, newFakeBookingRepo(), lock, &fakePublisher{}, 5*time.Minute, audit)

	if _, err := uc.LockSeats(ctx, "st1", []string{"A1"}, "user-a"); err == nil {
		t.Fatal("expected an error when the lock backend itself fails")
	}
	if len(audit.logs) != 1 || audit.logs[0].EventType != domain.AuditSystemError {
		t.Fatalf("expected one SYSTEM_ERROR audit entry, got %+v", audit.logs)
	}
}
