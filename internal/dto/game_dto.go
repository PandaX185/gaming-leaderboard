package dto

import "time"

type CreateGameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50" example:"Space Invaders"`
}

type UpdateGameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50" example:"Space Invaders Deluxe"`
}

type GameResponse struct {
	ID        string    `json:"id" example:"60c72b2f9b1d4c3a5f8e4b1"`
	Name      string    `json:"name" example:"Space Invaders"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GameScoreResponse struct {
	PlayerID   string    `json:"player_id" example:"60c72b2f9b1d4c3a5f8e4b1"`
	PlayerName string    `json:"player_name" example:"player123"`
	Score      int       `json:"score" example:"100"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
