package repository

import (
	"context"
	"errors"
	"gaming-leaderboard/internal/dto"
	internalErrors "gaming-leaderboard/internal/errors"
	"gaming-leaderboard/internal/log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlayerRepository interface {
	Insert(context.Context, *dto.CreatePlayerRequest) (int, error)
	UpdateScore(context.Context, *dto.UpdateScoreRequest) error
	GetByID(context.Context, int) (*dto.PlayerResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Count(context.Context) (int, error)
}

type postgresPlayerRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPlayerRepository(db *pgxpool.Pool) PlayerRepository {
	return &postgresPlayerRepository{db: db}
}

func (r *postgresPlayerRepository) Insert(ctx context.Context, req *dto.CreatePlayerRequest) (int, error) {
	log.Info("PlayerRepository Insert called username=%s", req.Username)
	var id int
	err := r.db.QueryRow(ctx, "insert into players(username, password, created_at, updated_at) values ($1, $2, $3, $4) returning id", req.Username, req.Password, req.CreatedAt, req.UpdatedAt).Scan(&id)
	if err != nil {
		log.Error("PlayerRepository Insert failed username=%s err=%v", req.Username, err)
		return 0, err
	}
	log.Info("PlayerRepository Insert success id=%d username=%s", id, req.Username)
	return id, nil
}

func (r *postgresPlayerRepository) UpdateScore(ctx context.Context, req *dto.UpdateScoreRequest) error {
	log.Info("PlayerRepository UpdateScore called playerID=%d gameID=%d score=%d", req.PlayerID, req.GameID, req.Score)
	_, err := r.db.Exec(ctx, "update scores set score = $1, updated_at = now() where player_id = $2 and game_id = $3", req.Score, req.PlayerID, req.GameID)
	if err != nil {
		log.Error("PlayerRepository UpdateScore failed playerID=%d gameID=%d err=%v", req.PlayerID, req.GameID, err)
		return err
	}
	log.Info("PlayerRepository UpdateScore success playerID=%d gameID=%d", req.PlayerID, req.GameID)
	return nil
}

func (r *postgresPlayerRepository) GetByID(ctx context.Context, id int) (*dto.PlayerResponse, error) {
	log.Info("PlayerRepository GetByID called id=%d", id)
	var player dto.PlayerResponse
	if err := r.db.
		QueryRow(ctx, "select id, username, created_at, updated_at from players where id = $1", id).
		Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("PlayerRepository GetByID not found id=%d", id)
			return nil, internalErrors.NewNotFound("player not found", err)
		}
		log.Error("PlayerRepository GetByID failed id=%d err=%v", id, err)
		return nil, err
	}
	log.Info("PlayerRepository GetByID success id=%d", id)
	return &player, nil
}

func (r *postgresPlayerRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	log.Info("PlayerRepository GetAll called page=%d pageSize=%d", params.Page, params.PageSize)
	rows, err := r.db.Query(ctx, "select id, username, created_at, updated_at from players limit $1 offset $2", params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		log.Error("PlayerRepository GetAll query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var players []dto.PlayerResponse
	for rows.Next() {
		var player dto.PlayerResponse
		if err := rows.Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt); err != nil {
			log.Error("PlayerRepository GetAll scan failed: %v", err)
			return nil, err
		}
		players = append(players, player)
	}

	totalCount, err := r.Count(ctx)
	if err != nil {
		log.Error("PlayerRepository GetAll count failed: %v", err)
		return nil, err
	}

	log.Info("PlayerRepository GetAll success items=%d total=%d", len(players), totalCount)
	return &dto.PaginatedResponse{
		Items:      players,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalCount,
	}, nil
}

func (r *postgresPlayerRepository) Count(ctx context.Context) (int, error) {
	log.Info("PlayerRepository Count called")
	var count int
	if err := r.db.QueryRow(ctx, "select count(*) from players").Scan(&count); err != nil {
		log.Error("PlayerRepository Count failed: %v", err)
		return 0, err
	}
	log.Info("PlayerRepository Count success count=%d", count)
	return count, nil
}
