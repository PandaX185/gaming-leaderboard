package model

import (
	"gaming-leaderboard/internal/dto"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PlayerGame struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	PlayerID  bson.ObjectID `bson:"player_id"`
	GameID    bson.ObjectID `bson:"game_id"`
	Score     int           `bson:"score"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
}

func (p PlayerGame) ToResponse() *dto.ScoreResponse {
	return &dto.ScoreResponse{
		PlayerID:  p.PlayerID.Hex(),
		Score:     p.Score,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
