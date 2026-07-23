package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"cinema-ticket/backend/internal/delivery/http/middleware"
	"cinema-ticket/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth           *usecase.AuthUsecase
	frontendOrigin string
}

func NewAuthHandler(auth *usecase.AuthUsecase, frontendOrigin string) *AuthHandler {
	return &AuthHandler{auth: auth, frontendOrigin: frontendOrigin}
}

func (h *AuthHandler) Login(c *gin.Context) {
	state := randomState()
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)
	c.Redirect(http.StatusFound, h.auth.LoginURL(state))
}

func (h *AuthHandler) Callback(c *gin.Context) {
	expectedState, _ := c.Cookie("oauth_state")
	if state := c.Query("state"); state == "" || state != expectedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, _, err := h.auth.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, h.frontendOrigin+"/oauth/callback?token="+token)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.auth.Me(c.Request.Context(), middleware.UserID(c))
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
