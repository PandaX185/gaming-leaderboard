package repository

import (
	"context"
	"fmt"
	"gaming-leaderboard/internal/consts"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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
	RebuildFromMongo(context.Context, *mongo.Database) error
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

func (c *redisLeaderboardCache) RebuildFromMongo(ctx context.Context, db *mongo.Database) error {
	if err := c.clearAllLeaderboards(ctx); err != nil {
		return err
	}

	type scoreDoc struct {
		PlayerID primitive.ObjectID `bson:"player_id"`
		GameID   primitive.ObjectID `bson:"game_id"`
		Score    int                `bson:"score"`
	}

	cursor, err := db.Collection(consts.PlayerGameCollection).Find(
		ctx,
		bson.M{},
		options.Find().SetProjection(bson.M{"player_id": 1, "game_id": 1, "score": 1}),
	)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

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

	for cursor.Next(ctx) {
		var doc scoreDoc
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		key := LeaderboardKey(doc.GameID.Hex())
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(doc.Score), Member: doc.PlayerID.Hex()})
		queued++

		if queued >= 500 {
			if err := flush(); err != nil {
				return err
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return err
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
