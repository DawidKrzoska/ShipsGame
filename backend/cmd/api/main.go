package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"shipsgame/internal/config"
	httpapi "shipsgame/internal/http"
	redisstore "shipsgame/internal/store/redis"
	"shipsgame/internal/ws"
)

func main() {
	cfg := config.Load()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	redisClient := redisstore.NewClient(redisstore.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer func() {
		_ = redisClient.Close()
	}()
	if err := redisClient.Ping(context.Background()); err != nil {
		logger.Printf("redis ping failed: %v", err)
	}

	hub := ws.NewHub()
	go hub.Run()
	wsServer := &ws.Server{
		Hub:       hub,
		Store:     redisClient,
		JWTSecret: cfg.JWTSecret,
		Logger:    logger,
	}

	mux := httpapi.NewRouter(wsServer.Handler())

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Printf("listening on %s", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
