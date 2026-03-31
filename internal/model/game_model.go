package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Game struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name,unique"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
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
