package repository

import (
	"context"
	"gaming-leaderboard/internal/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository interface {
	Insert(context.Context, *dto.CreateGameRequest) (*dto.GameResponse, error)
	GetByID(context.Context, int) (*dto.GameResponse, error)
	GetAll(context.Context, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	GetScores(context.Context, int, *dto.PaginationParams) (*dto.PaginatedResponse, error)
	Update(context.Context, int, *dto.UpdateGameRequest) (*dto.GameResponse, error)
	Delete(context.Context, int) error
}

type postgresGameRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGameRepository(db *pgxpool.Pool) GameRepository {
	return &postgresGameRepository{db: db}
}

func (r *postgresGameRepository) Insert(ctx context.Context, req *dto.CreateGameRequest) (*dto.GameResponse, error) {
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "insert into games(name) values ($1) returning *", req.Name).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *postgresGameRepository) GetByID(ctx context.Context, id int) (*dto.GameResponse, error) {
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "select id, name, created_at, updated_at from games where id = $1", id).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *postgresGameRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	rows, err := r.db.Query(ctx, "select id, name, created_at, updated_at from games limit $1 offset $2", params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []dto.GameResponse
	for rows.Next() {
		var game dto.GameResponse
		if err := rows.Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	var totalCount int
	if err := r.db.QueryRow(ctx, "select count(*) from games").Scan(&totalCount); err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse{
		Items:      games,
		TotalItems: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (r *postgresGameRepository) GetScores(ctx context.Context, gameID int, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	rows, err := r.db.Query(ctx, "select player_id, score, created_at, updated_at from scores where game_id = $1 limit $2 offset $3", gameID, params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []dto.ScoreResponse
	for rows.Next() {
		var score dto.ScoreResponse
		if err := rows.Scan(&score.PlayerID, &score.Score, &score.CreatedAt, &score.UpdatedAt); err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}

	var totalCount int
	if err := r.db.QueryRow(ctx, "select count(*) from scores where game_id = $1", gameID).Scan(&totalCount); err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse{
		Items:      scores,
		TotalItems: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (r *postgresGameRepository) Update(ctx context.Context, id int, req *dto.UpdateGameRequest) (*dto.GameResponse, error) {
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "update games set name = $1, updated_at = now() where id = $2 returning id, name, created_at, updated_at", req.Name, id).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *postgresGameRepository) Delete(ctx context.Context, id int) error {
	if _, err := r.db.Exec(ctx, "delete from games where id = $1", id); err != nil {
		return err
	}
	return nil
}
