package handlers

import (
	"errors"
	"net/http"

	"cinema-ticket/backend/internal/delivery/http/middleware"
	"cinema-ticket/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookings *usecase.BookingUsecase
}

func NewBookingHandler(bookings *usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

type createBookingRequest struct {
	ShowtimeID string   `json:"showtime_id" binding:"required"`
	SeatLabels []string `json:"seat_labels" binding:"required,min=1"`
	LockToken  string   `json:"lock_token" binding:"required"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookings.CreateBooking(c.Request.Context(), middleware.UserID(c), req.ShowtimeID, req.SeatLabels, req.LockToken)
	if err != nil {
		if errors.Is(err, usecase.ErrLockExpired) || errors.Is(err, usecase.ErrNotSeatOwner) {
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, booking)
}

func (h *BookingHandler) Get(c *gin.Context) {
	booking, err := h.bookings.GetBooking(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}
	c.JSON(http.StatusOK, booking)
}

type payRequest struct {
	Result string `json:"result"`
}

func (h *BookingHandler) Pay(c *gin.Context) {
	var req payRequest
	_ = c.ShouldBindJSON(&req)
	success := req.Result != "fail"

	booking, err := h.bookings.Pay(c.Request.Context(), c.Param("id"), middleware.UserID(c), success)
	if err != nil {
		if errors.Is(err, usecase.ErrLockExpired) {
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, usecase.ErrBookingNotPending) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, booking)
}
