package ws

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"shipsgame/internal/game"
	redisstore "shipsgame/internal/store/redis"
)

type messageEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func newTestServer(t *testing.T) (*Server, *redisstore.Client, func()) {
	server := miniredis.RunT(t)
	store := redisstore.NewClient(redisstore.Config{Addr: server.Addr()})
	hub := NewHub()
	go hub.Run()

	wsServer := &Server{
		Hub:       hub,
		Store:     store,
		JWTSecret: "secret",
		Logger:    log.New(io.Discard, "", 0),
	}

	cleanup := func() {
		_ = store.Close()
		server.Close()
	}

	return wsServer, store, cleanup
}

func registerClient(hub *Hub, client *Client) {
	hub.register <- client
	time.Sleep(10 * time.Millisecond)
}

func readMessage(t *testing.T, ch <-chan []byte) messageEnvelope {
	select {
	case data := <-ch:
		var msg messageEnvelope
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("unmarshal message: %v", err)
		}
		return msg
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for message")
	}
	return messageEnvelope{}
}

func TestHandleMessageInvalidJSON(t *testing.T) {
	wsServer, _, cleanup := newTestServer(t)
	defer cleanup()

	client := &Client{Hub: wsServer.Hub, GameID: "game", Player: "p1", Send: make(chan []byte, 1)}
	registerClient(wsServer.Hub, client)

	wsServer.handleMessage(client, []byte("not-json"))

	msg := readMessage(t, client.Send)
	if msg.Type != "error" {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHandleMessageUnknownType(t *testing.T) {
	wsServer, _, cleanup := newTestServer(t)
	defer cleanup()

	client := &Client{Hub: wsServer.Hub, GameID: "game", Player: "p1", Send: make(chan []byte, 1)}
	registerClient(wsServer.Hub, client)

	env := ClientMessage{Type: "unknown", Payload: json.RawMessage(`{}`)}
	data, _ := json.Marshal(env)
	wsServer.handleMessage(client, data)

	msg := readMessage(t, client.Send)
	if msg.Type != "error" {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHandlePlaceShipsMismatch(t *testing.T) {
	wsServer, _, cleanup := newTestServer(t)
	defer cleanup()

	client := &Client{Hub: wsServer.Hub, GameID: "game", Player: "p1", Send: make(chan []byte, 1)}
	registerClient(wsServer.Hub, client)

	payload := PlaceShipsPayload{GameID: "other", Ships: []ShipPayload{}}
	body, _ := json.Marshal(payload)
	env := ClientMessage{Type: "place_ships", Payload: body}
	data, _ := json.Marshal(env)
	wsServer.handleMessage(client, data)

	msg := readMessage(t, client.Send)
	if msg.Type != "error" {
		t.Fatalf("expected error, got %s", msg.Type)
	}
}

func TestHandlePlaceShipsSuccess(t *testing.T) {
	wsServer, store, cleanup := newTestServer(t)
	defer cleanup()

	meta, err := store.CreateGame(context.Background())
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	client := &Client{Hub: wsServer.Hub, GameID: meta.ID, Player: "p1", Send: make(chan []byte, 2)}
	registerClient(wsServer.Hub, client)

	payload := PlaceShipsPayload{
		GameID: meta.ID,
		Ships: []ShipPayload{
			{Type: "destroyer", Cells: []CoordPayload{{Row: 0, Col: 0}, {Row: 0, Col: 1}}},
		},
	}
	body, _ := json.Marshal(payload)
	env := ClientMessage{Type: "place_ships", Payload: body}
	data, _ := json.Marshal(env)
	wsServer.handleMessage(client, data)

	msg := readMessage(t, client.Send)
	if msg.Type != "game_state" {
		t.Fatalf("expected game_state, got %s", msg.Type)
	}
}

func TestHandleFireBroadcasts(t *testing.T) {
	wsServer, store, cleanup := newTestServer(t)
	defer cleanup()

	meta, err := store.CreateGame(context.Background())
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	p1 := redisstore.ShipsPlacement{
		game.Destroyer: {{Row: 0, Col: 0}, {Row: 0, Col: 1}},
	}
	p2 := redisstore.ShipsPlacement{
		game.Destroyer: {{Row: 2, Col: 0}, {Row: 2, Col: 1}},
	}
	if err := store.PlaceShips(context.Background(), meta.ID, "p1", p1); err != nil {
		t.Fatalf("place p1: %v", err)
	}
	if err := store.PlaceShips(context.Background(), meta.ID, "p2", p2); err != nil {
		t.Fatalf("place p2: %v", err)
	}

	client := &Client{Hub: wsServer.Hub, GameID: meta.ID, Player: "p1", Send: make(chan []byte, 4)}
	registerClient(wsServer.Hub, client)

	payload := FirePayload{GameID: meta.ID, Coord: CoordPayload{Row: 2, Col: 0}}
	body, _ := json.Marshal(payload)
	env := ClientMessage{Type: "fire", Payload: body}
	data, _ := json.Marshal(env)
	wsServer.handleMessage(client, data)

	first := readMessage(t, client.Send)
	second := readMessage(t, client.Send)
	if first.Type == second.Type {
		t.Fatalf("expected distinct messages, got %s", first.Type)
	}
	if (first.Type != "shot_result" && second.Type != "shot_result") ||
		(first.Type != "turn_changed" && second.Type != "turn_changed") {
		t.Fatalf("expected shot_result and turn_changed, got %s and %s", first.Type, second.Type)
	}
}
