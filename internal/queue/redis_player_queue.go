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

	"gaming-leaderboard/metrics"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ConsumerGroup = "worker-group"
	ReclaimTime   = time.Minute * 1
)

type RedisPlayerQueue struct {
	rdb              *redis.Client
	repo             repository.PlayerRepository
	leaderboardCache repository.LeaderboardCache
	consumerID       string
}

func NewRedisPlayerQueue(rdb *redis.Client, repo repository.PlayerRepository, leaderboardCache repository.LeaderboardCache) IQueue {
	err := rdb.XGroupCreateMkStream(context.Background(), consts.ScoreEvents, ConsumerGroup, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "exists") {
		log.Printf("Error creating consumer group: %v", err)
	}

	hostname, _ := os.Hostname()
	consumerID := hostname + "-" + uuid.NewString()

	go func() {
		for {
			length, err := rdb.XLen(context.Background(), consts.ScoreEvents).Result()
			if err == nil {
				metrics.QueueSize.Set(float64(length))
				metrics.QueueSizeByStream.WithLabelValues(consts.ScoreEvents).Set(float64(length))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return &RedisPlayerQueue{
		rdb:              rdb,
		repo:             repo,
		leaderboardCache: leaderboardCache,
		consumerID:       consumerID,
	}
}

func (q *RedisPlayerQueue) PublishEvent(ctx context.Context, data any) error {
	var eventType string
	switch data.(type) {
	case *dto.CreatePlayerRequest:
		eventType = "PlayerCreated"
	case *dto.UpdateScoreEvent:
		eventType = "ScoreUpdated"
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
		metrics.QueuePublishedTotal.WithLabelValues(eventType, "error").Inc()
		return err
	}
	metrics.QueuePublishedTotal.WithLabelValues(eventType, "success").Inc()
	return nil
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
					Streams:  []string{consts.ScoreEvents, ">"},
					Count:    50,
					Block:    time.Second * 5,
				}).Result()

				if err == redis.Nil {
					return
				} else if err != nil {
					log.Printf("Error reading from Redis Group: %v", err)
					metrics.QueueReadErrorsTotal.WithLabelValues("xreadgroup").Inc()
					time.Sleep(time.Second)
					return
				}

				for _, stream := range streams {
					q.processMessages(stream.Messages, events, "readgroup")
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
					Stream:   consts.ScoreEvents,
					Group:    ConsumerGroup,
					Consumer: q.consumerID,
					MinIdle:  ReclaimTime,
					Start:    "0-0",
					Count:    50,
				}).Result()

				if err != nil && err != redis.Nil {
					log.Printf("Error during XAutoClaim: %v", err)
					metrics.QueueReadErrorsTotal.WithLabelValues("xautoclaim").Inc()
					time.Sleep(time.Second)
				} else if len(messages) > 0 {
					log.Printf("Reclaimed %d pending messages", len(messages))
					q.processMessages(messages, events, "autoclaim")
				}

				time.Sleep(time.Second * 10)
			}()
		}
	}()

	return events
}

func (q *RedisPlayerQueue) processMessages(messages []redis.XMessage, events chan Event, source string) {
	for _, message := range messages {
		eventType, _ := message.Values["type"].(string)
		payloadStr, _ := message.Values["payload"].(string)
		metrics.QueueConsumedTotal.WithLabelValues(eventType, source).Inc()

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
			var req dto.UpdateScoreEvent
			if err := json.Unmarshal([]byte(payloadStr), &req); err == nil {
				event.Payload = &req
				dbHandler = func(workerCtx context.Context, p any) error {
					e := p.(*dto.UpdateScoreEvent)
					err := q.repo.UpdateScore(workerCtx, &dto.UpdateScoreRequest{
						PlayerID: e.PlayerID,
						GameID:   e.GameID,
						Score:    e.Score,
					})
					if err != nil {
						return err
					}
					if q.leaderboardCache == nil {
						return nil
					}
					return q.leaderboardCache.IncrementScore(workerCtx, e.GameID, e.PlayerID, e.Score)
				}
			}
		}

		if dbHandler != nil {
			msgID := message.ID
			event.Handler = func(workerCtx context.Context, p any) error {
				err := dbHandler(workerCtx, p)
				if err == nil {
					ackErr := q.rdb.XAck(workerCtx, consts.ScoreEvents, ConsumerGroup, msgID).Err()
					if ackErr != nil {
						log.Printf("Failed to XACK message %s: %v", msgID, ackErr)
						metrics.QueueAckTotal.WithLabelValues(eventType, "error").Inc()
					} else {
						metrics.QueueAckTotal.WithLabelValues(eventType, "success").Inc()
						if eventType == "ScoreUpdated" {
							scoreEvent, ok := p.(*dto.UpdateScoreEvent)
							if ok {
								q.emitScoreDeltaUpdate(workerCtx, scoreEvent)
							}
						}
					}
				}
				return err
			}
			event.Ack = func(workerCtx context.Context) error {
				ackErr := q.rdb.XAck(workerCtx, consts.ScoreEvents, ConsumerGroup, msgID).Err()
				if ackErr != nil {
					metrics.QueueAckTotal.WithLabelValues(eventType, "error").Inc()
					return ackErr
				}
				metrics.QueueAckTotal.WithLabelValues(eventType, "success").Inc()
				return nil
			}
			events <- event
		}
	}
}

func (q *RedisPlayerQueue) emitScoreDeltaUpdate(ctx context.Context, scoreEvent *dto.UpdateScoreEvent) {
	leaderboardKey := repository.LeaderboardKey(scoreEvent.GameID)
	score, scoreErr := q.rdb.ZScore(ctx, leaderboardKey, scoreEvent.PlayerID).Result()
	rank, rankErr := q.rdb.ZRevRank(ctx, leaderboardKey, scoreEvent.PlayerID).Result()
	if scoreErr != nil || rankErr != nil {
		log.Printf("Failed to resolve score/rank for game %s player %s: scoreErr=%v rankErr=%v", scoreEvent.GameID, scoreEvent.PlayerID, scoreErr, rankErr)
		return
	}

	stream := repository.LeaderboardUpdatesStream(scoreEvent.GameID)
	if addErr := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{
			"type":      "score_update",
			"player_id": scoreEvent.PlayerID,
			"score":     score,
			"rank":      rank + 1,
		},
	}).Err(); addErr != nil {
		log.Printf("Failed to append leaderboard stream update for game %s: %v", scoreEvent.GameID, addErr)
	}
}
