package model

import (
"time"
"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerGame struct {
ID        primitive.ObjectID `bson:"_id,omitempty"`
PlayerID  primitive.ObjectID `bson:"player_id"`
GameID    primitive.ObjectID `bson:"game_id"`
Score     int                `bson:"score"`
CreatedAt time.Time          `bson:"created_at"`
UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

