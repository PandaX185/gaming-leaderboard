package model

import (
	"gaming-leaderboard/internal/dto"
	"time"
)

type Game struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (g Game) FromCreateDTO(data *dto.CreateGameRequest) Game {
	return Game{
		Name:      data.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (g Game) ToResponse() *dto.GameResponse {
	return &dto.GameResponse{
		ID:        g.ID,
		Name:      g.Name,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
