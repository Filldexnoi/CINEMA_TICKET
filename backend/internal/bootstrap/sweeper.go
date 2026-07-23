package bootstrap

import (
	"context"
	"log"
	"time"

	"cinema-ticket/backend/internal/usecase"
	"cinema-ticket/backend/internal/usecase/ports"
)

const sweepInterval = 20 * time.Second

func RunSweeper(ctx context.Context, seatsRepo ports.ShowtimeSeatRepository, seatUsecase *usecase.SeatUsecase) {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			expired, err := seatsRepo.FindExpiredLocked(ctx, time.Now().UTC())
			if err != nil {
				log.Printf("sweeper: query error: %v", err)
				continue
			}
			for _, seat := range expired {
				if err := seatUsecase.ExpireSeat(ctx, seat.ShowtimeID, seat.SeatLabel); err != nil {
					log.Printf("sweeper: expire seat %s/%s: %v", seat.ShowtimeID, seat.SeatLabel, err)
				}
			}
		}
	}
}
