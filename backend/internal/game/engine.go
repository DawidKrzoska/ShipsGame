package game

import "errors"

const BoardSize = 10

type ShipType string

const (
	Carrier    ShipType = "carrier"
	Battleship ShipType = "battleship"
	Cruiser    ShipType = "cruiser"
	Submarine  ShipType = "submarine"
	Destroyer  ShipType = "destroyer"
)

var StandardShipSet = map[ShipType]int{
	Carrier:    5,
	Battleship: 4,
	Cruiser:    3,
	Submarine:  3,
	Destroyer:  2,
}

var (
	ErrOutOfBounds      = errors.New("coordinate out of bounds")
	ErrOverlap          = errors.New("ship placement overlaps existing ship")
	ErrShipAlreadyPlaced = errors.New("ship type already placed")
	ErrUnknownShipType  = errors.New("unknown ship type")
	ErrAlreadyShot      = errors.New("coordinate already shot")
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type Coord struct {
	Row int
	Col int
}

func (c Coord) InBounds() bool {
	return c.Row >= 0 && c.Row < BoardSize && c.Col >= 0 && c.Col < BoardSize
}

type ShotOutcome int

const (
	ShotMiss ShotOutcome = iota
	ShotHit
	ShotSunk
)

type ShotResult struct {
	Outcome  ShotOutcome
	ShipType ShipType
}

type Ship struct {
	Type  ShipType
	Size  int
	Cells []Coord
	Hits  map[Coord]bool
}

func (s *Ship) isSunk() bool {
	return len(s.Hits) == len(s.Cells)
}

type Board struct {
	ships    map[ShipType]*Ship
	occupied map[Coord]ShipType
	shots    map[Coord]ShotOutcome
}

func NewBoard() *Board {
	return &Board{
		ships:    make(map[ShipType]*Ship),
		occupied: make(map[Coord]ShipType),
		shots:    make(map[Coord]ShotOutcome),
	}
}

func (b *Board) PlaceShip(shipType ShipType, start Coord, orientation Orientation) error {
	if _, exists := b.ships[shipType]; exists {
		return ErrShipAlreadyPlaced
	}

	size, ok := StandardShipSet[shipType]
	if !ok {
		return ErrUnknownShipType
	}

	cells := make([]Coord, 0, size)
	for i := 0; i < size; i++ {
		coord := start
		switch orientation {
		case Horizontal:
			coord.Col += i
		case Vertical:
			coord.Row += i
		default:
			return ErrOutOfBounds
		}

		if !coord.InBounds() {
			return ErrOutOfBounds
		}
		if _, occupied := b.occupied[coord]; occupied {
			return ErrOverlap
		}
		cells = append(cells, coord)
	}

	ship := &Ship{
		Type:  shipType,
		Size:  size,
		Cells: cells,
		Hits:  make(map[Coord]bool),
	}

	b.ships[shipType] = ship
	for _, coord := range cells {
		b.occupied[coord] = shipType
	}

	return nil
}

func (b *Board) FireAt(coord Coord) (ShotResult, error) {
	if !coord.InBounds() {
		return ShotResult{}, ErrOutOfBounds
	}
	if _, already := b.shots[coord]; already {
		return ShotResult{}, ErrAlreadyShot
	}

	if shipType, hit := b.occupied[coord]; hit {
		ship := b.ships[shipType]
		ship.Hits[coord] = true
		if ship.isSunk() {
			b.shots[coord] = ShotSunk
			return ShotResult{Outcome: ShotSunk, ShipType: shipType}, nil
		}
		b.shots[coord] = ShotHit
		return ShotResult{Outcome: ShotHit, ShipType: shipType}, nil
	}

	b.shots[coord] = ShotMiss
	return ShotResult{Outcome: ShotMiss}, nil
}

func (b *Board) AllShipsSunk() bool {
	if len(b.ships) == 0 {
		return false
	}
	for _, ship := range b.ships {
		if !ship.isSunk() {
			return false
		}
	}
	return true
}
