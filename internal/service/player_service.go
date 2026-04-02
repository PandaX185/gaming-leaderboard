package service

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
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
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	generatedID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	data.ID = generatedID.String()

	s.playerQ.PublishEvent(ctx, queue.Event{
		Type:    consts.PlayerCreatedEvent,
		Payload: data,
		Handler: func(workerCtx context.Context, p any) error {
			return s.repo.Insert(workerCtx, p.(*dto.CreatePlayerRequest))
		},
		Attempt: 0,
	})

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
	return s.repo.GetAll(ctx, params)
}
