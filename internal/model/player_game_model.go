package model

import (
	"gaming-leaderboard/internal/dto"
	"time"
)

type PlayerGame struct {
	PlayerID  string    `db:"player_id" json:"player_id"`
	GameID    string    `db:"game_id" json:"game_id"`
	Score     int       `db:"score" json:"score"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (p PlayerGame) ToResponse() *dto.ScoreResponse {
	return &dto.ScoreResponse{
		PlayerID:  p.PlayerID,
		GameID:    p.GameID,
		Score:     p.Score,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
