package redisstore

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

type GameState struct {
	GameID string
	Turn   string
	Status string
	Winner string
	Shots  map[string]string
	Ships  map[string][][]int
}

func (c *Client) GetMeta(ctx context.Context, gameID string) (GameMeta, error) {
	fields, err := c.client.HGetAll(ctx, gameMetaKey(gameID)).Result()
	if err != nil {
		return GameMeta{}, err
	}
	if len(fields) == 0 {
		return GameMeta{}, ErrGameNotFound
	}
	return parseMeta(fields), nil
}

func (c *Client) GetState(ctx context.Context, gameID string, player string) (GameState, error) {
	meta, err := c.GetMeta(ctx, gameID)
	if err != nil {
		return GameState{}, err
	}

	shots, err := c.client.HGetAll(ctx, shotsKey(gameID, player)).Result()
	if err != nil {
		return GameState{}, err
	}

	shipsJSON, err := c.client.HGet(ctx, boardKey(gameID, player), "ships").Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return GameState{}, err
	}

	ships := map[string][][]int{}
	if shipsJSON != "" {
		_ = json.Unmarshal([]byte(shipsJSON), &ships)
	}

	return GameState{
		GameID: meta.ID,
		Turn:   meta.Turn,
		Status: meta.Status,
		Winner: meta.Winner,
		Shots:  shots,
		Ships:  ships,
	}, nil
}
