package httpapi

import (
	"log"
	"net/http"
)

type RouterConfig struct {
	WsHandler    http.Handler
	GamesHandler *GamesHandler
	CORS         CORSConfig
	Logger       *log.Logger
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
	RegisterHealth(mux)
	if cfg.WsHandler != nil {
		mux.Handle("/ws", cfg.WsHandler)
	}
	if cfg.GamesHandler != nil {
		cfg.GamesHandler.Register(mux)
	}
	handler := CORS(cfg.CORS, mux)
	if cfg.Logger != nil {
		handler = RequestLogger(cfg.Logger, handler)
	}
	return handler
}
