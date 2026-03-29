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

func NewQueue(queueType string, repo repository.PlayerRepository, rdb *redis.Client) IQueue {
	switch queueType {
	case "redis":
		if rdb != nil {
			return NewRedisPlayerQueue(rdb, repo)
		}
		log.Println("Redis client not available, falling back to in-memory queue")
		return NewInMemoryPlayerQueue(repo)
	case "in-memory":
		return NewInMemoryPlayerQueue(repo)
	default:
		log.Printf("Unknown queue type '%s', defaulting to in-memory queue", queueType)
		return NewInMemoryPlayerQueue(repo)
	}
}
