package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type Store struct {
	db DB
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{db: pool}
}

var _ DB = (*pgxpool.Pool)(nil)

const (
	statusFinished = "finished"
)

type GameSummary struct {
	GameID     string
	Player1ID  string
	Player2ID  string
	WinnerID   string
	LoserID    string
	Status     string
	StartedAt  *time.Time
	FinishedAt *time.Time
}

type GameEvent struct {
	Seq       int
	EventType string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type LeaderboardEntry struct {
	UserID     string
	Wins       int
	Losses     int
	TotalGames int
}

func (s *Store) SaveGame(ctx context.Context, summary GameSummary, events []GameEvent) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO games (id, player1_id, player2_id, winner_id, status, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, summary.GameID, summary.Player1ID, summary.Player2ID, summary.WinnerID, summary.Status, summary.StartedAt, summary.FinishedAt)
	if err != nil {
		return err
	}

	for _, event := range events {
		_, err = tx.Exec(ctx, `
			INSERT INTO game_events (game_id, seq, event_type, payload, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, summary.GameID, event.Seq, event.EventType, event.Payload, event.CreatedAt)
		if err != nil {
			return err
		}
	}

	if summary.Status == statusFinished && summary.WinnerID != "" {
		if err := upsertLeaderboard(tx, ctx, summary.WinnerID, 1, 0); err != nil {
			return err
		}
		if summary.LoserID != "" {
			if err := upsertLeaderboard(tx, ctx, summary.LoserID, 0, 1); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func upsertLeaderboard(tx pgx.Tx, ctx context.Context, userID string, winsDelta, lossesDelta int) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO leaderboard (user_id, wins, losses, total_games)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET wins = leaderboard.wins + EXCLUDED.wins,
			losses = leaderboard.losses + EXCLUDED.losses,
			total_games = leaderboard.total_games + EXCLUDED.total_games,
			updated_at = now()
	`, userID, winsDelta, lossesDelta, winsDelta+lossesDelta)
	return err
}

func (s *Store) GetLeaderboard(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	rows, err := s.db.Query(ctx, `
		SELECT user_id, wins, losses, total_games
		FROM leaderboard
		ORDER BY wins DESC, total_games DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{}
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.Wins, &entry.Losses, &entry.TotalGames); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (s *Store) GetGameEvents(ctx context.Context, gameID string) ([]GameEvent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT seq, event_type, payload, created_at
		FROM game_events
		WHERE game_id = $1
		ORDER BY seq ASC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []GameEvent{}
	for rows.Next() {
		var entry GameEvent
		if err := rows.Scan(&entry.Seq, &entry.EventType, &entry.Payload, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
