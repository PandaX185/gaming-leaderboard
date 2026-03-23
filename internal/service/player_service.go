package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type PlayerService struct {
	repo repository.PlayerRepository
}

func NewPlayerService(repo repository.PlayerRepository) *PlayerService {
	return &PlayerService{
		repo: repo,
	}
}

func (s *PlayerService) CreatePlayer(ctx context.Context, data *dto.CreatePlayerRequest) (*dto.PlayerResponse, error) {
	return s.repo.Insert(ctx, data)
}

func (s *PlayerService) UpdatePlayerScore(ctx context.Context, id string, gameId string, score int) (*dto.PlayerResponse, error) {
        return s.repo.UpdateScore(ctx, id, gameId, score)
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) GetAllPlayers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}
