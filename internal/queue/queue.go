package queue

import (
	"context"
	"gaming-leaderboard/internal/repository"
	"log"

	"github.com/redis/go-redis/v9"
)

type IQueue interface {
	PublishEvent(ctx context.Context, data any) error
	GetEvents() chan Event
}

func NewQueue(queueType string, repo any, rdb *redis.Client, leaderboardCache repository.LeaderboardCache) IQueue {
	if queueType != "redis" {
		log.Panicf("Unsupported queue type '%s': only 'redis' is allowed", queueType)
	}
	if rdb == nil {
		log.Panic("Redis queue selected but Redis client is nil")
	}

	switch r := repo.(type) {
	case repository.PlayerRepository:
		return NewRedisPlayerQueue(rdb, r, leaderboardCache)
	case repository.ScoreRepository:
		return NewScoreQueue(rdb, r, leaderboardCache)
	default:
		log.Panic("Unknown repository type for queue")
		return nil
	}
}
