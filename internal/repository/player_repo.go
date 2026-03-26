package repository

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PlayerRepository interface {
	Insert(context.Context, *dto.CreatePlayerRequest) error
	UpdateScore(context.Context, *dto.UpdateScoreRequest) error
	GetByID(context.Context, string) (*dto.PlayerResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Count(context.Context) (int, error)
}

func NewMongoPlayerRepository(db *mongo.Database) PlayerRepository {
	return &mongoPlayerRepository{
		db: db,
	}
}

type mongoPlayerRepository struct {
	db *mongo.Database
}

func (r *mongoPlayerRepository) Insert(ctx context.Context, data *dto.CreatePlayerRequest) error {
	doc := model.Player{}.FromDTO(data)
	_, err := r.db.Collection(consts.PlayerCollection).InsertOne(ctx, doc)
	if err != nil {
		return err
	}
	return nil
}

func (r *mongoPlayerRepository) UpdateScore(ctx context.Context, data *dto.UpdateScoreRequest) error {
	objID, err := primitive.ObjectIDFromHex(data.PlayerID)
	if err != nil {
		return err
	}
	gameObjID, err := primitive.ObjectIDFromHex(data.GameID)
	if err != nil {
		return err
	}

	filter := bson.M{"player_id": objID, "game_id": gameObjID}
	update := bson.M{
		"$inc": bson.M{"score": data.Score},
		"$set": bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"created_at": time.Now(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)

	_, err = r.db.Collection(consts.PlayerGameCollection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	return nil
}

func (r *mongoPlayerRepository) GetByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var player model.Player
	err = r.db.Collection(consts.PlayerCollection).FindOne(ctx, bson.M{"_id": objID}).Decode(&player)
	if err != nil {
		return nil, err
	}
	return player.ToResponse(), nil
}

func (r *mongoPlayerRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	sortField := "updated_at"
	if params.Sort != "" {
		sortField = params.Sort
	}
	sortOrder := -1
	if params.Order == "asc" {
		sortOrder = 1
	}

	skip := (params.Page - 1) * params.PageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(params.PageSize)).SetSort(bson.M{sortField: sortOrder})
	cursor, err := r.db.Collection(consts.PlayerCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	var players []model.Player
	if err = cursor.All(ctx, &players); err != nil {
		return nil, err
	}

	var responses []*dto.PlayerResponse
	for _, player := range players {
		responses = append(responses, player.ToResponse())
	}

	return &dto.PaginatedResponse{
		Items:    responses,
		Page:     params.Page,
		PageSize: params.PageSize,
		HasPrev:  params.Page > 1,
	}, nil
}

func (r *mongoPlayerRepository) Count(ctx context.Context) (int, error) {
	count, err := r.db.Collection(consts.PlayerCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
