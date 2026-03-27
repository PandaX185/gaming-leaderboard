package model

import (
	"context"
	"gaming-leaderboard/internal/consts"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes(ctx context.Context, db *mongo.Database) []interface{} {
	playerIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
	}
	_, err := db.Collection(consts.PlayerCollection).Indexes().CreateMany(ctx, playerIndexes)
	if err != nil {
		panic("Error creating player index: " + err.Error())
	}

	gameIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
	}
	_, err = db.Collection(consts.GameCollection).Indexes().CreateMany(ctx, gameIndexes)
	if err != nil {
		panic("Error creating game index: " + err.Error())
	}

	scoreIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "score", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "player_id", Value: 1},
				{Key: "game_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err = db.Collection(consts.PlayerGameCollection).Indexes().CreateMany(ctx, scoreIndexes)
	if err != nil {
		panic("Error creating score index: " + err.Error())
	}

	return []interface{}{playerIndexes, gameIndexes, scoreIndexes}
}
