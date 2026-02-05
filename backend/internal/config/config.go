package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddr    string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	PostgresDSN   string
	CORSOrigins   []string
}

func Load() Config {
	port := getenv("PORT", "8080")
	redisDB, _ := strconv.Atoi(getenv("REDIS_DB", "0"))

	return Config{
		ServerAddr:    ":" + port,
		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		JWTSecret:     getenv("JWT_SECRET", ""),
		PostgresDSN:   getenv("POSTGRES_DSN", "postgres://ships:ships@localhost:5432/ships?sslmode=disable"),
		CORSOrigins:   splitCSV(getenv("CORS_ORIGINS", "*")),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	raw := strings.Split(value, ",")
	out := make([]string, 0, len(raw))
	for _, entry := range raw {
		trimmed := strings.TrimSpace(entry)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
