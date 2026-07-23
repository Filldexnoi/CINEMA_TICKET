package usecase

import "errors"

var (
	ErrSeatUnavailable   = errors.New("seat is not available")
	ErrNotSeatOwner      = errors.New("seat is not locked by this user")
	ErrLockExpired       = errors.New("seat hold has expired")
	ErrBookingNotPending = errors.New("booking is not pending payment")
)
