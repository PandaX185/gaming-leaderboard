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

func NewQueue(queueType string, repo repository.PlayerRepository, rdb *redis.Client, leaderboardCache repository.LeaderboardCache) IQueue {
	if queueType != "redis" {
		log.Panicf("Unsupported queue type '%s': only 'redis' is allowed", queueType)
	}
	if rdb == nil {
		log.Panic("Redis queue selected but Redis client is nil")
	}

	return NewRedisPlayerQueue(rdb, repo, leaderboardCache)
}
