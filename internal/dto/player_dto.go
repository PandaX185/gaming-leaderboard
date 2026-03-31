package dto

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CreatePlayerRequest struct {
	ID        bson.ObjectID `json:"-"`
	Username  string        `json:"username" validate:"required,min=3" example:"player1"`
	Password  string        `json:"password" validate:"required,min=6" example:"secret123"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type UpdateScoreRequest struct {
	PlayerID string `json:"-"`
	GameID   string `json:"game_id" validate:"required" example:"60c72b2f9b1d4c3a5f8e4b1"`
	Score    int    `json:"score" validate:"required,gt=0" example:"1500"`
}

type UpdateScoreEvent struct {
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
	GameID   string `json:"game_id"`
	Score    int    `json:"score"`
}

type PlayerResponse struct {
	ID        string    `json:"id" example:"60c72b2f9b1d4c3a5f8e4b1"`
	Username  string    `json:"username" example:"player1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ScoreUpdated struct {
	Message string `json:"message" example:"Score updated successfully"`
}
