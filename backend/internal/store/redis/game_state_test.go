package redisstore

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"shipsgame/internal/game"
)

func newTestClient(t *testing.T) (*Client, func()) {
	server := miniredis.RunT(t)
	client := NewClient(Config{Addr: server.Addr()})

	return client, func() {
		_ = client.Close()
		server.Close()
	}
}

func TestCreateAndJoinGame(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	ctx := context.Background()
	meta, err := client.CreateGame(ctx)
	if err != nil {
		t.Fatalf("create game error: %v", err)
	}
	if meta.ID == "" || meta.JoinCode == "" {
		t.Fatalf("expected id and join code")
	}

	joined, player, err := client.JoinGame(ctx, meta.JoinCode)
	if err != nil {
		t.Fatalf("join game error: %v", err)
	}
	if joined.ID != meta.ID {
		t.Fatalf("expected same game id")
	}
	if player != playerTwo {
		t.Fatalf("expected player two, got %s", player)
	}
}

func TestPlaceShipsAndActivate(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	ctx := context.Background()
	meta, err := client.CreateGame(ctx)
	if err != nil {
		t.Fatalf("create game error: %v", err)
	}

	p1Placement := ShipsPlacement{
		game.Destroyer: {{Row: 0, Col: 0}, {Row: 0, Col: 1}},
		game.Submarine: {{Row: 2, Col: 0}, {Row: 3, Col: 0}, {Row: 4, Col: 0}},
	}
	if err := client.PlaceShips(ctx, meta.ID, playerOne, p1Placement); err != nil {
		t.Fatalf("place ships p1 error: %v", err)
	}

	p2Placement := ShipsPlacement{
		game.Destroyer: {{Row: 5, Col: 5}, {Row: 5, Col: 6}},
		game.Submarine: {{Row: 7, Col: 0}, {Row: 7, Col: 1}, {Row: 7, Col: 2}},
	}
	if err := client.PlaceShips(ctx, meta.ID, playerTwo, p2Placement); err != nil {
		t.Fatalf("place ships p2 error: %v", err)
	}

	metaFields, err := client.client.HGetAll(ctx, gameMetaKey(meta.ID)).Result()
	if err != nil {
		t.Fatalf("meta read error: %v", err)
	}
	if metaFields["status"] != "active" {
		t.Fatalf("expected active status, got %s", metaFields["status"])
	}
}

func TestFireFlow(t *testing.T) {
	client, cleanup := newTestClient(t)
	defer cleanup()

	ctx := context.Background()
	meta, err := client.CreateGame(ctx)
	if err != nil {
		t.Fatalf("create game error: %v", err)
	}

	p1Placement := ShipsPlacement{
		game.Destroyer: {{Row: 0, Col: 0}, {Row: 0, Col: 1}},
	}
	p2Placement := ShipsPlacement{
		game.Destroyer: {{Row: 5, Col: 5}, {Row: 5, Col: 6}},
	}

	if err := client.PlaceShips(ctx, meta.ID, playerOne, p1Placement); err != nil {
		t.Fatalf("place ships p1 error: %v", err)
	}
	if err := client.PlaceShips(ctx, meta.ID, playerTwo, p2Placement); err != nil {
		t.Fatalf("place ships p2 error: %v", err)
	}

	result, err := client.Fire(ctx, meta.ID, playerOne, game.Coord{Row: 5, Col: 5})
	if err != nil {
		t.Fatalf("fire error: %v", err)
	}
	if result.Outcome != game.ShotHit {
		t.Fatalf("expected hit, got %v", result.Outcome)
	}

	_, err = client.Fire(ctx, meta.ID, playerOne, game.Coord{Row: 5, Col: 5})
	if err == nil {
		t.Fatalf("expected already shot error")
	}

	if _, err := client.Fire(ctx, meta.ID, playerTwo, game.Coord{Row: 0, Col: 0}); err != nil {
		t.Fatalf("fire p2 error: %v", err)
	}

	result, err = client.Fire(ctx, meta.ID, playerOne, game.Coord{Row: 5, Col: 6})
	if err != nil {
		t.Fatalf("fire error: %v", err)
	}
	if result.Outcome != game.ShotSunk {
		t.Fatalf("expected sunk, got %v", result.Outcome)
	}
}
