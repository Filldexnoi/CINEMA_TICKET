package jwt

import (
	"errors"
	"time"

	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/golang-jwt/jwt/v5"
)

const sessionTTL = 7 * 24 * time.Hour

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
}

func NewIssuer(secret string) *Issuer {
	return &Issuer{secret: []byte(secret)}
}

var _ ports.SessionIssuer = (*Issuer)(nil)

func (i *Issuer) Issue(userID string) (string, error) {
	now := time.Now()
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(sessionTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(i.secret)
}

func (i *Issuer) Verify(tokenString string) (string, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}
	return c.UserID, nil
}
