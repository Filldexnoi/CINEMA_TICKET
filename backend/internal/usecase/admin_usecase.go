package usecase

import (
	"context"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

const auditLogPageSize = 200

type AdminUsecase struct {
	bookings  ports.BookingRepository
	auditLogs ports.AuditLogRepository
}

func NewAdminUsecase(bookings ports.BookingRepository, auditLogs ports.AuditLogRepository) *AdminUsecase {
	return &AdminUsecase{bookings: bookings, auditLogs: auditLogs}
}

func (u *AdminUsecase) ListBookings(ctx context.Context, filter ports.BookingFilter) ([]*domain.Booking, error) {
	return u.bookings.ListFiltered(ctx, filter)
}

func (u *AdminUsecase) ListAuditLogs(ctx context.Context) ([]*domain.AuditLog, error) {
	return u.auditLogs.ListRecent(ctx, auditLogPageSize)
}
