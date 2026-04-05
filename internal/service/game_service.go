package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type GameService struct {
	repo  repository.GameRepository
	cache repository.LeaderboardCache
}

func NewGameService(repo repository.GameRepository, cache repository.LeaderboardCache) *GameService {
	return &GameService{
		repo:  repo,
		cache: cache,
	}
}

func (s *GameService) CreateGame(ctx context.Context, data *dto.CreateGameRequest) (*dto.GameResponse, error) {
	res, err := s.repo.Insert(ctx, data)
	if err != nil {
		return nil, err
	}

	if err = s.cache.IncrementGameCount(ctx); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *GameService) GetGameByID(ctx context.Context, id string) (*dto.GameResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GameService) GetAllGames(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	res, err := s.repo.GetAll(ctx, params)
	if err != nil {
		return nil, err
	}

	res.TotalItems, err = s.cache.GetTotalGamesCount(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *GameService) GetGameScores(ctx context.Context, id string, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetScores(ctx, id, params)
}

func (s *GameService) UpdateGame(ctx context.Context, id string, data *dto.UpdateGameRequest) (*dto.GameResponse, error) {
	return s.repo.Update(ctx, id, data)
}

func (s *GameService) DeleteGame(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
