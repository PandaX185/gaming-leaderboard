package service

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, data *dto.CreateUserRequest) (*dto.UserResponse, error) {
	return s.repo.Insert(ctx, data)
}

func (s *UserService) UpdateUserScore(ctx context.Context, id string, score int) (*dto.UserResponse, error) {
	return s.repo.UpdateScore(ctx, id, score)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetAllUsers(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	return s.repo.GetAll(ctx, params)
}
