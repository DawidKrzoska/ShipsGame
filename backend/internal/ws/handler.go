package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Hub       *Hub
	Upgrader  websocket.Upgrader
	OnMessage func(*Client, []byte)
	Auth      func(*http.Request) (string, string, error)
	OnConnect func(*Client)
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		Hub: hub,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gameID := ""
	player := ""
	if h.Auth != nil {
		var err error
		gameID, player, err = h.Auth(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	} else {
		gameID = r.URL.Query().Get("game_id")
		player = r.URL.Query().Get("player")
		if gameID == "" || player == "" {
			http.Error(w, "missing game_id or player", http.StatusBadRequest)
			return
		}
	}

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Hub:    h.Hub,
		Conn:   conn,
		GameID: gameID,
		Player: player,
		Send:   make(chan []byte, 256),
	}

	h.Hub.register <- client
	if h.OnConnect != nil {
		h.OnConnect(client)
	}

	go client.writePump()
	go client.readPump(h.OnMessage)
}
