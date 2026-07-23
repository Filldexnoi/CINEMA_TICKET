package usecase

import (
	"context"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

type CatalogUsecase struct {
	movies    ports.MovieRepository
	showtimes ports.ShowtimeRepository
}

func NewCatalogUsecase(movies ports.MovieRepository, showtimes ports.ShowtimeRepository) *CatalogUsecase {
	return &CatalogUsecase{movies: movies, showtimes: showtimes}
}

func (u *CatalogUsecase) ListMovies(ctx context.Context) ([]*domain.Movie, error) {
	return u.movies.List(ctx)
}

func (u *CatalogUsecase) GetMovie(ctx context.Context, id string) (*domain.Movie, error) {
	return u.movies.FindByID(ctx, id)
}

func (u *CatalogUsecase) ListShowtimesForMovie(ctx context.Context, movieID string) ([]*domain.Showtime, error) {
	return u.showtimes.ListByMovie(ctx, movieID)
}

func (u *CatalogUsecase) GetShowtime(ctx context.Context, id string) (*domain.Showtime, error) {
	return u.showtimes.FindByID(ctx, id)
}
