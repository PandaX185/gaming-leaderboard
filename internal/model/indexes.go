package model

import (
	"context"
	"gaming-leaderboard/internal/consts"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes(ctx context.Context, db *mongo.Database) []interface{} {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := db.Collection(consts.UserCollection).Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic("Error creating index: " + err.Error())
	}
	return []interface{}{indexModel}
}
