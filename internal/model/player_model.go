package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Player struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username,unique"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

func (u Player) FromDTO(data *dto.CreatePlayerRequest) Player {
	return Player{
		Username:  data.Username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u Player) ToResponse() *dto.PlayerResponse {
	return &dto.PlayerResponse{
		ID:        u.ID.Hex(),
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
