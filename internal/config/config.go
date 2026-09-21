package config

import "os"

// Config holds everything read from the environment. No external
// dependency is needed for this — a handful of os.Getenv calls with
// sane local-dev defaults is enough for a service this size.
type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/magtapp?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
