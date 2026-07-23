package domain

import "time"

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID         string    `bson:"_id,omitempty" json:"id"`
	GoogleSub  string    `bson:"google_sub" json:"-"`
	Email      string    `bson:"email" json:"email"`
	Name       string    `bson:"name" json:"name"`
	PictureURL string    `bson:"picture_url" json:"picture_url"`
	Role       Role      `bson:"role" json:"role"`
	CreatedAt  time.Time `bson:"created_at" json:"created_at"`
}
