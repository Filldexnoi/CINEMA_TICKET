package handlers

import (
	"errors"
	"net/http"

	"cinema-ticket/backend/internal/delivery/http/middleware"
	"cinema-ticket/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type SeatHandler struct {
	seats *usecase.SeatUsecase
}

func NewSeatHandler(seats *usecase.SeatUsecase) *SeatHandler {
	return &SeatHandler{seats: seats}
}

func (h *SeatHandler) GetSeatMap(c *gin.Context) {
	seats, err := h.seats.GetSeatMapSnapshot(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, seats)
}

type lockRequest struct {
	SeatLabels []string `json:"seat_labels" binding:"required,min=1"`
}

func (h *SeatHandler) Lock(c *gin.Context) {
	var req lockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.seats.LockSeats(c.Request.Context(), c.Param("id"), req.SeatLabels, middleware.UserID(c))
	if err != nil {
		if errors.Is(err, usecase.ErrSeatUnavailable) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lock_token": token})
}

type unlockRequest struct {
	SeatLabels []string `json:"seat_labels" binding:"required,min=1"`
	LockToken  string   `json:"lock_token" binding:"required"`
}

func (h *SeatHandler) Unlock(c *gin.Context) {
	var req unlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.seats.UnlockSeats(c.Request.Context(), c.Param("id"), req.SeatLabels, req.LockToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "released"})
}
