package mongo

import (
	"context"
	"errors"

	"cinema-ticket/backend/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MovieRepo struct{ col *mongo.Collection }

func NewMovieRepo(db *mongo.Database) *MovieRepo { return &MovieRepo{col: db.Collection("movies")} }

func (r *MovieRepo) List(ctx context.Context) ([]*domain.Movie, error) {
	cur, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var movies []*domain.Movie
	if err := cur.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepo) FindByID(ctx context.Context, id string) (*domain.Movie, error) {
	var m domain.Movie
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MovieRepo) Upsert(ctx context.Context, movie *domain.Movie) error {
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": movie.ID}, movie, options.Replace().SetUpsert(true))
	return err
}

type CinemaRepo struct{ col *mongo.Collection }

func NewCinemaRepo(db *mongo.Database) *CinemaRepo {
	return &CinemaRepo{col: db.Collection("cinemas")}
}

func (r *CinemaRepo) FindByID(ctx context.Context, id string) (*domain.Cinema, error) {
	var c domain.Cinema
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CinemaRepo) Upsert(ctx context.Context, cinema *domain.Cinema) error {
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": cinema.ID}, cinema, options.Replace().SetUpsert(true))
	return err
}

type ShowtimeRepo struct{ col *mongo.Collection }

func NewShowtimeRepo(db *mongo.Database) *ShowtimeRepo {
	return &ShowtimeRepo{col: db.Collection("showtimes")}
}

func (r *ShowtimeRepo) ListByMovie(ctx context.Context, movieID string) ([]*domain.Showtime, error) {
	cur, err := r.col.Find(ctx, bson.M{"movie_id": movieID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var showtimes []*domain.Showtime
	if err := cur.All(ctx, &showtimes); err != nil {
		return nil, err
	}
	return showtimes, nil
}

func (r *ShowtimeRepo) FindByID(ctx context.Context, id string) (*domain.Showtime, error) {
	var s domain.Showtime
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShowtimeRepo) Upsert(ctx context.Context, showtime *domain.Showtime) error {
	_, err := r.col.ReplaceOne(ctx, bson.M{"_id": showtime.ID}, showtime, options.Replace().SetUpsert(true))
	return err
}
