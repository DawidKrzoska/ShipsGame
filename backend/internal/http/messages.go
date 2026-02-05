package httpapi

type OpponentJoinedPayload struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
}
