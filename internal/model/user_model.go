package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username,unique"`
	Score     int                `bson:"score"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

func (u User) FromDTO(data *dto.CreateUserRequest) User {
	return User{
		Username:  data.Username,
		Score:     data.Score,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u User) ToResponse() *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID.Hex(),
		Username:  u.Username,
		Score:     u.Score,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
