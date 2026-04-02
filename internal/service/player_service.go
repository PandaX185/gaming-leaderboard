package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	"time"

	"github.com/google/uuid"
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
	playerId, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	data = &dto.CreatePlayerRequest{
		ID:        playerId.String(),
		Username:  data.Username,
		Password:  data.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.playerQ.PublishEvent(ctx, data); err != nil {
		return nil, err
	}
	return model.Player{}.FromDTO(data).ToResponse(), nil
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) GetAllPlayers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}
