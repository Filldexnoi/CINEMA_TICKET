package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "user_id"

type TokenVerifier interface {
	Verify(tokenString string) (userID string, err error)
}

func Auth(verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		userID, err := verifier.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	v, _ := c.Get(ContextUserIDKey)
	userID, _ := v.(string)
	return userID
}
