package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"cinema-ticket/backend/internal/domain"

	"github.com/gin-gonic/gin"
)

type fakeUserRoleLookup struct {
	users map[string]*domain.User
}

func (f *fakeUserRoleLookup) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return f.users[id], nil
}

func withFakeUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

func newAdminTestRouter(users *fakeUserRoleLookup, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin-only", withFakeUserID(userID), AdminOnly(users), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestAdminOnly_RejectsNonAdminUser(t *testing.T) {
	users := &fakeUserRoleLookup{users: map[string]*domain.User{
		"user-1": {ID: "user-1", Role: domain.RoleUser},
	}}
	router := newAdminTestRouter(users, "user-1")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin-only", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a non-admin user, got %d", rec.Code)
	}
}

func TestAdminOnly_AllowsAdminUser(t *testing.T) {
	users := &fakeUserRoleLookup{users: map[string]*domain.User{
		"user-2": {ID: "user-2", Role: domain.RoleAdmin},
	}}
	router := newAdminTestRouter(users, "user-2")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin-only", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for an admin user, got %d", rec.Code)
	}
}

func TestAdminOnly_RejectsUnknownUser(t *testing.T) {
	users := &fakeUserRoleLookup{users: map[string]*domain.User{}}
	router := newAdminTestRouter(users, "ghost-user")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin-only", nil))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for an unknown user id, got %d", rec.Code)
	}
}
