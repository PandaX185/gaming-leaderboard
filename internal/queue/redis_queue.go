package queue

import (
	"context"
	"encoding/json"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/metrics"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	rdb        *redis.Client
	consumerId string
}

func NewRedisQueue(repo any, rdb *redis.Client) IQueue {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.
		XGroupCreateMkStream(ctx, consts.ScoreEvents, consts.ConsumerGroup, "0-0").
		Err(); err != nil && err != redis.Nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			log.Error("Failed to create Redis consumer group: %v", err)
		}
	}

	hostname, _ := os.Hostname()
	consumerID := hostname + "-" + uuid.NewString()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		for {
			length, err := rdb.XLen(ctx, consts.ScoreEvents).Result()
			if err == nil {
				metrics.QueueSize.Set(float64(length))
				metrics.QueueSizeByStream.WithLabelValues(consts.ScoreEvents).Set(float64(length))
			}
			time.Sleep(5 * time.Second)
		}
	}()

	return &RedisQueue{
		rdb:        rdb,
		consumerId: consumerID,
	}
}

func (q *RedisQueue) PublishEvent(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: consts.ScoreEvents,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]interface{}{"data": string(data)},
	}).Err(); err != nil {
		if !strings.Contains(err.Error(), "BUSYGROUP") {
			return err
		}
	}
	metrics.QueuePublishedTotal.WithLabelValues(event.Type, "success").Inc()
	return nil
}

func (q *RedisQueue) GetEvents() chan Event {
	events := make(chan Event)

	go q.readStream(events)

	go q.autoClaim(events)

	return events
}

func (q *RedisQueue) readStream(events chan Event) {
	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
			defer cancel()

			streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    consts.ConsumerGroup,
				Consumer: q.consumerId,
				Streams:  []string{consts.ScoreEvents, ">"},
				Count:    50,
				Block:    time.Second * 5,
			}).Result()

			if err == redis.Nil {
				return
			} else if err != nil {
				log.Error("Error reading from Redis Group: %v", err)
				metrics.QueueReadErrorsTotal.WithLabelValues("xreadgroup").Inc()
				time.Sleep(time.Second)
				return
			}

			for _, stream := range streams {
				q.processMessages(stream.Messages, events, "readgroup")
			}
		}()
	}
}

func (q *RedisQueue) autoClaim(events chan Event) {
	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
				Stream:   consts.ScoreEvents,
				Group:    consts.ConsumerGroup,
				Consumer: q.consumerId,
				MinIdle:  consts.ReclaimTime,
				Start:    "0-0",
				Count:    50,
			}).Result()

			if err != nil && err != redis.Nil {
				log.Error("Error during XAutoClaim: %v", err)
				metrics.QueueReadErrorsTotal.WithLabelValues("xautoclaim").Inc()
				time.Sleep(time.Second)
			} else if len(messages) > 0 {
				log.Info("Reclaimed %d pending messages", len(messages))
				q.processMessages(messages, events, "autoclaim")
			}

			time.Sleep(time.Second * 5)
		}()
	}
}

func (q *RedisQueue) processMessages(messages []redis.XMessage, events chan Event, source string) {
	for _, message := range messages {
		dataRaw, ok := message.Values["data"]
		if !ok || dataRaw == nil {
			log.Error("Missing data in message: %v", message.ID)
			metrics.QueueReadErrorsTotal.WithLabelValues(source).Inc()
			continue
		}
		dataStr, ok := dataRaw.(string)
		if !ok {
			log.Error("data is not a string in message: %v", message.ID)
			metrics.QueueReadErrorsTotal.WithLabelValues(source).Inc()
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
			log.Error("Failed to unmarshal event: %v", err)
			metrics.QueueReadErrorsTotal.WithLabelValues(source).Inc()
			continue
		}

		event.Ack = func(ctx context.Context) error {
			err := q.rdb.XAck(ctx, consts.ScoreEvents, consts.ConsumerGroup, message.ID).Err()
			if err != nil {
				metrics.QueueAckTotal.WithLabelValues(event.Type, "error").Inc()
			} else {
				metrics.QueueAckTotal.WithLabelValues(event.Type, "success").Inc()
			}
			return err
		}

		events <- event
		metrics.QueueConsumedTotal.WithLabelValues(event.Type, source).Inc()
	}
}
