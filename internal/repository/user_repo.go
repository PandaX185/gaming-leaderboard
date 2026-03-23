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

type UserRepository interface {
	Insert(context.Context, *dto.CreateUserRequest) (*dto.UserResponse, error)
	UpdateScore(context.Context, string, int) (*dto.UserResponse, error)
	GetByID(context.Context, string) (*dto.UserResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		db: db,
	}
}

type userRepository struct {
	db *mongo.Database
}

func (r *userRepository) Insert(ctx context.Context, data *dto.CreateUserRequest) (*dto.UserResponse, error) {
	doc := model.User{}.FromDTO(data)
	doc.ID = primitive.NewObjectID()
	_, err := r.db.Collection(consts.UserCollection).InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc.ToResponse(), nil
}

func (r *userRepository) UpdateScore(ctx context.Context, id string, score int) (*dto.UserResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Collection(consts.UserCollection).UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$inc": bson.M{"score": score},
		"$set": bson.M{"updated_at": time.Now()},
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = r.db.Collection(consts.UserCollection).FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

func (r *userRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	skip := (params.Page - 1) * params.PageSize
	cursor, err := r.db.Collection(consts.UserCollection).Find(ctx, bson.M{}, options.Find().SetSkip(int64(skip)).SetLimit(int64(params.PageSize)))
	if err != nil {
		return nil, err
	}
	var users []model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	var responses []*dto.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}
	count, err := r.db.Collection(consts.UserCollection).CountDocuments(ctx, bson.M{})
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
