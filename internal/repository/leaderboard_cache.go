package repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/redis/go-redis/v9"
)

const leaderboardKeyPattern = "leaderboard:game:%s"
const leaderboardUpdatesStreamPattern = "leaderboard:updates:game:%s"

func LeaderboardKey(gameID string) string {
	return fmt.Sprintf(leaderboardKeyPattern, gameID)
}

func LeaderboardUpdatesStream(gameID string) string {
	return fmt.Sprintf(leaderboardUpdatesStreamPattern, gameID)
}

type LeaderboardCache interface {
	IncrementScore(context.Context, string, string, int) error
	RebuildFromDb(ctx context.Context, repo ScoreRepository) error
}

type redisLeaderboardCache struct {
	rdb *redis.Client
}

func NewRedisLeaderboardCache(rdb *redis.Client) LeaderboardCache {
	return &redisLeaderboardCache{rdb: rdb}
}

func (c *redisLeaderboardCache) IncrementScore(ctx context.Context, gameID string, playerID string, delta int) error {
	if delta == 0 {
		return nil
	}

	key := LeaderboardKey(gameID)
	return c.rdb.ZIncrBy(ctx, key, float64(delta), playerID).Err()
}

func (c *redisLeaderboardCache) RebuildFromDb(ctx context.Context, repo ScoreRepository) error {
	if err := c.clearAllLeaderboards(ctx); err != nil {
		return err
	}

	scores, err := repo.GetAllLeaderboards(ctx)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	queued := 0

	flush := func() error {
		if queued == 0 {
			return nil
		}
		_, execErr := pipe.Exec(ctx)
		queued = 0
		return execErr
	}

	for _, score := range slices.Collect(scores) {
		key := LeaderboardKey(score.GameID)
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(score.Score), Member: score.PlayerID})
		queued++

		if queued >= 500 {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

func (c *redisLeaderboardCache) clearAllLeaderboards(ctx context.Context) error {
	var cursor uint64
	for {
		keys, next, err := c.rdb.Scan(ctx, cursor, fmt.Sprintf(leaderboardKeyPattern, "*"), 1000).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.rdb.Unlink(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
