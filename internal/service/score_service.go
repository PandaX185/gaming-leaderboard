package service

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
)

type ScoreService struct {
	repo repository.ScoreRepository
	scoreQ queue.IQueue
}

func NewScoreService(repo repository.ScoreRepository, scoreQ queue.IQueue) *ScoreService {
	return &ScoreService{
		repo: repo,
		scoreQ: scoreQ,
	}
}

func (s *ScoreService) UpdateScore(ctx context.Context, playerID string, gameID string, score int) (*dto.ScoreUpdated, error) {
	data := &dto.UpdateScoreEvent{
		PlayerID: playerID,
		GameID:   gameID,
		Score:    score,
	}

	if err := s.scoreQ.PublishEvent(ctx, queue.Event{
		Type: consts.ScoreUpdatedEvent,
		Payload: data,
		Handler: func(workerCtx context.Context, p any) error {
			return s.repo.UpdateScore(workerCtx, playerID, gameID, score)
		},
	}); err != nil {
		return nil, err
	}

	return &dto.ScoreUpdated{
		Message: "Score update queued successfully",
	}, nil
}
