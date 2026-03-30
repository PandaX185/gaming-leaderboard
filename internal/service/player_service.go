package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	data = &dto.CreatePlayerRequest{
		ID:        primitive.NewObjectID(),
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

func (s *PlayerService) UpdatePlayerScore(ctx context.Context, id string, gameId string, score int) (*dto.ScoreUpdated, error) {
	player, err := s.GetPlayerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data := &dto.UpdateScoreEvent{
		PlayerID: id,
		Username: player.Username,
		GameID:   gameId,
		Score:    score,
	}

	if err := s.playerQ.PublishEvent(ctx, data); err != nil {
		return nil, err
	}
	return &dto.ScoreUpdated{
		Message: "Score update queued successfully",
	}, nil
}

func (s *PlayerService) GetPlayerByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) GetAllPlayers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}
