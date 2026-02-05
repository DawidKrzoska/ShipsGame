package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerAddr    string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
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
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
