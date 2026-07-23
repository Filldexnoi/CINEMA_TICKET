package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

func Seed(ctx context.Context, cinemas ports.CinemaRepository, movies ports.MovieRepository, showtimes ports.ShowtimeRepository, seats ports.ShowtimeSeatRepository) error {
	now := time.Now().UTC()

	cinemaList := []*domain.Cinema{
		{ID: "cinema-downtown", Name: "โรงภาพยนตร์สยาม", City: "กรุงเทพฯ"},
		{ID: "cinema-riverside", Name: "โรงภาพยนตร์เจ้าพระยา", City: "กรุงเทพฯ"},
	}
	for _, c := range cinemaList {
		if err := cinemas.Upsert(ctx, c); err != nil {
			return fmt.Errorf("seed cinema %s: %w", c.ID, err)
		}
	}

	movieList := []*domain.Movie{
		{ID: "movie-galaxy-raiders", Title: "ผีบ้านผีเรือน", Description: "ครอบครัวหนึ่งย้ายเข้าบ้านเก่าแล้วต้องเผชิญกับผีที่แสนกวน", DurationMinutes: 115, Genre: "สยองขวัญ-ตลก", Rating: "น13+"},
		{ID: "movie-the-last-harbor", Title: "นักสู้เลือดมวยไทย", Description: "นักมวยหนุ่มต้องกลับสังเวียนเพื่อกู้ศักดิ์ศรีครอบครัว", DurationMinutes: 105, Genre: "แอคชั่น", Rating: "น13+"},
		{ID: "movie-midnight-in-lumen", Title: "รักนี้ที่เชียงใหม่", Description: "หญิงสาวกรุงเทพฯ ตกหลุมรักหนุ่มชาวเชียงใหม่ระหว่างทริปหนีความวุ่นวาย", DurationMinutes: 125, Genre: "โรแมนติก-ดราม่า", Rating: "ทั่วไป"},
	}
	for _, m := range movieList {
		if err := movies.Upsert(ctx, m); err != nil {
			return fmt.Errorf("seed movie %s: %w", m.ID, err)
		}
	}

	const rows, cols = 8, 10
	basePrice := 150.0

	for _, m := range movieList {
		for i, offset := range []time.Duration{2 * time.Hour, 5 * time.Hour} {
			cinema := cinemaList[i%len(cinemaList)]
			start := now.Add(offset)
			showtimeID := fmt.Sprintf("%s-show-%d", m.ID, i+1)

			st := &domain.Showtime{
				ID:        showtimeID,
				MovieID:   m.ID,
				CinemaID:  cinema.ID,
				HallName:  fmt.Sprintf("โรงที่ %d", i+1),
				StartTime: start,
				EndTime:   start.Add(time.Duration(m.DurationMinutes) * time.Minute),
				Rows:      rows,
				Cols:      cols,
				BasePrice: basePrice,
			}
			if err := showtimes.Upsert(ctx, st); err != nil {
				return fmt.Errorf("seed showtime %s: %w", showtimeID, err)
			}

			if err := seedSeatsIfMissing(ctx, seats, showtimeID, rows, cols, basePrice); err != nil {
				return fmt.Errorf("seed seats for %s: %w", showtimeID, err)
			}
		}
	}

	log.Printf("seed: %d cinemas, %d movies, %d showtimes ready", len(cinemaList), len(movieList), len(movieList)*2)
	return nil
}

func seedSeatsIfMissing(ctx context.Context, seats ports.ShowtimeSeatRepository, showtimeID string, rows, cols int, price float64) error {
	existing, err := seats.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	now := time.Now().UTC()
	for r := 0; r < rows; r++ {
		rowLabel := string(rune('A' + r))
		for col := 1; col <= cols; col++ {
			seatLabel := fmt.Sprintf("%s%d", rowLabel, col)
			seat := &domain.ShowtimeSeat{
				ID:         fmt.Sprintf("%s:%s", showtimeID, seatLabel),
				ShowtimeID: showtimeID,
				SeatLabel:  seatLabel,
				Row:        r,
				Col:        col,
				Status:     domain.SeatAvailable,
				Price:      price,
				UpdatedAt:  now,
			}
			if err := seats.Insert(ctx, seat); err != nil {
				return err
			}
		}
	}
	return nil
}
