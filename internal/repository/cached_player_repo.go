package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	playerByIDCacheKey = "player:by-id:%s"
	playersListVersion = "players:list:version"
)

type cachedPlayerRepository struct {
	base PlayerRepository
	rdb  *redis.Client
	ttl  time.Duration
}

func NewCachedPlayerRepository(base PlayerRepository, rdb *redis.Client) PlayerRepository {
	return &cachedPlayerRepository{
		base: base,
		rdb:  rdb,
		ttl:  5 * time.Minute,
	}
}

func (r *cachedPlayerRepository) Insert(ctx context.Context, data *dto.CreatePlayerRequest) error {
	err := r.base.Insert(ctx, data)
	if err != nil {
		return err
	}

	playerResp := model.Player{}.FromDTO(data).ToResponse()
	cacheKey := fmt.Sprintf(playerByIDCacheKey, playerResp.ID)
	payload, _ := json.Marshal(playerResp)
	_ = r.rdb.Set(ctx, cacheKey, payload, r.ttl).Err()

	_ = r.rdb.Incr(ctx, "players:total_count").Err()

	r.bumpListVersion(ctx)
	return nil
}

func (r *cachedPlayerRepository) UpdateScore(ctx context.Context, data *dto.UpdateScoreRequest) error {
	err := r.base.UpdateScore(ctx, data)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf(playerByIDCacheKey, data.PlayerID)
	_ = r.rdb.Del(ctx, cacheKey).Err()

	return nil
}

func (r *cachedPlayerRepository) Count(ctx context.Context) (int, error) {
	cached, err := r.rdb.Get(ctx, "players:total_count").Result()
	if err == nil {
		count, err := strconv.Atoi(cached)
		if err == nil {
			return count, nil
		}
	}

	count, err := r.base.Count(ctx)
	if err != nil {
		return 0, err
	}

	_ = r.rdb.Set(ctx, "players:total_count", strconv.Itoa(count), r.ttl).Err()
	return count, nil
}

func (r *cachedPlayerRepository) GetByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	cacheKey := fmt.Sprintf(playerByIDCacheKey, id)
	cached, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		result := &dto.PlayerResponse{}
		if err := json.Unmarshal([]byte(cached), result); err == nil {
			return result, nil
		}
	}

	result, err := r.base.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(result)
	if err == nil {
		_ = r.rdb.Set(ctx, cacheKey, payload, r.ttl).Err()
	}
	return result, nil
}

func (r *cachedPlayerRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	sortField := "updated_at"
	if params.Sort != "" {
		sortField = params.Sort
	}
	sortOrder := -1
	if params.Order == "asc" {
		sortOrder = 1
	}

	version := r.getListVersion(ctx)
	cacheKey := fmt.Sprintf(
		"players:list:v=%d:page=%d:size=%d:sort=%s:order=%d",
		version,
		params.Page,
		params.PageSize,
		sortField,
		sortOrder,
	)

	cached, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		result := &dto.PaginatedResponse{}
		if err := json.Unmarshal([]byte(cached), result); err == nil {
			return result, nil
		}
	}

	result, err := r.base.GetAll(ctx, params)
	if err != nil {
		return nil, err
	}
	result.TotalItems, _ = r.Count(ctx)
	totalPages := result.TotalItems / params.PageSize
	if result.TotalItems%params.PageSize != 0 {
		totalPages++
	}
	result.TotalPages = totalPages
	result.HasNext = params.Page < totalPages

	payload, err := json.Marshal(result)
	if err == nil {
		_ = r.rdb.Set(ctx, cacheKey, payload, r.ttl).Err()
	}
	return result, nil
}

func (r *cachedPlayerRepository) getListVersion(ctx context.Context) int64 {
	value, err := r.rdb.Get(ctx, playersListVersion).Result()
	if err != nil {
		return 1
	}
	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 1
	}
	return version
}

func (r *cachedPlayerRepository) bumpListVersion(ctx context.Context) {
	_, err := r.rdb.Incr(ctx, playersListVersion).Result()
	if err != nil {
		_ = r.rdb.Set(ctx, playersListVersion, "2", 0).Err()
	}
}
