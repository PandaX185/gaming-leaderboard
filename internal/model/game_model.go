package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Game struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name,unique"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
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
		ID:        g.ID.Hex(),
		Name:      g.Name,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}
