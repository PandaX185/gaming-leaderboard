package queue

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
	"gaming-leaderboard/metrics"
	"time"
)

type InMemoryPlayerQueue struct {
	events chan Event
	repo   repository.PlayerRepository
}

func NewInMemoryPlayerQueue(r repository.PlayerRepository) IQueue {
	q := &InMemoryPlayerQueue{
		events: make(chan Event, 1024),
		repo:   r,
	}

	go func() {
		for {
			metrics.QueueSize.Set(float64(len(q.events)))
			metrics.QueueSizeByStream.WithLabelValues("in_memory_players").Set(float64(len(q.events)))
			time.Sleep(5 * time.Second)
		}
	}()

	return q
}

func (q *InMemoryPlayerQueue) PublishEvent(ctx context.Context, data any) error {
	switch v := data.(type) {
	case *dto.CreatePlayerRequest:
		metrics.QueuePublishedTotal.WithLabelValues("PlayerCreated", "success").Inc()
		q.events <- Event{
			Type:    "PlayerCreated",
			Payload: v,
			Handler: func(workerCtx context.Context, payload any) error {
				return q.repo.Insert(workerCtx, payload.(*dto.CreatePlayerRequest))
			},
			Attempt: 0,
		}
	case *dto.UpdateScoreEvent:
		metrics.QueuePublishedTotal.WithLabelValues("ScoreUpdated", "success").Inc()
		q.events <- Event{
			Type:    "ScoreUpdated",
			Payload: v,
			Handler: func(workerCtx context.Context, payload any) error {
				e := payload.(*dto.UpdateScoreEvent)
				return q.repo.UpdateScore(workerCtx, &dto.UpdateScoreRequest{
					PlayerID: e.PlayerID,
					GameID:   e.GameID,
					Score:    e.Score,
				})
			},
			Attempt: 0,
		}
	}
	return nil
}

func (q *InMemoryPlayerQueue) GetEvents() chan Event {
	return q.events
}
