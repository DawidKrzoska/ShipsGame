package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN string
}

type Client struct {
	Pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, err
	}

	return &Client{Pool: pool}, nil
}

func (c *Client) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}
