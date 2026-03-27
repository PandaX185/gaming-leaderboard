package db

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Init(uri string) (*mongo.Client, error) {
	return mongo.Connect(options.Client().ApplyURI(uri).SetMaxPoolSize(0))
}
