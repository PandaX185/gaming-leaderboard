package model

import (
	"gaming-leaderboard/internal/dto"
	"time"
)

type Player struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u Player) FromDTO(data *dto.CreatePlayerRequest) Player {
	return Player{
		Username:  data.Username,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}
}

func (u Player) ToResponse() *dto.PlayerResponse {
	return &dto.PlayerResponse{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
