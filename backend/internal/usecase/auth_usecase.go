package usecase

import (
	"context"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

type AuthUsecase struct {
	google ports.GoogleAuthClient
	users  ports.UserRepository
	jwt    ports.SessionIssuer
}

func NewAuthUsecase(google ports.GoogleAuthClient, users ports.UserRepository, jwt ports.SessionIssuer) *AuthUsecase {
	return &AuthUsecase{google: google, users: users, jwt: jwt}
}

func (u *AuthUsecase) LoginURL(state string) string {
	return u.google.AuthCodeURL(state)
}

func (u *AuthUsecase) HandleCallback(ctx context.Context, code string) (jwtToken string, user *domain.User, err error) {
	profile, err := u.google.Exchange(ctx, code)
	if err != nil {
		return "", nil, err
	}

	user, err = u.users.Upsert(ctx, &domain.User{
		GoogleSub:  profile.Sub,
		Email:      profile.Email,
		Name:       profile.Name,
		PictureURL: profile.Picture,
	})
	if err != nil {
		return "", nil, err
	}

	jwtToken, err = u.jwt.Issue(user.ID)
	if err != nil {
		return "", nil, err
	}
	return jwtToken, user, nil
}

func (u *AuthUsecase) Me(ctx context.Context, userID string) (*domain.User, error) {
	return u.users.FindByID(ctx, userID)
}
