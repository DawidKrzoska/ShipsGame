package httpapi

import "net/http"

func NewRouter(wsHandler http.Handler, corsConfig CORSConfig) http.Handler {
	mux := http.NewServeMux()
	RegisterHealth(mux)
	if wsHandler != nil {
		mux.Handle("/ws", wsHandler)
	}
	return CORS(corsConfig, mux)
}
