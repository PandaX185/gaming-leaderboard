package db

import (
	"context"
	"gaming-leaderboard/internal/log"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(uri string) *redis.Client {
	var client *redis.Client
	var err error
	for range 5 {
		client = redis.NewClient(&redis.Options{Addr: uri})
		err = client.Ping(context.Background()).Err()
		if err == nil {
			return client
		}
		time.Sleep(2 * time.Second)
	}
	log.Panicf("failed to connect to redis %v", err.Error())
	return nil
}
