package model

import (
	"context"
	"gaming-leaderboard/internal/consts"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes(ctx context.Context, db *mongo.Database) []interface{} {
	playerIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "playername", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := db.Collection(consts.PlayerCollection).Indexes().CreateOne(ctx, playerIndexModel)
	if err != nil {
		panic("Error creating player index: " + err.Error())
	}

	gameIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = db.Collection(consts.GameCollection).Indexes().CreateOne(ctx, gameIndexModel)
	if err != nil {
		panic("Error creating game index: " + err.Error())
	}

	return []interface{}{playerIndexModel, gameIndexModel}
}
