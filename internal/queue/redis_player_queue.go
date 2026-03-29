package queue

import (
	"context"
	"encoding/json"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gaming-leaderboard/metrics"
)

const (
	ConsumerGroup = "worker-group"
	ReclaimTime   = time.Minute * 1
)

type RedisPlayerQueue struct {
	rdb        *redis.Client
	repo       repository.PlayerRepository
	consumerID string
}

func NewRedisPlayerQueue(rdb *redis.Client, repo repository.PlayerRepository) IQueue {
	err := rdb.XGroupCreateMkStream(context.Background(), consts.PlayerGameCollection, ConsumerGroup, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "exists") {
		log.Printf("Error creating consumer group: %v", err)
	}

	hostname, _ := os.Hostname()
	consumerID := hostname + "-" + uuid.NewString()

	go func() {
		for {
			length, err := rdb.XLen(context.Background(), consts.PlayerGameCollection).Result()
			if err == nil {
				metrics.QueueSize.Set(float64(length))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return &RedisPlayerQueue{
		rdb:        rdb,
		repo:       repo,
		consumerID: consumerID,
	}
}

func (q *RedisPlayerQueue) PublishEvent(ctx context.Context, data any) error {
	var eventType string
	switch data.(type) {
	case *dto.CreatePlayerRequest:
		eventType = "PlayerCreated"
	case *dto.UpdateScoreRequest:
		eventType = "ScoreUpdated"
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: consts.PlayerGameCollection,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{
			"type":    eventType,
			"payload": string(payloadBytes),
		},
	}).Err()
}

func (q *RedisPlayerQueue) GetEvents() chan Event {
	events := make(chan Event)

	go func() {
		for {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
				defer cancel()

				streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    ConsumerGroup,
					Consumer: q.consumerID,
					Streams:  []string{consts.PlayerGameCollection, ">"},
					Count:    50,
					Block:    time.Second * 5,
				}).Result()

				if err == redis.Nil {
					return
				} else if err != nil {
					log.Printf("Error reading from Redis Group: %v", err)
					time.Sleep(time.Second)
					return
				}

				for _, stream := range streams {
					q.processMessages(stream.Messages, events)
				}
			}()
		}
	}()

	go func() {
		for {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
					Stream:   consts.PlayerGameCollection,
					Group:    ConsumerGroup,
					Consumer: q.consumerID,
					MinIdle:  ReclaimTime,
					Start:    "0-0",
					Count:    50,
				}).Result()

				if err != nil && err != redis.Nil {
					log.Printf("Error during XAutoClaim: %v", err)
					time.Sleep(time.Second)
				} else if len(messages) > 0 {
					log.Printf("Reclaimed %d pending messages", len(messages))
					q.processMessages(messages, events)
				}

				time.Sleep(time.Second * 10)
			}()
		}
	}()

	return events
}

func (q *RedisPlayerQueue) processMessages(messages []redis.XMessage, events chan Event) {
	for _, message := range messages {
		eventType, _ := message.Values["type"].(string)
		payloadStr, _ := message.Values["payload"].(string)

		var event Event
		event.Type = eventType
		event.Attempt = 0

		var dbHandler func(workerCtx context.Context, p any) error

		switch eventType {
		case "PlayerCreated":
			var req dto.CreatePlayerRequest
			if err := json.Unmarshal([]byte(payloadStr), &req); err == nil {
				event.Payload = &req
				dbHandler = func(workerCtx context.Context, p any) error {
					return q.repo.Insert(workerCtx, p.(*dto.CreatePlayerRequest))
				}
			}
		case "ScoreUpdated":
			var req dto.UpdateScoreRequest
			if err := json.Unmarshal([]byte(payloadStr), &req); err == nil {
				event.Payload = &req
				dbHandler = func(workerCtx context.Context, p any) error {
					return q.repo.UpdateScore(workerCtx, p.(*dto.UpdateScoreRequest))
				}
			}
		}

		if dbHandler != nil {
			msgID := message.ID
			event.Handler = func(workerCtx context.Context, p any) error {
				err := dbHandler(workerCtx, p)
				if err == nil {
					ackErr := q.rdb.XAck(workerCtx, consts.PlayerGameCollection, ConsumerGroup, msgID).Err()
					if ackErr != nil {
						log.Printf("Failed to XACK message %s: %v", msgID, ackErr)
					}
				}
				return err
			}
			event.Ack = func(workerCtx context.Context) error {
				return q.rdb.XAck(workerCtx, consts.PlayerGameCollection, ConsumerGroup, msgID).Err()
			}
			events <- event
		}
	}
}
