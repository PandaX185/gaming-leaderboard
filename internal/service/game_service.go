package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type GameService struct {
	repo repository.GameRepository
}

func NewGameService(repo repository.GameRepository) *GameService {
	return &GameService{
		repo: repo,
	}
}

func (s *GameService) CreateGame(ctx context.Context, data *dto.CreateGameRequest) (*dto.GameResponse, error) {
	return s.repo.Insert(ctx, data)
}

func (s *GameService) GetGameByID(ctx context.Context, id int) (*dto.GameResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GameService) GetAllGames(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}

func (s *GameService) GetGameScores(ctx context.Context, id int, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetScores(ctx, id, params)
}

func (s *GameService) UpdateGame(ctx context.Context, id int, data *dto.UpdateGameRequest) (*dto.GameResponse, error) {
	return s.repo.Update(ctx, id, data)
}

func (s *GameService) DeleteGame(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
