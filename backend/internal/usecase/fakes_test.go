package usecase

import (
	"context"
	"sync"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

type fakeSeatRepo struct {
	mu    sync.Mutex
	seats map[string]*domain.ShowtimeSeat
}

func newFakeSeatRepo() *fakeSeatRepo {
	return &fakeSeatRepo{seats: make(map[string]*domain.ShowtimeSeat)}
}

func seatKey(showtimeID, seatLabel string) string { return showtimeID + "|" + seatLabel }

func (r *fakeSeatRepo) Insert(ctx context.Context, seat *domain.ShowtimeSeat) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seats[seatKey(seat.ShowtimeID, seat.SeatLabel)] = seat
	return nil
}

func (r *fakeSeatRepo) ListByShowtime(ctx context.Context, showtimeID string) ([]*domain.ShowtimeSeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ShowtimeSeat
	for _, s := range r.seats {
		if s.ShowtimeID == showtimeID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *fakeSeatRepo) FindOne(ctx context.Context, showtimeID, seatLabel string) (*domain.ShowtimeSeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.seats[seatKey(showtimeID, seatLabel)], nil
}

func (r *fakeSeatRepo) MarkLocked(ctx context.Context, showtimeID, seatLabel, userID, lockToken string, expiresAt, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seat := r.seats[seatKey(showtimeID, seatLabel)]
	if seat == nil {
		return false, nil
	}
	stale := seat.Status == domain.SeatLocked && seat.LockExpiresAt != nil && seat.LockExpiresAt.Before(now)
	if seat.Status != domain.SeatAvailable && !stale {
		return false, nil
	}
	seat.Status = domain.SeatLocked
	seat.LockedBy = userID
	seat.LockToken = lockToken
	exp := expiresAt
	seat.LockExpiresAt = &exp
	seat.UpdatedAt = now
	return true, nil
}

func (r *fakeSeatRepo) MarkAvailable(ctx context.Context, showtimeID, seatLabel string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	seat := r.seats[seatKey(showtimeID, seatLabel)]
	if seat == nil {
		return nil
	}
	seat.Status = domain.SeatAvailable
	seat.LockedBy = ""
	seat.LockToken = ""
	seat.LockExpiresAt = nil
	seat.BookingID = ""
	return nil
}

func (r *fakeSeatRepo) MarkBooked(ctx context.Context, showtimeID, seatLabel, lockToken, bookingID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seat := r.seats[seatKey(showtimeID, seatLabel)]
	if seat == nil || seat.Status != domain.SeatLocked || seat.LockToken != lockToken {
		return false, nil
	}
	seat.Status = domain.SeatBooked
	seat.BookingID = bookingID
	seat.LockedBy = ""
	seat.LockToken = ""
	seat.LockExpiresAt = nil
	return true, nil
}

func (r *fakeSeatRepo) AttachBooking(ctx context.Context, showtimeID, seatLabel, bookingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if seat := r.seats[seatKey(showtimeID, seatLabel)]; seat != nil {
		seat.BookingID = bookingID
	}
	return nil
}

func (r *fakeSeatRepo) FindExpiredLocked(ctx context.Context, now time.Time) ([]*domain.ShowtimeSeat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.ShowtimeSeat
	for _, s := range r.seats {
		if s.Status == domain.SeatLocked && s.LockExpiresAt != nil && s.LockExpiresAt.Before(now) {
			out = append(out, s)
		}
	}
	return out, nil
}

type fakeBookingRepo struct {
	mu       sync.Mutex
	bookings map[string]*domain.Booking
}

func newFakeBookingRepo() *fakeBookingRepo {
	return &fakeBookingRepo{bookings: make(map[string]*domain.Booking)}
}

func (r *fakeBookingRepo) Create(ctx context.Context, b *domain.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bookings[b.ID] = b
	return nil
}

func (r *fakeBookingRepo) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bookings[id], nil
}

func (r *fakeBookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus, paidAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	b := r.bookings[id]
	if b == nil {
		return nil
	}
	b.Status = status
	if paidAt != nil {
		b.PaidAt = paidAt
	}
	return nil
}

func (r *fakeBookingRepo) ListFiltered(ctx context.Context, filter ports.BookingFilter) ([]*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*domain.Booking
	for _, b := range r.bookings {
		out = append(out, b)
	}
	return out, nil
}

type fakeLock struct {
	mu         sync.Mutex
	store      map[string]string
	acquireErr error
}

func newFakeLock() *fakeLock {
	return &fakeLock{store: make(map[string]string)}
}

func (l *fakeLock) Acquire(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.acquireErr != nil {
		return false, l.acquireErr
	}
	if _, exists := l.store[key]; exists {
		return false, nil
	}
	l.store[key] = token
	return true, nil
}

func (l *fakeLock) ReleaseIfMatch(ctx context.Context, key, token string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.store[key] == token {
		delete(l.store, key)
		return true, nil
	}
	return false, nil
}

func (l *fakeLock) IsHeldByToken(ctx context.Context, key, token string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.store[key] == token, nil
}

type fakePublisher struct {
	mu     sync.Mutex
	events []domain.SeatEvent
}

func (p *fakePublisher) Publish(ctx context.Context, event domain.SeatEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
	return nil
}

type fakeAuditLogRepo struct {
	mu   sync.Mutex
	logs []*domain.AuditLog
}

func (r *fakeAuditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs = append(r.logs, log)
	return nil
}

func (r *fakeAuditLogRepo) ListRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.logs, nil
}

type fakeUserRepo struct {
	users map[string]*domain.User
}

func newFakeUserRepo() *fakeUserRepo { return &fakeUserRepo{users: make(map[string]*domain.User)} }

func (r *fakeUserRepo) FindByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	for _, u := range r.users {
		if u.GoogleSub == sub {
			return u, nil
		}
	}
	return nil, nil
}

func (r *fakeUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.users[id], nil
}

func (r *fakeUserRepo) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	r.users[u.ID] = u
	return u, nil
}

type fakeShowtimeRepo struct {
	showtimes map[string]*domain.Showtime
}

func newFakeShowtimeRepo() *fakeShowtimeRepo {
	return &fakeShowtimeRepo{showtimes: make(map[string]*domain.Showtime)}
}

func (r *fakeShowtimeRepo) ListByMovie(ctx context.Context, movieID string) ([]*domain.Showtime, error) {
	var out []*domain.Showtime
	for _, s := range r.showtimes {
		if s.MovieID == movieID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *fakeShowtimeRepo) FindByID(ctx context.Context, id string) (*domain.Showtime, error) {
	return r.showtimes[id], nil
}

func (r *fakeShowtimeRepo) Upsert(ctx context.Context, s *domain.Showtime) error {
	r.showtimes[s.ID] = s
	return nil
}
