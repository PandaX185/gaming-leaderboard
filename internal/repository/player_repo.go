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
	Insert(context.Context, *dto.CreatePlayerRequest) error
	UpdateScore(context.Context, *dto.UpdateScoreRequest) error
	GetByID(context.Context, string) (*dto.PlayerResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Count(ctx context.Context) (int, error)
}

type postgresPlayerRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPlayerRepository(db *pgxpool.Pool) PlayerRepository {
	return &postgresPlayerRepository{db: db}
}

func (r *postgresPlayerRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, "select count(*) from players").Scan(&count); err != nil {
		log.Error("PlayerRepository Count failed: %v", err)
		return 0, err
	}
	return count, nil
}

func (r *postgresPlayerRepository) Insert(ctx context.Context, req *dto.CreatePlayerRequest) error {
	_, err := r.db.Exec(ctx, "insert into players(id, username, password, created_at, updated_at) values ($1, $2, $3, $4, $5)", req.ID, req.Username, req.Password, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		log.Error("PlayerRepository Insert failed id=%s username=%s err=%v", req.ID, req.Username, err)
		return err
	}
	return nil
}

func (r *postgresPlayerRepository) UpdateScore(ctx context.Context, req *dto.UpdateScoreRequest) error {
	_, err := r.db.Exec(ctx, "update scores set score = $1, updated_at = now() where player_id = $2 and game_id = $3", req.Score, req.PlayerID, req.GameID)
	if err != nil {
		log.Error("PlayerRepository UpdateScore failed playerID=%s gameID=%s err=%v", req.PlayerID, req.GameID, err)
		return err
	}
	return nil
}

func (r *postgresPlayerRepository) GetByID(ctx context.Context, id string) (*dto.PlayerResponse, error) {
	var player dto.PlayerResponse
	if err := r.db.
		QueryRow(ctx, "select id, username, created_at, updated_at from players where id = $1", id).
		Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("PlayerRepository GetByID not found id=%s", id)
			return nil, internalErrors.NewNotFound("player not found", err)
		}
		log.Error("PlayerRepository GetByID failed id=%s err=%v", id, err)
		return nil, err
	}
	return &player, nil
}

func (r *postgresPlayerRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	rows, err := r.db.
		Query(ctx, "select id, username, created_at, updated_at from players order by updated_at desc limit $1 offset $2",
			params.PageSize, (params.Page-1)*params.PageSize)
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

	return &dto.PaginatedResponse{
		Items:    players,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}
