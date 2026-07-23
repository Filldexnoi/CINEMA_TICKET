package middleware

import (
	"context"
	"net/http"

	"cinema-ticket/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserRoleLookup interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

func AdminOnly(users UserRoleLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := users.FindByID(c.Request.Context(), UserID(c))
		if err != nil || user == nil || user.Role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
