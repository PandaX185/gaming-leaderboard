package repository

import (
	"context"
	"errors"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"time"

	
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
	repo := &mongoPlayerRepository{
		db:            db,
		insertBatchCh: make(chan insertJob, 2048),
		batchSize:     100,
		flushInterval: 500 * time.Millisecond,
	}

	go repo.startInsertBatchWorker()

	return repo
}

type mongoPlayerRepository struct {
	db            *mongo.Database
	insertBatchCh chan insertJob
	batchSize     int
	flushInterval time.Duration
}

type insertJob struct {
	doc    model.Player
	result chan error
}

func (r *mongoPlayerRepository) Insert(ctx context.Context, data *dto.CreatePlayerRequest) error {
	doc := model.Player{}.FromDTO(data)
	job := insertJob{
		doc:    doc,
		result: make(chan error, 1),
	}

	select {
	case r.insertBatchCh <- job:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-job.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *mongoPlayerRepository) startInsertBatchWorker() {
	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()

	batch := make([]insertJob, 0, r.batchSize)

	flush := func(items []insertJob) {
		if len(items) == 0 {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		models := make([]mongo.WriteModel, 0, len(items))
		for _, item := range items {
			models = append(models, mongo.NewInsertOneModel().SetDocument(item.doc))
		}

		errByIndex := make([]error, len(items))
		_, err := r.db.Collection(consts.PlayerCollection).BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
		if err != nil {
			if bulkErr, ok := errors.AsType[mongo.BulkWriteException](err); ok {
				for _, writeErr := range bulkErr.WriteErrors {
					if writeErr.Index >= 0 && writeErr.Index < len(errByIndex) {
						errByIndex[writeErr.Index] = writeErr
					}
				}
				for i := range errByIndex {
					if errByIndex[i] == nil && bulkErr.WriteConcernError != nil {
						errByIndex[i] = bulkErr.WriteConcernError
					}
				}
			} else {
				for i := range errByIndex {
					errByIndex[i] = err
				}
			}
		}

		for i, item := range items {
			item.result <- errByIndex[i]
			close(item.result)
		}
	}

	for {
		select {
		case job := <-r.insertBatchCh:
			batch = append(batch, job)
			if len(batch) >= r.batchSize {
				flush(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				flush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (r *mongoPlayerRepository) UpdateScore(ctx context.Context, data *dto.UpdateScoreRequest) error {
	objID, err := bson.ObjectIDFromHex(data.PlayerID)
	if err != nil {
		return err
	}
	gameObjID, err := bson.ObjectIDFromHex(data.GameID)
	if err != nil {
		return err
	}

	filter := bson.M{"player_id": objID, "game_id": gameObjID}
	update := bson.M{
		"$inc": bson.M{"score": data.Score},
		"$set": bson.M{"updated_at": time.Now()},
		"$setOnInsert": bson.M{
			"_id":        bson.NewObjectID(),
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
	objID, err := bson.ObjectIDFromHex(id)
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
