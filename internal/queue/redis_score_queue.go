package queue

import (
	"context"
	"encoding/json"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
	"gaming-leaderboard/metrics"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ScoreQueue struct {
	rdb              *redis.Client
	scoreRepo        repository.ScoreRepository
	leaderboardCache repository.LeaderboardCache
}

func NewScoreQueue(rdb *redis.Client, scoreRepo repository.ScoreRepository, leaderboardCache repository.LeaderboardCache) *ScoreQueue {
	if rdb == nil {
		log.Panic("Redis queue selected but Redis client is nil")
	}
	return &ScoreQueue{
		rdb:              rdb,
		scoreRepo:        scoreRepo,
		leaderboardCache: leaderboardCache,
	}
}

func (q *ScoreQueue) PublishEvent(ctx context.Context, data any) error {
	var eventType string
	switch data.(type) {
	case *dto.UpdateScoreEvent:
		eventType = "ScoreUpdated"
	default:
		return nil
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: consts.ScoreEvents,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{
			"type":    eventType,
			"payload": string(payloadBytes),
		},
	}).Err()
	if err != nil {
		metrics.QueuePublishedTotal.WithLabelValues(eventType+"_score", "error").Inc()
		return err
	}
	metrics.QueuePublishedTotal.WithLabelValues(eventType+"_score", "success").Inc()
	return nil
}

func (q *ScoreQueue) GetEvents() chan Event {
	events := make(chan Event)

	err := q.rdb.XGroupCreateMkStream(context.Background(), consts.ScoreEvents, consts.ScoreConsumerGroup, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "exists") {
		log.Printf("Error creating score consumer group: %v", err)
	}

	go func() {
		for {
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
				defer cancel()

				streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    consts.ScoreConsumerGroup,
					Consumer: "score-consumer",
					Streams:  []string{consts.ScoreEvents, ">"},
					Count:    50,
					Block:    time.Second * 5,
				}).Result()

				if err == redis.Nil {
					return
				} else if err != nil {
					log.Printf("Error reading from Score Redis Group: %v", err)
					metrics.QueueReadErrorsTotal.WithLabelValues("score_xreadgroup").Inc()
					time.Sleep(time.Second)
					return
				}

				for _, stream := range streams {
					for _, message := range stream.Messages {
						eventType, _ := message.Values["type"].(string)
						payloadStr, _ := message.Values["payload"].(string)
						metrics.QueueConsumedTotal.WithLabelValues(eventType+"_score", "readgroup").Inc()

						var event Event
						event.Type = eventType
						event.Attempt = 0

						if eventType == "ScoreUpdated" {
							var req dto.UpdateScoreEvent
							if err := json.Unmarshal([]byte(payloadStr), &req); err == nil {
								event.Payload = &req
								event.Handler = func(workerCtx context.Context, p any) error {
									updateReq, _ := p.(*dto.UpdateScoreEvent)
									if err := q.scoreRepo.UpdateScore(workerCtx, updateReq.GameID, updateReq.PlayerID, updateReq.Score); err != nil {
										return err
									}
									return nil
								}
								event.Ack = func(workerCtx context.Context) error {
									return q.rdb.XAck(workerCtx, consts.ScoreEvents, consts.ScoreConsumerGroup, message.ID).Err()
								}
								events <- event
							}
						}
					}
				}
			}()
		}
	}()

	return events
}
