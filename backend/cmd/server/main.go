package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"cinema-ticket/backend/internal/bootstrap"
	"cinema-ticket/backend/internal/config"
	httpdelivery "cinema-ticket/backend/internal/delivery/http"
	"cinema-ticket/backend/internal/delivery/http/handlers"
	"cinema-ticket/backend/internal/delivery/ws"
	"cinema-ticket/backend/internal/infra/jwt"
	"cinema-ticket/backend/internal/infra/kafka"
	"cinema-ticket/backend/internal/infra/oauth"
	infraredis "cinema-ticket/backend/internal/infra/redis"
	mongorepo "cinema-ticket/backend/internal/repository/mongo"
	"cinema-ticket/backend/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	db, err := mongorepo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := mongorepo.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("mongo ensure indexes: %v", err)
	}

	redisClient, err := infraredis.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}

	userRepo := mongorepo.NewUserRepo(db)
	movieRepo := mongorepo.NewMovieRepo(db)
	cinemaRepo := mongorepo.NewCinemaRepo(db)
	showtimeRepo := mongorepo.NewShowtimeRepo(db)
	seatRepo := mongorepo.NewShowtimeSeatRepo(db)
	bookingRepo := mongorepo.NewBookingRepo(db)
	auditLogRepo := mongorepo.NewAuditLogRepo(db)

	if err := bootstrap.Seed(ctx, cinemaRepo, movieRepo, showtimeRepo, seatRepo); err != nil {
		log.Fatalf("seed: %v", err)
	}

	lock := infraredis.NewLock(redisClient)
	producer := kafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()
	googleAuth := oauth.NewGoogleAuth(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL)
	jwtIssuer := jwt.NewIssuer(cfg.JWTSecret)
	hub := ws.NewHub()

	authUsecase := usecase.NewAuthUsecase(googleAuth, userRepo, jwtIssuer)
	catalogUsecase := usecase.NewCatalogUsecase(movieRepo, showtimeRepo)
	seatUsecase := usecase.NewSeatUsecase(seatRepo, bookingRepo, lock, producer, cfg.LockTTL, auditLogRepo)
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, seatRepo, lock, producer, showtimeRepo, userRepo, auditLogRepo)
	adminUsecase := usecase.NewAdminUsecase(bookingRepo, auditLogRepo)

	go infraredis.ListenForExpiry(ctx, redisClient, func(showtimeID, seatLabel string) {
		if err := seatUsecase.ExpireSeat(ctx, showtimeID, seatLabel); err != nil {
			log.Printf("expiry listener: %v", err)
		}
	})
	go bootstrap.RunSweeper(ctx, seatRepo, seatUsecase)
	go kafka.RunConsumer(ctx, cfg.KafkaBrokers, hub)
	go kafka.RunNotificationConsumer(ctx, cfg.KafkaBrokers)

	router := httpdelivery.NewRouter(httpdelivery.Deps{
		Auth:           handlers.NewAuthHandler(authUsecase, cfg.FrontendOrigin),
		Catalog:        handlers.NewCatalogHandler(catalogUsecase),
		Seat:           handlers.NewSeatHandler(seatUsecase),
		Booking:        handlers.NewBookingHandler(bookingUsecase),
		Admin:          handlers.NewAdminHandler(adminUsecase),
		Verifier:       jwtIssuer,
		Users:          userRepo,
		Hub:            hub,
		WSVerifier:     jwtIssuer,
		FrontendOrigin: cfg.FrontendOrigin,
		Mongo:          mongorepo.NewPinger(db),
		Redis:          infraredis.NewPinger(redisClient),
		Kafka:          kafka.NewPinger(cfg.KafkaBrokers),
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
