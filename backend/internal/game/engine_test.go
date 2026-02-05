package game

import "testing"

func TestPlaceShipOutOfBounds(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Carrier, Coord{Row: 9, Col: 6}, Horizontal); err != ErrOutOfBounds {
		t.Fatalf("expected out of bounds error, got %v", err)
	}
}

func TestPlaceShipOutOfBoundsVertical(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Battleship, Coord{Row: 8, Col: 0}, Vertical); err != ErrOutOfBounds {
		t.Fatalf("expected out of bounds error, got %v", err)
	}
}

func TestPlaceShipOverlap(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Destroyer, Coord{Row: 0, Col: 0}, Horizontal); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := board.PlaceShip(Submarine, Coord{Row: 0, Col: 0}, Vertical); err != ErrOverlap {
		t.Fatalf("expected overlap error, got %v", err)
	}
}

func TestPlaceShipDuplicateType(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Cruiser, Coord{Row: 2, Col: 2}, Horizontal); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := board.PlaceShip(Cruiser, Coord{Row: 4, Col: 4}, Horizontal); err != ErrShipAlreadyPlaced {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestPlaceShipUnknownType(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(ShipType("frigate"), Coord{Row: 1, Col: 1}, Horizontal); err != ErrUnknownShipType {
		t.Fatalf("expected unknown ship error, got %v", err)
	}
}

func TestPlaceShipInvalidOrientation(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Destroyer, Coord{Row: 1, Col: 1}, Orientation(99)); err != ErrOutOfBounds {
		t.Fatalf("expected out of bounds error for invalid orientation, got %v", err)
	}
}

func TestFireHitMissSunk(t *testing.T) {
	board := NewBoard()

	if err := board.PlaceShip(Destroyer, Coord{Row: 0, Col: 0}, Horizontal); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := board.FireAt(Coord{Row: 5, Col: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Outcome != ShotMiss {
		t.Fatalf("expected miss, got %v", result.Outcome)
	}

	result, err = board.FireAt(Coord{Row: 0, Col: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Outcome != ShotHit {
		t.Fatalf("expected hit, got %v", result.Outcome)
	}
	if result.ShipType != Destroyer {
		t.Fatalf("expected destroyer, got %v", result.ShipType)
	}

	result, err = board.FireAt(Coord{Row: 0, Col: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Outcome != ShotSunk {
		t.Fatalf("expected sunk, got %v", result.Outcome)
	}
	if result.ShipType != Destroyer {
		t.Fatalf("expected destroyer, got %v", result.ShipType)
	}
}

func TestFireOutOfBounds(t *testing.T) {
	board := NewBoard()

	if _, err := board.FireAt(Coord{Row: -1, Col: 0}); err != ErrOutOfBounds {
		t.Fatalf("expected out of bounds error, got %v", err)
	}
}

func TestFireAlreadyShot(t *testing.T) {
	board := NewBoard()

	if _, err := board.FireAt(Coord{Row: 1, Col: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := board.FireAt(Coord{Row: 1, Col: 1}); err != ErrAlreadyShot {
		t.Fatalf("expected already shot error, got %v", err)
	}
}

func TestAllShipsSunk(t *testing.T) {
	board := NewBoard()

	if board.AllShipsSunk() {
		t.Fatalf("expected false when no ships placed")
	}

	if err := board.PlaceShip(Destroyer, Coord{Row: 0, Col: 0}, Horizontal); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := board.PlaceShip(Submarine, Coord{Row: 2, Col: 2}, Vertical); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _ = board.FireAt(Coord{Row: 0, Col: 0})
	_, _ = board.FireAt(Coord{Row: 0, Col: 1})
	if board.AllShipsSunk() {
		t.Fatalf("expected false with remaining ships")
	}

	_, _ = board.FireAt(Coord{Row: 2, Col: 2})
	_, _ = board.FireAt(Coord{Row: 3, Col: 2})
	_, _ = board.FireAt(Coord{Row: 4, Col: 2})

	if !board.AllShipsSunk() {
		t.Fatalf("expected true when all ships sunk")
	}
}
