package dto

import "time"

type CreateGameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50" example:"Space Invaders"`
}

type UpdateGameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50" example:"Space Invaders Deluxe"`
}

type GameResponse struct {
	ID        string       `json:"id"`
	Name      string    `json:"name" example:"Space Invaders"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}