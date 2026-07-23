package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	MongoURI           string
	RedisAddr          string
	KafkaBrokers       []string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
	FrontendOrigin     string
	LockTTL            time.Duration
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://localhost:27017/cinema"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers:       []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		FrontendOrigin:     getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		LockTTL:            time.Duration(getEnvInt("LOCK_TTL_SECONDS", 300)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
