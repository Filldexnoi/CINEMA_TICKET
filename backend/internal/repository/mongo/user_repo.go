package mongo

import (
	"context"
	"errors"
	"time"

	"cinema-ticket/backend/internal/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection("users")}
}

func (r *UserRepo) FindByGoogleSub(ctx context.Context, googleSub string) (*domain.User, error) {
	var u domain.User
	err := r.col.FindOne(ctx, bson.M{"google_sub": googleSub}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Upsert(ctx context.Context, user *domain.User) (*domain.User, error) {
	existing, err := r.FindByGoogleSub(ctx, user.GoogleSub)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		user.ID = existing.ID
		user.CreatedAt = existing.CreatedAt

		user.Role = existing.Role
		_, err := r.col.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": bson.M{
			"email":       user.Email,
			"name":        user.Name,
			"picture_url": user.PictureURL,
		}})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	user.ID = uuid.NewString()
	user.CreatedAt = time.Now().UTC()
	user.Role = domain.RoleUser
	if _, err := r.col.InsertOne(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
