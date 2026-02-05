package httpapi

import "net/http"

type RouterConfig struct {
	WsHandler    http.Handler
	GamesHandler *GamesHandler
	CORS         CORSConfig
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
	return CORS(cfg.CORS, mux)
}
