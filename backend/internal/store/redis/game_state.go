package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"shipsgame/internal/game"
)

var (
	ErrGameNotFound     = errors.New("game not found")
	ErrGameFull         = errors.New("game already has two players")
	ErrInvalidJoinCode  = errors.New("invalid join code")
	ErrNotPlayerTurn    = errors.New("not player's turn")
	ErrGameNotActive    = errors.New("game not active")
	ErrPlayerNotReady   = errors.New("player not ready")
	ErrOpponentNotReady = errors.New("opponent not ready")
	ErrInvalidPlayer    = errors.New("invalid player")
	ErrInvalidPlacement = errors.New("invalid ship placement")
)

const (
	playerOne = "p1"
	playerTwo = "p2"
)

type GameMeta struct {
	ID          string
	JoinCode    string
	Status      string
	Turn        string
	Winner      string
	P1Ready     bool
	P2Ready     bool
	P1Joined    bool
	P2Joined    bool
	P1Remaining int
	P2Remaining int
}

type ShotResult struct {
	Outcome  game.ShotOutcome
	ShipType game.ShipType
}

type ShipsPlacement map[game.ShipType][]game.Coord

func (c *Client) CreateGame(ctx context.Context) (GameMeta, error) {
	id, err := randomHex(12)
	if err != nil {
		return GameMeta{}, err
	}
	joinCode, err := randomHex(3)
	if err != nil {
		return GameMeta{}, err
	}

	metaKey := gameMetaKey(id)
	joinKey := joinCodeKey(joinCode)
	pipe := c.client.TxPipeline()

	pipe.HSet(ctx, metaKey, map[string]any{
		"id":           id,
		"join_code":    joinCode,
		"status":       "waiting",
		"turn":         playerOne,
		"winner":       "",
		"p1_ready":     0,
		"p2_ready":     0,
		"p1_joined":    1,
		"p2_joined":    0,
		"p1_remaining": 0,
		"p2_remaining": 0,
	})
	pipe.Set(ctx, joinKey, id, 0)

	if _, err := pipe.Exec(ctx); err != nil {
		return GameMeta{}, err
	}

	return GameMeta{
		ID:       id,
		JoinCode: joinCode,
		Status:   "waiting",
		Turn:     playerOne,
		Winner:   "",
	}, nil
}

func (c *Client) JoinGame(ctx context.Context, joinCode string) (GameMeta, string, error) {
	id, err := c.client.Get(ctx, joinCodeKey(joinCode)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return GameMeta{}, "", ErrInvalidJoinCode
		}
		return GameMeta{}, "", err
	}

	metaKey := gameMetaKey(id)
	res, err := joinGameScript.Run(ctx, c.client, []string{metaKey}).Result()
	if err != nil {
		return GameMeta{}, "", err
	}
	if resStr, ok := res.(string); ok && resStr != "OK" {
		if strings.HasPrefix(resStr, "ERR:") {
			return GameMeta{}, "", errors.New(strings.TrimPrefix(resStr, "ERR:"))
		}
		return GameMeta{}, "", errors.New(resStr)
	}

	fields, err := c.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return GameMeta{}, "", err
	}
	if len(fields) == 0 {
		return GameMeta{}, "", ErrGameNotFound
	}

	meta := parseMeta(fields)
	return meta, playerTwo, nil
}

func (c *Client) PlaceShips(ctx context.Context, gameID string, player string, placement ShipsPlacement) error {
	if player != playerOne && player != playerTwo {
		return ErrInvalidPlayer
	}

	board := game.NewBoard()
	for shipType, coords := range placement {
		orientation, start, err := validateCoords(shipType, coords)
		if err != nil {
			return err
		}

		if err := board.PlaceShip(shipType, start, orientation); err != nil {
			return err
		}
	}

	shipsJSON, err := json.Marshal(board.Ships())
	if err != nil {
		return err
	}

	occupancy := board.Occupied()
	occMap := make(map[string]string, len(occupancy))
	for coord, shipType := range occupancy {
		occMap[coordKey(coord)] = string(shipType)
	}

	remainingByShip := make(map[string]int, len(board.Ships()))
	remainingTotal := 0
	for shipType, coords := range board.Ships() {
		remainingByShip[string(shipType)] = len(coords)
		remainingTotal += len(coords)
	}

	keys := []string{
		gameMetaKey(gameID),
		boardKey(gameID, player),
		occupancyKey(gameID, player),
		shipsKey(gameID, player),
	}

	args := []any{
		player,
		string(shipsJSON),
		remainingTotal,
	}
	for field, value := range occMap {
		args = append(args, field, value)
	}
	args = append(args, "__ships__")
	for shipType, remaining := range remainingByShip {
		args = append(args, shipType, remaining)
	}

	res, err := placeShipsScript.Run(ctx, c.client, keys, args...).Result()
	if err != nil {
		return err
	}
	if resStr, ok := res.(string); ok && resStr != "OK" {
		return errors.New(resStr)
	}
	return nil
}

func (c *Client) Fire(ctx context.Context, gameID string, player string, coord game.Coord) (ShotResult, error) {
	if player != playerOne && player != playerTwo {
		return ShotResult{}, ErrInvalidPlayer
	}
	if !coord.InBounds() {
		return ShotResult{}, game.ErrOutOfBounds
	}

	args := []any{player, coordKey(coord)}
	res, err := fireScript.Run(ctx, c.client, []string{
		gameMetaKey(gameID),
		shotsKey(gameID, player),
		shotsKey(gameID, opponent(player)),
		occupancyKey(gameID, opponent(player)),
		shipsKey(gameID, opponent(player)),
	}, args...).Result()
	if err != nil {
		return ShotResult{}, err
	}

	resultStr, ok := res.(string)
	if !ok {
		return ShotResult{}, errors.New("unexpected redis response")
	}
	if strings.HasPrefix(resultStr, "ERR:") {
		return ShotResult{}, errors.New(strings.TrimPrefix(resultStr, "ERR:"))
	}

	parts := strings.Split(resultStr, ":")
	outcome := parts[0]
	shot := ShotResult{}
	switch outcome {
	case "miss":
		shot.Outcome = game.ShotMiss
	case "hit":
		shot.Outcome = game.ShotHit
	case "sunk":
		shot.Outcome = game.ShotSunk
		if len(parts) == 2 {
			shot.ShipType = game.ShipType(parts[1])
		}
	default:
		return ShotResult{}, errors.New("unknown shot outcome")
	}

	return shot, nil
}

func validateCoords(shipType game.ShipType, coords []game.Coord) (game.Orientation, game.Coord, error) {
	size, ok := game.StandardShipSet[shipType]
	if !ok {
		return 0, game.Coord{}, game.ErrUnknownShipType
	}
	if len(coords) != size {
		return 0, game.Coord{}, ErrInvalidPlacement
	}

	seen := make(map[game.Coord]bool, len(coords))
	minRow, maxRow := coords[0].Row, coords[0].Row
	minCol, maxCol := coords[0].Col, coords[0].Col
	row := coords[0].Row
	col := coords[0].Col
	sameRow := true
	sameCol := true

	for _, coord := range coords {
		if !coord.InBounds() {
			return 0, game.Coord{}, game.ErrOutOfBounds
		}
		if seen[coord] {
			return 0, game.Coord{}, ErrInvalidPlacement
		}
		seen[coord] = true
		if coord.Row != row {
			sameRow = false
		}
		if coord.Col != col {
			sameCol = false
		}
		if coord.Row < minRow {
			minRow = coord.Row
		}
		if coord.Row > maxRow {
			maxRow = coord.Row
		}
		if coord.Col < minCol {
			minCol = coord.Col
		}
		if coord.Col > maxCol {
			maxCol = coord.Col
		}
	}

	switch {
	case sameRow && !sameCol:
		if maxCol-minCol != size-1 {
			return 0, game.Coord{}, ErrInvalidPlacement
		}
		for c := minCol; c <= maxCol; c++ {
			if !seen[game.Coord{Row: row, Col: c}] {
				return 0, game.Coord{}, ErrInvalidPlacement
			}
		}
		return game.Horizontal, game.Coord{Row: row, Col: minCol}, nil
	case sameCol && !sameRow:
		if maxRow-minRow != size-1 {
			return 0, game.Coord{}, ErrInvalidPlacement
		}
		for r := minRow; r <= maxRow; r++ {
			if !seen[game.Coord{Row: r, Col: col}] {
				return 0, game.Coord{}, ErrInvalidPlacement
			}
		}
		return game.Vertical, game.Coord{Row: minRow, Col: col}, nil
	case sameRow && sameCol:
		return 0, game.Coord{}, ErrInvalidPlacement
	default:
		return 0, game.Coord{}, ErrInvalidPlacement
	}
}

func coordKey(coord game.Coord) string {
	return fmt.Sprintf("%d,%d", coord.Row, coord.Col)
}

func opponent(player string) string {
	if player == playerOne {
		return playerTwo
	}
	return playerOne
}

func gameMetaKey(id string) string {
	return fmt.Sprintf("game:%s:meta", id)
}

func boardKey(id, player string) string {
	return fmt.Sprintf("game:%s:board:%s", id, player)
}

func occupancyKey(id, player string) string {
	return fmt.Sprintf("game:%s:occupancy:%s", id, player)
}

func shipsKey(id, player string) string {
	return fmt.Sprintf("game:%s:ships:%s", id, player)
}

func shotsKey(id, player string) string {
	return fmt.Sprintf("game:%s:shots:%s", id, player)
}

func joinCodeKey(code string) string {
	return fmt.Sprintf("game:join:%s", code)
}

func parseMeta(fields map[string]string) GameMeta {
	return GameMeta{
		ID:          fields["id"],
		JoinCode:    fields["join_code"],
		Status:      fields["status"],
		Turn:        fields["turn"],
		Winner:      fields["winner"],
		P1Ready:     fields["p1_ready"] == "1",
		P2Ready:     fields["p2_ready"] == "1",
		P1Joined:    fields["p1_joined"] == "1",
		P2Joined:    fields["p2_joined"] == "1",
		P1Remaining: atoi(fields["p1_remaining"]),
		P2Remaining: atoi(fields["p2_remaining"]),
	}
}

func atoi(value string) int {
	if value == "" {
		return 0
	}
	var out int
	_, _ = fmt.Sscanf(value, "%d", &out)
	return out
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

var placeShipsScript = redis.NewScript(`
local meta = KEYS[1]
local board = KEYS[2]
local occupancy = KEYS[3]
local ships = KEYS[4]

local player = ARGV[1]
local ships_json = ARGV[2]
local remaining_total = ARGV[3]

if redis.call('EXISTS', meta) == 0 then
  return 'ERR:game_not_found'
end

local status = redis.call('HGET', meta, 'status')
if status ~= 'waiting' and status ~= 'placing' then
  return 'ERR:invalid_status'
end

local ready_field = player .. '_ready'
local already_ready = redis.call('HGET', meta, ready_field)
if already_ready == '1' then
  return 'ERR:already_ready'
end

redis.call('DEL', board)
redis.call('DEL', occupancy)
redis.call('DEL', ships)

redis.call('HSET', board, 'ships', ships_json)

local idx = 4
while idx <= #ARGV and ARGV[idx] ~= '__ships__' do
  redis.call('HSET', occupancy, ARGV[idx], ARGV[idx + 1])
  idx = idx + 2
end

idx = idx + 1
while idx <= #ARGV do
  redis.call('HSET', ships, ARGV[idx], ARGV[idx + 1])
  idx = idx + 2
end

redis.call('HSET', meta, ready_field, 1)
redis.call('HSET', meta, player .. '_remaining', remaining_total)

local p1_ready = redis.call('HGET', meta, 'p1_ready')
local p2_ready = redis.call('HGET', meta, 'p2_ready')
if p1_ready == '1' and p2_ready == '1' then
  redis.call('HSET', meta, 'status', 'active')
else
  redis.call('HSET', meta, 'status', 'placing')
end

return 'OK'
`)

var fireScript = redis.NewScript(`
local meta = KEYS[1]
local shooter_shots = KEYS[2]
local opponent_shots = KEYS[3]
local opponent_occupancy = KEYS[4]
local opponent_ships = KEYS[5]

local player = ARGV[1]
local coord = ARGV[2]

if redis.call('EXISTS', meta) == 0 then
  return 'ERR:game_not_found'
end

local status = redis.call('HGET', meta, 'status')
if status ~= 'active' then
  return 'ERR:game_not_active'
end

local turn = redis.call('HGET', meta, 'turn')
if turn ~= player then
  return 'ERR:not_player_turn'
end

local already = redis.call('HGET', shooter_shots, coord)
if already then
  return 'ERR:already_shot'
end

local ship_type = redis.call('HGET', opponent_occupancy, coord)
if not ship_type then
  redis.call('HSET', shooter_shots, coord, 'miss')
  local next = (player == 'p1') and 'p2' or 'p1'
  redis.call('HSET', meta, 'turn', next)
  return 'miss'
end

redis.call('HSET', shooter_shots, coord, 'hit')
local next = (player == 'p1') and 'p2' or 'p1'

local remaining = tonumber(redis.call('HINCRBY', opponent_ships, ship_type, -1))
local remaining_total_field = (player == 'p1') and 'p2_remaining' or 'p1_remaining'
local remaining_total = tonumber(redis.call('HINCRBY', meta, remaining_total_field, -1))

if remaining == 0 then
  redis.call('HSET', shooter_shots, coord, 'sunk:' .. ship_type)
  if remaining_total == 0 then
    redis.call('HSET', meta, 'status', 'finished')
    redis.call('HSET', meta, 'winner', player)
  else
    redis.call('HSET', meta, 'turn', next)
  end
  return 'sunk:' .. ship_type
end

redis.call('HSET', meta, 'turn', next)
return 'hit'
`)

var joinGameScript = redis.NewScript(`
local meta = KEYS[1]

if redis.call('EXISTS', meta) == 0 then
  return 'ERR:game_not_found'
end

local p2_joined = redis.call('HGET', meta, 'p2_joined')
if p2_joined == '1' then
  return 'ERR:game_full'
end

redis.call('HSET', meta, 'p2_joined', 1)
return 'OK'
`)
