package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/queue"
)

type ScoreService struct {
	scoreQ queue.IQueue
}

func NewScoreService(scoreQ queue.IQueue) *ScoreService {
	return &ScoreService{
		scoreQ: scoreQ,
	}
}

func (s *ScoreService) UpdateScore(ctx context.Context, playerID string, gameID int, score int) (*dto.ScoreUpdated, error) {
	data := &dto.UpdateScoreEvent{
		PlayerID: playerID,
		GameID:   gameID,
		Score:    score,
	}

	if err := s.scoreQ.PublishEvent(ctx, data); err != nil {
		return nil, err
	}

	return &dto.ScoreUpdated{
		Message: "Score update queued successfully",
	}, nil
}
