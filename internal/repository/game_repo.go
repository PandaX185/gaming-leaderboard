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

type GameRepository interface {
	Insert(context.Context, *dto.CreateGameRequest) (*dto.GameResponse, error)
	GetByID(context.Context, string) (*dto.GameResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Update(context.Context, string, *dto.UpdateGameRequest) (*dto.GameResponse, error)
	Delete(context.Context, string) error
}

func NewGameRepository(db *mongo.Database) GameRepository {
	return &gameRepository{
		db: db,
	}
}

type gameRepository struct {
	db *mongo.Database
}

func (r *gameRepository) Insert(ctx context.Context, data *dto.CreateGameRequest) (*dto.GameResponse, error) {
	doc := model.Game{}.FromCreateDTO(data)
	doc.ID = primitive.NewObjectID()

	_, err := r.db.Collection(consts.GameCollection).InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc.ToResponse(), nil
}

func (r *gameRepository) GetByID(ctx context.Context, id string) (*dto.GameResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var game model.Game
	err = r.db.Collection(consts.GameCollection).FindOne(ctx, bson.M{"_id": objID}).Decode(&game)
	if err != nil {
		return nil, err
	}
	return game.ToResponse(), nil
}

func (r *gameRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	skip := (params.Page - 1) * params.PageSize
	cursor, err := r.db.Collection(consts.GameCollection).Find(ctx, bson.M{}, options.Find().SetSkip(int64(skip)).SetLimit(int64(params.PageSize)))
	if err != nil {
		return nil, err
	}
	var games []model.Game
	if err = cursor.All(ctx, &games); err != nil {
		return nil, err
	}

	var responses []*dto.GameResponse
	for _, game := range games {
		responses = append(responses, game.ToResponse())
	}
	count, err := r.db.Collection(consts.GameCollection).CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	totalPages := int(count) / params.PageSize
	if int(count)%params.PageSize > 0 {
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

func (r *gameRepository) Update(ctx context.Context, id string, data *dto.UpdateGameRequest) (*dto.GameResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$set": bson.M{
			"name":       data.Name,
			"updated_at": time.Now(),
		},
	}

	_, err = r.db.Collection(consts.GameCollection).UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *gameRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.db.Collection(consts.GameCollection).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
