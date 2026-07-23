package ports

import (
	"context"

	"cinema-ticket/backend/internal/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.SeatEvent) error
}

type Broadcaster interface {
	BroadcastToShowtime(showtimeID string, payload []byte)
}
