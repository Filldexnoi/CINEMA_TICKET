package http

import (
	"context"
	"net/http"
	"time"

	"cinema-ticket/backend/internal/delivery/http/handlers"
	"cinema-ticket/backend/internal/delivery/http/middleware"
	"cinema-ticket/backend/internal/delivery/ws"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Deps struct {
	Auth           *handlers.AuthHandler
	Catalog        *handlers.CatalogHandler
	Seat           *handlers.SeatHandler
	Booking        *handlers.BookingHandler
	Admin          *handlers.AdminHandler
	Verifier       middleware.TokenVerifier
	Users          middleware.UserRoleLookup
	Hub            *ws.Hub
	WSVerifier     ws.TokenVerifier
	FrontendOrigin string
	Mongo          Pinger
	Redis          Pinger
	Kafka          Pinger
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(d.FrontendOrigin))

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/health/ready", readinessHandler(d))

	auth := r.Group("/auth")
	{
		auth.GET("/google/login", d.Auth.Login)
		auth.GET("/google/callback", d.Auth.Callback)
	}

	api := r.Group("/api")
	{
		api.GET("/movies", d.Catalog.ListMovies)
		api.GET("/movies/:id", d.Catalog.GetMovie)
		api.GET("/movies/:id/showtimes", d.Catalog.ListShowtimes)
		api.GET("/showtimes/:id", d.Catalog.GetShowtime)
		api.GET("/showtimes/:id/seats", d.Seat.GetSeatMap)

		authed := api.Group("")
		authed.Use(middleware.Auth(d.Verifier))
		{
			authed.GET("/me", d.Auth.Me)
			authed.POST("/showtimes/:id/seats/lock", d.Seat.Lock)
			authed.POST("/showtimes/:id/seats/unlock", d.Seat.Unlock)
			authed.POST("/bookings", d.Booking.Create)
			authed.GET("/bookings/:id", d.Booking.Get)
			authed.POST("/bookings/:id/pay", d.Booking.Pay)

			admin := authed.Group("/admin")
			admin.Use(middleware.AdminOnly(d.Users))
			{
				admin.GET("/bookings", d.Admin.ListBookings)
				admin.GET("/audit-logs", d.Admin.ListAuditLogs)
			}
		}
	}

	r.GET("/ws/showtimes/:id", ws.Handler(d.Hub, d.WSVerifier))

	return r
}

func readinessHandler(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}
		ready := true

		ping := func(name string, p Pinger) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()
			if err := p.Ping(ctx); err != nil {
				checks[name] = err.Error()
				ready = false
			} else {
				checks[name] = "ok"
			}
		}

		ping("mongo", d.Mongo)
		ping("redis", d.Redis)
		ping("kafka", d.Kafka)

		status := http.StatusOK
		statusText := "ready"
		if !ready {
			status = http.StatusServiceUnavailable
			statusText = "not ready"
		}
		c.JSON(status, gin.H{"status": statusText, "checks": checks})
	}
}
