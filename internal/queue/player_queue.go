package queue

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type PlayerQueue struct {
	events chan Event
	repo   repository.PlayerRepository
}

func NewPlayerQueue(r repository.PlayerRepository) *PlayerQueue {
	return &PlayerQueue{
		events: make(chan Event, 1024),
		repo:   r,
	}
}

func (q *PlayerQueue) PublishEvent(ctx context.Context, data any) error {
	switch v := data.(type) {
	case *dto.CreatePlayerRequest:
		q.events <- Event{
			Type:    "PlayerCreated",
			Payload: v,
			Handler: func(workerCtx context.Context, payload any) error {
				return q.repo.Insert(workerCtx, payload.(*dto.CreatePlayerRequest))
			},
			Attempt: 0,
		}
	case *dto.UpdateScoreRequest:
		q.events <- Event{
			Type:    "ScoreUpdated",
			Payload: v,
			Handler: func(workerCtx context.Context, payload any) error {
				return q.repo.UpdateScore(workerCtx, payload.(*dto.UpdateScoreRequest))
			},
			Attempt: 0,
		}
	}
	return nil
}

func (q *PlayerQueue) GetEvents() chan Event {
	return q.events
}
