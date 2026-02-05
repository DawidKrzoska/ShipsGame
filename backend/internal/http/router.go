package httpapi

import "net/http"

func NewRouter(wsHandler http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterHealth(mux)
	if wsHandler != nil {
		mux.Handle("/ws", wsHandler)
	}
	return mux
}
