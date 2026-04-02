package dto

import (
	"time"
)

type CreatePlayerRequest struct {
	ID        string    `json:"-"`
	Username  string    `json:"username" validate:"required,min=3" example:"player1"`
	Password  string    `json:"password" validate:"required,min=6" example:"secret123"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateScoreRequest struct {
	PlayerID string `json:"player_id" validate:"required"`
	GameID   string `json:"game_id" validate:"required"`
	Score    int    `json:"score" validate:"required,gt=0" example:"1500"`
}

type UpdateScoreEvent struct {
	PlayerID string `json:"player_id"`
	GameID   string `json:"game_id"`
	Score    int    `json:"score"`
}

type PlayerResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username" example:"player1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ScoreUpdated struct {
	Message string `json:"message" example:"Score updated successfully"`
}
