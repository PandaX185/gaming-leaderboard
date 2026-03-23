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
	Insert(context.Context, *dto.CreatePlayerRequest) (*dto.PlayerResponse, error)
	UpdateScore(context.Context, string, string, int) (*dto.PlayerResponse, error)
	GetByID(context.Context, string) (*dto.PlayerResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
}

func NewMongoPlayerRepository(db *mongo.Database) PlayerRepository {
	return &mongoPlayerRepository{
		db: db,
	}
}

type mongoPlayerRepository struct {
	db *mongo.Database
}

func (r *mongoPlayerRepository) Insert(ctx context.Context, data *dto.CreatePlayerRequest) (*dto.PlayerResponse, error) {
	doc := model.Player{}.FromDTO(data)
	doc.ID = primitive.NewObjectID()
	_, err := r.db.Collection(consts.PlayerCollection).InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc.ToResponse(), nil
}

func (r *mongoPlayerRepository) UpdateScore(ctx context.Context, id string, gameId string, score int) (*dto.PlayerResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	gameObjID, err := primitive.ObjectIDFromHex(gameId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"player_id": objID, "game_id": gameObjID}
	update := bson.M{
		"$inc": bson.M{"score": score},
		"$set": bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"created_at": time.Now(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)

	_, err = r.db.Collection(consts.PlayerGameCollection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
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
	count, err := r.db.Collection(consts.PlayerCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	totalPages := int(count) / params.PageSize
	if int(count)%params.PageSize != 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		TotalItems: int(count),
		Items:      responses,
		TotalPages: totalPages,
		Page:       params.Page,
		PageSize:   params.PageSize,
		HasNext:    params.Page*params.PageSize < int(count),
		HasPrev:    params.Page > 1,
	}, nil
}
