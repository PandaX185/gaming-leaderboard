package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Player struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Username  string        `bson:"username,unique"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
}

func (u Player) FromDTO(data *dto.CreatePlayerRequest) Player {
	return Player{
		ID:        data.ID,
		Username:  data.Username,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
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
