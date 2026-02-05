package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"shipsgame/internal/auth"
	redisstore "shipsgame/internal/store/redis"
)

type GamesHandler struct {
	Store     *redisstore.Client
	JWTSecret string
	Logger    *log.Logger
}

type CreateGameResponse struct {
	GameID   string `json:"game_id"`
	JoinCode string `json:"join_code"`
	Player   string `json:"player"`
	Token    string `json:"token"`
}

type JoinGameRequest struct {
	JoinCode string `json:"join_code"`
}

type JoinGameResponse struct {
	GameID string `json:"game_id"`
	Player string `json:"player"`
	Token  string `json:"token"`
}

func (h *GamesHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/games", h.handleCreate)
	mux.HandleFunc("/games/join", h.handleJoin)
}

func (h *GamesHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.JWTSecret == "" {
		writeError(w, http.StatusInternalServerError, "missing JWT secret")
		return
	}

	meta, err := h.Store.CreateGame(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create game")
		return
	}

	claims := auth.Claims{
		GameID: meta.ID,
		Player: "p1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token, err := auth.SignToken(h.JWTSecret, claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	writeJSON(w, http.StatusOK, CreateGameResponse{
		GameID:   meta.ID,
		JoinCode: meta.JoinCode,
		Player:   "p1",
		Token:    token,
	})

	if h.Logger != nil {
		h.Logger.Printf("game created game_id=%s join_code=%s player=p1", meta.ID, meta.JoinCode)
	}
}

func (h *GamesHandler) handleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.JWTSecret == "" {
		writeError(w, http.StatusInternalServerError, "missing JWT secret")
		return
	}

	var req JoinGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.JoinCode == "" {
		writeError(w, http.StatusBadRequest, "join_code required")
		return
	}

	meta, player, err := h.Store.JoinGame(r.Context(), req.JoinCode)
	if err != nil {
		switch {
		case errors.Is(err, redisstore.ErrInvalidJoinCode):
			writeError(w, http.StatusBadRequest, "invalid join code")
		case errors.Is(err, redisstore.ErrGameFull):
			writeError(w, http.StatusConflict, "game full")
		default:
			writeError(w, http.StatusInternalServerError, "failed to join game")
		}
		return
	}

	claims := auth.Claims{
		GameID: meta.ID,
		Player: player,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token, err := auth.SignToken(h.JWTSecret, claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	writeJSON(w, http.StatusOK, JoinGameResponse{
		GameID: meta.ID,
		Player: player,
		Token:  token,
	})

	if h.Logger != nil {
		h.Logger.Printf("game joined game_id=%s player=%s", meta.ID, player)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
