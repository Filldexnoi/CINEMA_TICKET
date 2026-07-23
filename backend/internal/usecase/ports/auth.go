package ports

import "context"

type GoogleProfile struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

type GoogleAuthClient interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (*GoogleProfile, error)
}

type SessionIssuer interface {
	Issue(userID string) (string, error)
	Verify(tokenString string) (userID string, err error)
}
