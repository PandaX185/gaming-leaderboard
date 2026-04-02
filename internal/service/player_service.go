package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	"time"
)

type PlayerService struct {
	repo    repository.PlayerRepository
	playerQ queue.IQueue
}

func NewPlayerService(repo repository.PlayerRepository, playerQ queue.IQueue) *PlayerService {
	return &PlayerService{
		repo:    repo,
		playerQ: playerQ,
	}
}

func (s *PlayerService) CreatePlayer(ctx context.Context, data *dto.CreatePlayerRequest) (*dto.PlayerResponse, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	if err := s.playerQ.PublishEvent(ctx, data); err != nil {
		return nil, err
	}
	return model.Player{}.FromDTO(data).ToResponse(), nil
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id int) (*dto.PlayerResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) GetAllPlayers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}
