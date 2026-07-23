package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"cinema-ticket/backend/internal/domain"
)

func TestCreateBooking_FailsWhenLockNotHeld(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok-a", Price: 150})

	uc := NewBookingUsecase(newFakeBookingRepo(), seats, newFakeLock(), &fakePublisher{}, newFakeShowtimeRepo(), newFakeUserRepo(), &fakeAuditLogRepo{})

	if _, err := uc.CreateBooking(ctx, "user-a", "st1", []string{"A1"}, "tok-a"); !errors.Is(err, ErrLockExpired) {
		t.Fatalf("expected ErrLockExpired, got: %v", err)
	}
}

func TestCreateBooking_Success_DenormalizesMovieShowtimeAndUser(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	expiresAt := time.Now().Add(5 * time.Minute)
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok-a", LockExpiresAt: &expiresAt, Price: 150})
	lock := newFakeLock()
	lock.store[lockKey("st1", "A1")] = "tok-a"

	showtimes := newFakeShowtimeRepo()
	showtimes.showtimes["st1"] = &domain.Showtime{ID: "st1", MovieID: "movie-x", StartTime: time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)}
	users := newFakeUserRepo()
	users.users["user-a"] = &domain.User{ID: "user-a", Email: "a@test.local"}

	uc := NewBookingUsecase(newFakeBookingRepo(), seats, lock, &fakePublisher{}, showtimes, users, &fakeAuditLogRepo{})

	booking, err := uc.CreateBooking(ctx, "user-a", "st1", []string{"A1"}, "tok-a")
	if err != nil {
		t.Fatalf("CreateBooking failed: %v", err)
	}
	if booking.MovieID != "movie-x" {
		t.Errorf("expected movie_id denormalized to movie-x, got %q", booking.MovieID)
	}
	if booking.ShowtimeDate != "2026-08-01" {
		t.Errorf("expected showtime_date 2026-08-01, got %q", booking.ShowtimeDate)
	}
	if booking.UserEmail != "a@test.local" {
		t.Errorf("expected user_email denormalized, got %q", booking.UserEmail)
	}
	if booking.Status != domain.BookingPending {
		t.Errorf("expected status PENDING, got %s", booking.Status)
	}
	if booking.TotalAmount != 150 {
		t.Errorf("expected total_amount 150, got %v", booking.TotalAmount)
	}
}

func TestPay_Success_ConfirmsBookingAndSeatAndRecordsAudit(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	expiresAt := time.Now().Add(5 * time.Minute)
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok-a", LockExpiresAt: &expiresAt, Price: 150})
	bookings := newFakeBookingRepo()
	bookings.Create(ctx, &domain.Booking{ID: "bk1", UserID: "user-a", ShowtimeID: "st1", SeatLabels: []string{"A1"}, LockToken: "tok-a", Status: domain.BookingPending})
	lock := newFakeLock()
	lock.store[lockKey("st1", "A1")] = "tok-a"
	audit := &fakeAuditLogRepo{}

	uc := NewBookingUsecase(bookings, seats, lock, &fakePublisher{}, newFakeShowtimeRepo(), newFakeUserRepo(), audit)

	booking, err := uc.Pay(ctx, "bk1", "user-a", true)
	if err != nil {
		t.Fatalf("Pay failed: %v", err)
	}
	if booking.Status != domain.BookingConfirmed {
		t.Errorf("expected CONFIRMED, got %s", booking.Status)
	}

	seat, _ := seats.FindOne(ctx, "st1", "A1")
	if seat.Status != domain.SeatBooked {
		t.Errorf("expected seat BOOKED, got %s", seat.Status)
	}

	if held, _ := lock.IsHeldByToken(ctx, lockKey("st1", "A1"), "tok-a"); held {
		t.Error("expected the Redis lock key to be released after successful payment")
	}

	if len(audit.logs) != 1 || audit.logs[0].EventType != domain.AuditBookingSuccess {
		t.Fatalf("expected one BOOKING_SUCCESS audit entry, got %+v", audit.logs)
	}
}

func TestPay_Fail_ReleasesSeatsAndMarksBookingFailed(t *testing.T) {
	ctx := context.Background()
	seats := newFakeSeatRepo()
	expiresAt := time.Now().Add(5 * time.Minute)
	seats.Insert(ctx, &domain.ShowtimeSeat{ShowtimeID: "st1", SeatLabel: "A1", Status: domain.SeatLocked, LockedBy: "user-a", LockToken: "tok-a", LockExpiresAt: &expiresAt, Price: 150})
	bookings := newFakeBookingRepo()
	bookings.Create(ctx, &domain.Booking{ID: "bk1", UserID: "user-a", ShowtimeID: "st1", SeatLabels: []string{"A1"}, LockToken: "tok-a", Status: domain.BookingPending})
	lock := newFakeLock()
	lock.store[lockKey("st1", "A1")] = "tok-a"

	uc := NewBookingUsecase(bookings, seats, lock, &fakePublisher{}, newFakeShowtimeRepo(), newFakeUserRepo(), &fakeAuditLogRepo{})

	booking, err := uc.Pay(ctx, "bk1", "user-a", false)
	if err != nil {
		t.Fatalf("Pay(fail) should not itself error, got: %v", err)
	}
	if booking.Status != domain.BookingFailed {
		t.Errorf("expected FAILED, got %s", booking.Status)
	}

	seat, _ := seats.FindOne(ctx, "st1", "A1")
	if seat.Status != domain.SeatAvailable {
		t.Errorf("expected seat released back to AVAILABLE, got %s", seat.Status)
	}
}

func TestPay_RejectsWrongUser(t *testing.T) {
	ctx := context.Background()
	bookings := newFakeBookingRepo()
	bookings.Create(ctx, &domain.Booking{ID: "bk1", UserID: "user-a", Status: domain.BookingPending})
	uc := NewBookingUsecase(bookings, newFakeSeatRepo(), newFakeLock(), &fakePublisher{}, newFakeShowtimeRepo(), newFakeUserRepo(), &fakeAuditLogRepo{})

	if _, err := uc.Pay(ctx, "bk1", "user-b", true); !errors.Is(err, ErrBookingNotPending) {
		t.Fatalf("expected ErrBookingNotPending for a different user, got: %v", err)
	}
}
