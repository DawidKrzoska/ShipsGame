package ws

import (
	"encoding/json"
	"sync"
)

type HubMessage struct {
	GameID string
	Data   []byte
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan HubMessage

	mu    sync.RWMutex
	rooms map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan HubMessage),
		rooms:      make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			room := h.rooms[client.GameID]
			if room == nil {
				room = make(map[*Client]bool)
				h.rooms[client.GameID] = room
			}
			room[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			room := h.rooms[client.GameID]
			if room != nil {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
				}
				if len(room) == 0 {
					delete(h.rooms, client.GameID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			room := h.rooms[msg.GameID]
			for client := range room {
				select {
				case client.Send <- msg.Data:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastRaw(gameID string, data []byte) {
	h.broadcast <- HubMessage{GameID: gameID, Data: data}
}

func (h *Hub) Broadcast(gameID string, messageType string, payload any) {
	msg := ServerMessage{
		Type:    messageType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.broadcast <- HubMessage{GameID: gameID, Data: data}
}
