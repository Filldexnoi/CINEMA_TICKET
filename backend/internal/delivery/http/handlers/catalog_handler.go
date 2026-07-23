package handlers

import (
	"net/http"

	"cinema-ticket/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CatalogHandler struct {
	catalog *usecase.CatalogUsecase
}

func NewCatalogHandler(catalog *usecase.CatalogUsecase) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

func (h *CatalogHandler) ListMovies(c *gin.Context) {
	movies, err := h.catalog.ListMovies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func (h *CatalogHandler) GetMovie(c *gin.Context) {
	movie, err := h.catalog.GetMovie(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if movie == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}
	c.JSON(http.StatusOK, movie)
}

func (h *CatalogHandler) ListShowtimes(c *gin.Context) {
	showtimes, err := h.catalog.ListShowtimesForMovie(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, showtimes)
}

func (h *CatalogHandler) GetShowtime(c *gin.Context) {
	showtime, err := h.catalog.GetShowtime(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if showtime == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "showtime not found"})
		return
	}
	c.JSON(http.StatusOK, showtime)
}
