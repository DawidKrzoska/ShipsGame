package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v2"
)

func TestSaveGameWritesAll(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock pool: %v", err)
	}
	defer mock.Close()

	store := &Store{db: mock}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO games").WithArgs(
		"game-1", "p1", "p2", "p1", "finished", pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	eventPayload, _ := json.Marshal(map[string]string{"type": "fire"})
	mock.ExpectExec("INSERT INTO game_events").WithArgs(
		"game-1", 1, "shot", pgxmock.AnyArg(), pgxmock.AnyArg(),
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectExec("INSERT INTO leaderboard").WithArgs(
		"p1", 1, 0, 1,
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectExec("INSERT INTO leaderboard").WithArgs(
		"p2", 0, 1, 1,
	).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	mock.ExpectCommit()

	started := time.Now().Add(-time.Minute)
	finished := time.Now()

	err = store.SaveGame(context.Background(), GameSummary{
		GameID:     "game-1",
		Player1ID:  "p1",
		Player2ID:  "p2",
		WinnerID:   "p1",
		LoserID:    "p2",
		Status:     "finished",
		StartedAt:  &started,
		FinishedAt: &finished,
	}, []GameEvent{{
		Seq:       1,
		EventType: "shot",
		Payload:   eventPayload,
		CreatedAt: finished,
	}})
	if err != nil {
		t.Fatalf("save game: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestGetLeaderboard(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("mock pool: %v", err)
	}
	defer mock.Close()

	store := &Store{db: mock}

	rows := pgxmock.NewRows([]string{"user_id", "wins", "losses", "total_games"}).
		AddRow("u1", 2, 1, 3).
		AddRow("u2", 1, 0, 1)

	mock.ExpectQuery("SELECT user_id, wins, losses, total_games").WithArgs(10).WillReturnRows(rows)

	entries, err := store.GetLeaderboard(context.Background(), 10)
	if err != nil {
		t.Fatalf("get leaderboard: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
