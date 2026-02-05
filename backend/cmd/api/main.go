package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"shipsgame/internal/config"
	httpapi "shipsgame/internal/http"
)

func main() {
	cfg := config.Load()

	mux := httpapi.NewRouter()

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	logger.Printf("listening on %s", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}
