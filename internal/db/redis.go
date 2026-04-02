package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(uri string) (*redis.Client, error) {
	var client *redis.Client
	var err error
	for i := 0; i < 5; i++ {
		client = redis.NewClient(&redis.Options{Addr: uri})
		err = client.Ping(context.Background()).Err()
		if err == nil {
			return client, nil
		}
		time.Sleep(2 * time.Second)
	}
	return nil, err
}
