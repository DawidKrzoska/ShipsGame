package ws

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"shipsgame/internal/auth"
	"shipsgame/internal/game"
	redisstore "shipsgame/internal/store/redis"
)

type Server struct {
	Hub       *Hub
	Store     *redisstore.Client
	JWTSecret string
	Logger    *log.Logger
}

func (s *Server) Handler() http.Handler {
	handler := NewHandler(s.Hub)
	handler.OnMessage = s.handleMessage
	handler.Upgrader.CheckOrigin = func(*http.Request) bool { return true }
	handler.Auth = s.authenticate
	handler.OnConnect = s.SendInitialState
	return handler
}

func (s *Server) authenticate(r *http.Request) (string, string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", "", errors.New("missing authorization")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", "", errors.New("invalid authorization header")
	}

	claims, err := auth.ParseToken(parts[1], s.JWTSecret)
	if err != nil {
		return "", "", err
	}
	return claims.GameID, claims.Player, nil
}

func (s *Server) handleMessage(client *Client, message []byte) {
	var envelope ClientMessage
	if err := json.Unmarshal(message, &envelope); err != nil {
		s.sendError(client, "invalid message")
		return
	}

	switch envelope.Type {
	case "place_ships":
		s.handlePlaceShips(client, envelope.Payload)
	case "fire":
		s.handleFire(client, envelope.Payload)
	default:
		s.sendError(client, "unknown message type")
	}
}

func (s *Server) handlePlaceShips(client *Client, payload json.RawMessage) {
	var place PlaceShipsPayload
	if err := json.Unmarshal(payload, &place); err != nil {
		s.sendError(client, "invalid place_ships payload")
		return
	}
	if place.GameID != client.GameID {
		s.sendError(client, "game mismatch")
		return
	}

	placement := make(redisstore.ShipsPlacement, len(place.Ships))
	for _, ship := range place.Ships {
		cells := make([]game.Coord, 0, len(ship.Cells))
		for _, cell := range ship.Cells {
			cells = append(cells, game.Coord{Row: cell.Row, Col: cell.Col})
		}
		placement[game.ShipType(strings.ToLower(ship.Type))] = cells
	}

	if err := s.Store.PlaceShips(context.Background(), place.GameID, client.Player, placement); err != nil {
		if s.Logger != nil {
			s.Logger.Printf("ships place failed game_id=%s player=%s err=%v", place.GameID, client.Player, err)
		}
		s.sendError(client, err.Error())
		return
	}

	if s.Logger != nil {
		s.Logger.Printf("ships placed game_id=%s player=%s", place.GameID, client.Player)
	}

	state, err := s.Store.GetState(context.Background(), place.GameID, client.Player)
	if err == nil {
		s.broadcastState(place.GameID, state)
	}
}

func (s *Server) handleFire(client *Client, payload json.RawMessage) {
	var fire FirePayload
	if err := json.Unmarshal(payload, &fire); err != nil {
		s.sendError(client, "invalid fire payload")
		return
	}
	if fire.GameID != client.GameID {
		s.sendError(client, "game mismatch")
		return
	}

	result, err := s.Store.Fire(context.Background(), fire.GameID, client.Player, game.Coord{Row: fire.Coord.Row, Col: fire.Coord.Col})
	if err != nil {
		if s.Logger != nil {
			s.Logger.Printf("shot failed game_id=%s player=%s coord=%d,%d err=%v", fire.GameID, client.Player, fire.Coord.Row, fire.Coord.Col, err)
		}
		s.sendError(client, err.Error())
		return
	}

	if s.Logger != nil {
		s.Logger.Printf("shot fired game_id=%s player=%s coord=%d,%d outcome=%s", fire.GameID, client.Player, fire.Coord.Row, fire.Coord.Col, outcomeLabel(result.Outcome))
	}

	shotMsg := ServerMessage{
		Type: "shot_result",
		Payload: ShotResultPayload{
			GameID:  fire.GameID,
			Coord:   fire.Coord,
			Outcome: outcomeLabel(result.Outcome),
			Ship:    string(result.ShipType),
		},
	}
	if data, err := json.Marshal(shotMsg); err == nil {
		s.Hub.Broadcast(fire.GameID, data)
	}

	meta, err := s.Store.GetMeta(context.Background(), fire.GameID)
	if err == nil {
		turnMsg := ServerMessage{
			Type: "turn_changed",
			Payload: TurnChangedPayload{
				GameID: fire.GameID,
				Turn:   meta.Turn,
			},
		}
		if data, err := json.Marshal(turnMsg); err == nil {
			s.Hub.Broadcast(fire.GameID, data)
		}

		if meta.Status == "finished" {
			finished := ServerMessage{
				Type: "game_finished",
				Payload: GameFinishedPayload{
					GameID: fire.GameID,
					Winner: meta.Winner,
				},
			}
			if data, err := json.Marshal(finished); err == nil {
				s.Hub.Broadcast(fire.GameID, data)
			}
		}
	}
}

func (s *Server) SendInitialState(client *Client) {
	state, err := s.Store.GetState(context.Background(), client.GameID, client.Player)
	if err != nil {
		s.sendError(client, "failed to load game state")
		return
	}
	s.sendState(client, state)
}

func (s *Server) sendState(client *Client, state redisstore.GameState) {
	payload := GameStatePayload{
		GameID: state.GameID,
		Turn:   state.Turn,
		Status: state.Status,
		Winner: state.Winner,
		Shots:  state.Shots,
		Ships:  state.Ships,
	}
	msg := ServerMessage{Type: "game_state", Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	client.Send <- data
}

func (s *Server) broadcastState(gameID string, state redisstore.GameState) {
	payload := GameStatePayload{
		GameID: state.GameID,
		Turn:   state.Turn,
		Status: state.Status,
		Winner: state.Winner,
		Shots:  state.Shots,
		Ships:  state.Ships,
	}
	msg := ServerMessage{Type: "game_state", Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.Hub.Broadcast(gameID, data)
}

func (s *Server) sendError(client *Client, message string) {
	msg := ServerMessage{Type: "error", Payload: ErrorPayload{Message: message}}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	client.Send <- data
}

func outcomeLabel(outcome game.ShotOutcome) string {
	switch outcome {
	case game.ShotHit:
		return "hit"
	case game.ShotSunk:
		return "sunk"
	default:
		return "miss"
	}
}
