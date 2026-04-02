package dto

import "time"

type ScoreResponse struct {
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	Score     int       `json:"score" example:"100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
