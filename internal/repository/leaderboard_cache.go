package repository

import (
	"context"
	"fmt"
	"gaming-leaderboard/internal/dto"
	"iter"
	"slices"

	"github.com/redis/go-redis/v9"
)

const leaderboardKeyPattern = "leaderboard:game:%s"
const leaderboardUpdatesStreamPattern = "leaderboard:updates:game:%s"
const playerCountKey = "leaderboard:total_players"
const gameCountKey = "leaderboard:total_games"

func LeaderboardKey(gameID string) string {
	return fmt.Sprintf(leaderboardKeyPattern, gameID)
}

func LeaderboardUpdatesStream(gameID string) string {
	return fmt.Sprintf(leaderboardUpdatesStreamPattern, gameID)
}

type LeaderboardCache interface {
	IncrementScore(context.Context, string, string, int) error
	RebuildFromDb(context.Context, ScoreRepository, GameRepository, PlayerRepository) error
	GetTotalPlayersCount(ctx context.Context) (int, error)
	GetTotalGamesCount(ctx context.Context) (int, error)
	IncrementPlayerCount(ctx context.Context) error
	IncrementGameCount(ctx context.Context) error
}

type redisLeaderboardCache struct {
	rdb *redis.Client
}

func NewRedisLeaderboardCache(rdb *redis.Client) LeaderboardCache {
	return &redisLeaderboardCache{rdb: rdb}
}

func (c *redisLeaderboardCache) GetTotalPlayersCount(ctx context.Context) (int, error) {
	count, err := c.rdb.Get(ctx, playerCountKey).Int64()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *redisLeaderboardCache) GetTotalGamesCount(ctx context.Context) (int, error) {
	count, err := c.rdb.Get(ctx, gameCountKey).Int64()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (c *redisLeaderboardCache) IncrementPlayerCount(ctx context.Context) error {
	return c.rdb.Incr(ctx, playerCountKey).Err()
}

func (c *redisLeaderboardCache) IncrementGameCount(ctx context.Context) error {
	return c.rdb.Incr(ctx, gameCountKey).Err()
}

func (c *redisLeaderboardCache) IncrementScore(ctx context.Context, gameID string, playerID string, delta int) error {
	if delta == 0 {
		return nil
	}

	key := LeaderboardKey(fmt.Sprintf("%s", gameID))
	return c.rdb.ZIncrBy(ctx, key, float64(delta), playerID).Err()
}

func (c *redisLeaderboardCache) RebuildFromDb(ctx context.Context, scoreRepo ScoreRepository, gameRepo GameRepository, playerRepo PlayerRepository) error {
	if err := c.setTotalCount(ctx, playerRepo, gameRepo); err != nil {
		return err
	}

	if err := c.clearAllLeaderboards(ctx); err != nil {
		return err
	}

	scores, err := scoreRepo.GetAllLeaderboards(ctx)
	if err != nil {
		return err
	}

	return c.buildLeaderboards(ctx, scores)
}

func (c *redisLeaderboardCache) buildLeaderboards(ctx context.Context, scores iter.Seq[dto.ScoreResponse]) error {
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
		key := LeaderboardKey(fmt.Sprintf("%s", score.GameID))
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(score.Score), Member: fmt.Sprintf("%s", score.PlayerID)})
		queued++

		if queued >= 500 {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

func (c *redisLeaderboardCache) setTotalCount(ctx context.Context, playerRepo PlayerRepository, gameRepo GameRepository) error {
	totalPlayers, err := playerRepo.Count(ctx)
	if err != nil {
		return err
	}
	if err := c.rdb.Set(ctx, playerCountKey, totalPlayers, 0).Err(); err != nil {
		return err
	}

	totalGames, err := gameRepo.Count(ctx)
	if err != nil {
		return err
	}
	if err := c.rdb.Set(ctx, gameCountKey, totalGames, 0).Err(); err != nil {
		return err
	}

	return nil
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
