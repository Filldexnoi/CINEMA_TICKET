package handlers

import (
	"net/http"

	"cinema-ticket/backend/internal/usecase"
	"cinema-ticket/backend/internal/usecase/ports"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	admin *usecase.AdminUsecase
}

func NewAdminHandler(admin *usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{admin: admin}
}

func (h *AdminHandler) ListBookings(c *gin.Context) {
	filter := ports.BookingFilter{
		MovieID:      c.Query("movie_id"),
		ShowtimeDate: c.Query("date"),
		UserEmail:    c.Query("user_email"),
	}
	bookings, err := h.admin.ListBookings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	logs, err := h.admin.ListAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
