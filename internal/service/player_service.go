package service

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type PlayerService struct {
	repo    repository.PlayerRepository
	cache   repository.LeaderboardCache
	playerQ queue.IQueue
}

func NewPlayerService(repo repository.PlayerRepository, playerQ queue.IQueue, cache repository.LeaderboardCache) *PlayerService {
	return &PlayerService{
		repo:    repo,
		playerQ: playerQ,
		cache:   cache,
	}
}

func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hashed)
}

func (s *PlayerService) CreatePlayer(ctx context.Context, data *dto.CreatePlayerRequest) (*dto.PlayerResponse, error) {
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.Password = hashPassword(data.Password)

	generatedID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	data.ID = generatedID.String()

	if err := s.playerQ.PublishEvent(ctx, queue.Event{
		Type:    consts.PlayerCreatedEvent,
		Payload: data,
		Attempt: 0,
	}); err != nil {
		return nil, err
	}

	return &dto.PlayerResponse{
		ID:        data.ID,
		Username:  data.Username,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}, nil
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) GetAllPlayers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	res, err := s.repo.GetAll(ctx, params)
	if err != nil {
		return nil, err
	}

	res.TotalItems, err = s.cache.GetTotalPlayersCount(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
