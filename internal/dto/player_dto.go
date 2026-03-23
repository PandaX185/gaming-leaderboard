package dto

import "time"

type CreatePlayerRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20" example:"player1"`
	Password string `json:"password" validate:"required,min=6" example:"secret123"`
}

type UpdateScoreRequest struct {
	GameID string `json:"game_id" validate:"required" example:"60c72b2f9b1d4c3a5f8e4b1"`
	Score  int    `json:"score" validate:"required,gt=0" example:"1500"`
}

type PlayerResponse struct {
	ID        string    `json:"id" example:"60c72b2f9b1d4c3a5f8e4b1"`
	Username  string    `json:"username" example:"player1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
