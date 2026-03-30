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
	GetScores(context.Context, string, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Update(context.Context, string, *dto.UpdateGameRequest) (*dto.GameResponse, error)
	Delete(context.Context, string) error
}

func NewMongoGameRepository(db *mongo.Database) GameRepository {
	return &mongoGameRepository{
		db: db,
	}
}

type mongoGameRepository struct {
	db *mongo.Database
}

func (r *mongoGameRepository) Insert(ctx context.Context, data *dto.CreateGameRequest) (*dto.GameResponse, error) {
	doc := model.Game{}.FromCreateDTO(data)
	doc.ID = primitive.NewObjectID()

	_, err := r.db.Collection(consts.GameCollection).InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc.ToResponse(), nil
}

func (r *mongoGameRepository) GetByID(ctx context.Context, id string) (*dto.GameResponse, error) {
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

func (r *mongoGameRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
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
	cursor, err := r.db.Collection(consts.GameCollection).Find(ctx, bson.M{}, opts)
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

func (r *mongoGameRepository) GetScores(ctx context.Context, gameID string, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	objID, err := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		return nil, err
	}

	sortField := "score"
	if params.Sort != "" {
		sortField = params.Sort
	}
	sortOrder := -1
	if params.Order == "asc" {
		sortOrder = 1
	}

	skip := (params.Page - 1) * params.PageSize
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(params.PageSize)).SetSort(bson.M{sortField: sortOrder})
	cursor, err := r.db.Collection(consts.PlayerGameCollection).Find(ctx, bson.M{"game_id": objID}, opts)
	if err != nil {
		return nil, err
	}
	var scores []model.PlayerGame
	if err = cursor.All(ctx, &scores); err != nil {
		return nil, err
	}

	playerIDs := make([]string, 0, len(scores))
	seen := make(map[string]struct{}, len(scores))
	for _, score := range scores {
		id := score.PlayerID.Hex()
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		playerIDs = append(playerIDs, id)
	}

	nameByID := make(map[string]string, len(playerIDs))
	if len(playerIDs) > 0 {
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{"$expr": bson.M{"$in": []any{bson.M{"$toString": "$_id"}, playerIDs}}}}},
			{{Key: "$project", Value: bson.M{"_id": 0, "player_id": bson.M{"$toString": "$_id"}, "player_name": bson.M{"$ifNull": []any{"$username", "$username,unique"}}}}},
		}

		cursor, findErr := r.db.Collection(consts.PlayerCollection).Aggregate(ctx, pipeline)
		if findErr == nil {
			defer cursor.Close(ctx)
			for cursor.Next(ctx) {
				var row struct {
					PlayerID   string `bson:"player_id"`
					PlayerName string `bson:"player_name"`
				}
				if decodeErr := cursor.Decode(&row); decodeErr != nil {
					continue
				}
				nameByID[row.PlayerID] = row.PlayerName
			}
		}
	}

	var responses []*dto.GameScoreResponse
	for _, score := range scores {
		playerID := score.PlayerID.Hex()
		playerName := nameByID[playerID]

		responses = append(responses, &dto.GameScoreResponse{
			PlayerID:   playerID,
			PlayerName: playerName,
			Score:      score.Score,
			CreatedAt:  score.CreatedAt,
			UpdatedAt:  score.UpdatedAt,
		})
	}
	count, err := r.db.Collection(consts.PlayerGameCollection).CountDocuments(ctx, bson.M{"game_id": objID})
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

func (r *mongoGameRepository) Update(ctx context.Context, id string, data *dto.UpdateGameRequest) (*dto.GameResponse, error) {
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

func (r *mongoGameRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.db.Collection(consts.GameCollection).DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
