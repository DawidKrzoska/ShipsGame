package redisstore

import (
	"context"
	"testing"

	"shipsgame/internal/game"
)

func TestGetMetaNotFound(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	_, err := client.GetMeta(context.Background(), "missing")
	if err != ErrGameNotFound {
		t.Fatalf("expected ErrGameNotFound, got %v", err)
	}
}

func TestGetStateEmpty(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	meta, err := client.CreateGame(context.Background())
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	state, err := client.GetState(context.Background(), meta.ID, "p1")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.GameID != meta.ID {
		t.Fatalf("expected game id %s, got %s", meta.ID, state.GameID)
	}
	if len(state.Ships) != 0 {
		t.Fatalf("expected no ships, got %v", state.Ships)
	}
	if len(state.Shots) != 0 {
		t.Fatalf("expected no shots, got %v", state.Shots)
	}
}

func TestGetStateWithShips(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	meta, err := client.CreateGame(context.Background())
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	placement := ShipsPlacement{
		game.Destroyer: {{Row: 0, Col: 0}, {Row: 0, Col: 1}},
	}
	if err := client.PlaceShips(context.Background(), meta.ID, "p1", placement); err != nil {
		t.Fatalf("place ships: %v", err)
	}

	state, err := client.GetState(context.Background(), meta.ID, "p1")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}

	cells, ok := state.Ships["destroyer"]
	if !ok {
		t.Fatalf("expected destroyer in ships")
	}
	if len(cells) != 2 {
		t.Fatalf("expected 2 cells, got %d", len(cells))
	}
}
