package repository

import (
	"context"
	"gaming-leaderboard/internal/dto"

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
	var id int
	err := r.db.QueryRow(ctx, "insert into players(username, password, created_at, updated_at) values ($1, $2, $3, $4) returning id", req.Username, req.Password, req.CreatedAt, req.UpdatedAt).Scan(&id)
	return id, err
}

func (r *postgresPlayerRepository) UpdateScore(ctx context.Context, req *dto.UpdateScoreRequest) error {
	_, err := r.db.Exec(ctx, "update scores set score = $1, updated_at = now() where player_id = $2 and game_id = $3", req.Score, req.PlayerID, req.GameID)
	return err
}

func (r *postgresPlayerRepository) GetByID(ctx context.Context, id int) (*dto.PlayerResponse, error) {
	var player dto.PlayerResponse
	if err := r.db.
		QueryRow(ctx, "select id, username, created_at, updated_at from players where id = $1", id).
		Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt); err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *postgresPlayerRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	rows, err := r.db.Query(ctx, "select id, username, created_at, updated_at from players limit $1 offset $2", params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []dto.PlayerResponse
	for rows.Next() {
		var player dto.PlayerResponse
		if err := rows.Scan(&player.ID, &player.Username, &player.CreatedAt, &player.UpdatedAt); err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	totalCount, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse{
		Items:      players,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalCount,
	}, nil
}

func (r *postgresPlayerRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRow(ctx, "select count(*) from players").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
