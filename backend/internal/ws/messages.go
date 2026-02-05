package ws

import "encoding/json"

type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PlaceShipsPayload struct {
	GameID string        `json:"game_id"`
	Ships  []ShipPayload `json:"ships"`
}

type ShipPayload struct {
	Type  string         `json:"type"`
	Cells []CoordPayload `json:"cells"`
}

type FirePayload struct {
	GameID string       `json:"game_id"`
	Coord  CoordPayload `json:"coord"`
}

type CoordPayload struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type GameStatePayload struct {
	GameID string             `json:"game_id"`
	Turn   string             `json:"turn"`
	Status string             `json:"status"`
	Winner string             `json:"winner"`
	Shots  map[string]string  `json:"shots"`
	Ships  map[string][][]int `json:"ships"`
}

type ShotResultPayload struct {
	GameID  string       `json:"game_id"`
	Coord   CoordPayload `json:"coord"`
	Outcome string       `json:"outcome"`
	Ship    string       `json:"ship"`
}

type TurnChangedPayload struct {
	GameID string `json:"game_id"`
	Turn   string `json:"turn"`
}

type GameFinishedPayload struct {
	GameID string `json:"game_id"`
	Winner string `json:"winner"`
}

type OpponentJoinedPayload struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}
