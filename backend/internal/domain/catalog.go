package domain

import "time"

type Cinema struct {
	ID   string `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
	City string `bson:"city" json:"city"`
}

type Movie struct {
	ID              string `bson:"_id" json:"id"`
	Title           string `bson:"title" json:"title"`
	Description     string `bson:"description" json:"description"`
	PosterURL       string `bson:"poster_url" json:"poster_url"`
	DurationMinutes int    `bson:"duration_minutes" json:"duration_minutes"`
	Genre           string `bson:"genre" json:"genre"`
	Rating          string `bson:"rating" json:"rating"`
}

type Showtime struct {
	ID        string    `bson:"_id" json:"id"`
	MovieID   string    `bson:"movie_id" json:"movie_id"`
	CinemaID  string    `bson:"cinema_id" json:"cinema_id"`
	HallName  string    `bson:"hall_name" json:"hall_name"`
	StartTime time.Time `bson:"start_time" json:"start_time"`
	EndTime   time.Time `bson:"end_time" json:"end_time"`
	Rows      int       `bson:"rows" json:"rows"`
	Cols      int       `bson:"cols" json:"cols"`
	BasePrice float64   `bson:"base_price" json:"base_price"`
}
