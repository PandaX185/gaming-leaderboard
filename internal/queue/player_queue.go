package queue

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type PlayerEvent struct {
	Type    string
	Payload any
	Handler func(ctx context.Context, payload any) error
	Attempt int
}

type IPlayerQueue interface {
	PublishPlayerCreated(ctx context.Context, player *dto.CreatePlayerRequest) error
	PublishPlayerScoreUpdated(ctx context.Context, data *dto.UpdateScoreRequest) error
	GetEvents() chan PlayerEvent
}

type PlayerQueue struct {
	events chan PlayerEvent
	repo   repository.PlayerRepository
}

func NewPlayerQueue(r repository.PlayerRepository) *PlayerQueue {
	return &PlayerQueue{
		events: make(chan PlayerEvent, 1024),
		repo:   r,
	}
}

func (q *PlayerQueue) PublishPlayerCreated(ctx context.Context, player *dto.CreatePlayerRequest) error {
	event := PlayerEvent{
		Type:    "PlayerCreated",
		Payload: player,
		Handler: func(workerCtx context.Context, payload any) error {
			return q.repo.Insert(workerCtx, payload.(*dto.CreatePlayerRequest))
		},
		Attempt: 0,
	}
	q.events <- event
	return nil
}

func (q *PlayerQueue) PublishPlayerScoreUpdated(ctx context.Context, data *dto.UpdateScoreRequest) error {
	event := PlayerEvent{
		Type:    "PlayerScoreUpdated",
		Payload: data,
		Handler: func(workerCtx context.Context, payload any) error {
			return q.repo.UpdateScore(workerCtx, payload.(*dto.UpdateScoreRequest))
		},
		Attempt: 0,
	}
	q.events <- event
	return nil
}

func (q *PlayerQueue) GetEvents() chan PlayerEvent {
	return q.events
}
