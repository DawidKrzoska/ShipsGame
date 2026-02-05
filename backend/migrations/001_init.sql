CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  display_name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS games (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  player1_id uuid REFERENCES users(id),
  player2_id uuid REFERENCES users(id),
  winner_id uuid REFERENCES users(id),
  status text NOT NULL,
  started_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS games_status_idx ON games (status);
CREATE INDEX IF NOT EXISTS games_player1_idx ON games (player1_id);
CREATE INDEX IF NOT EXISTS games_player2_idx ON games (player2_id);

CREATE TABLE IF NOT EXISTS game_events (
  id bigserial PRIMARY KEY,
  game_id uuid NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  seq int NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (game_id, seq)
);

CREATE INDEX IF NOT EXISTS game_events_game_idx ON game_events (game_id);

CREATE TABLE IF NOT EXISTS leaderboard (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  wins int NOT NULL DEFAULT 0,
  losses int NOT NULL DEFAULT 0,
  total_games int NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS leaderboard_wins_idx ON leaderboard (wins DESC, total_games DESC);
