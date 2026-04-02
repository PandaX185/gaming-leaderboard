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
	log.Info("GameRepository Insert called name=%s", req.Name)
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "insert into games(name) values ($1) returning *", req.Name).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		log.Error("GameRepository Insert failed name=%s err=%v", req.Name, err)
		return nil, err
	}
	log.Info("GameRepository Insert success id=%d name=%s", game.ID, game.Name)
	return &game, nil
}

func (r *postgresGameRepository) GetByID(ctx context.Context, id int) (*dto.GameResponse, error) {
	log.Info("GameRepository GetByID called id=%d", id)
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "select id, name, created_at, updated_at from games where id = $1", id).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("GameRepository GetByID not found id=%d", id)
			return nil, internalErrors.NewNotFound("game not found", err)
		}
		log.Error("GameRepository GetByID failed id=%d err=%v", id, err)
		return nil, err
	}
	log.Info("GameRepository GetByID success id=%d", id)
	return &game, nil
}

func (r *postgresGameRepository) GetAll(ctx context.Context, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	log.Info("GameRepository GetAll called page=%d pageSize=%d", params.Page, params.PageSize)
	rows, err := r.db.Query(ctx, "select id, name, created_at, updated_at from games limit $1 offset $2", params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		log.Error("GameRepository GetAll query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var games []dto.GameResponse
	for rows.Next() {
		var game dto.GameResponse
		if err := rows.Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
			log.Error("GameRepository GetAll scan failed: %v", err)
			return nil, err
		}
		games = append(games, game)
	}

	var totalCount int
	if err := r.db.QueryRow(ctx, "select count(*) from games").Scan(&totalCount); err != nil {
		log.Error("GameRepository GetAll count failed: %v", err)
		return nil, err
	}

	log.Info("GameRepository GetAll success items=%d total=%d", len(games), totalCount)
	return &dto.PaginatedResponse{
		Items:      games,
		TotalItems: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (r *postgresGameRepository) GetScores(ctx context.Context, gameID int, params *dto.PaginationParams) (*dto.PaginatedResponse, error) {
	log.Info("GameRepository GetScores called gameID=%d page=%d pageSize=%d", gameID, params.Page, params.PageSize)
	rows, err := r.db.Query(ctx, "select player_id, score, created_at, updated_at from scores where game_id = $1 limit $2 offset $3", gameID, params.PageSize, (params.Page-1)*params.PageSize)
	if err != nil {
		log.Error("GameRepository GetScores query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var scores []dto.ScoreResponse
	for rows.Next() {
		var score dto.ScoreResponse
		if err := rows.Scan(&score.PlayerID, &score.Score, &score.CreatedAt, &score.UpdatedAt); err != nil {
			log.Error("GameRepository GetScores scan failed: %v", err)
			return nil, err
		}
		scores = append(scores, score)
	}

	var totalCount int
	if err := r.db.QueryRow(ctx, "select count(*) from scores where game_id = $1", gameID).Scan(&totalCount); err != nil {
		log.Error("GameRepository GetScores count failed: %v", err)
		return nil, err
	}

	log.Info("GameRepository GetScores success gameID=%d items=%d total=%d", gameID, len(scores), totalCount)
	return &dto.PaginatedResponse{
		Items:      scores,
		TotalItems: totalCount,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func (r *postgresGameRepository) Update(ctx context.Context, id int, req *dto.UpdateGameRequest) (*dto.GameResponse, error) {
	log.Info("GameRepository Update called id=%d name=%s", id, req.Name)
	var game dto.GameResponse
	if err := r.db.
		QueryRow(ctx, "update games set name = $1, updated_at = now() where id = $2 returning id, name, created_at, updated_at", req.Name, id).
		Scan(&game.ID, &game.Name, &game.CreatedAt, &game.UpdatedAt); err != nil {
		log.Error("GameRepository Update failed id=%d err=%v", id, err)
		return nil, err
	}
	log.Info("GameRepository Update success id=%d name=%s", id, game.Name)
	return &game, nil
}

func (r *postgresGameRepository) Delete(ctx context.Context, id int) error {
	log.Info("GameRepository Delete called id=%d", id)
	if _, err := r.db.Exec(ctx, "delete from games where id = $1", id); err != nil {
		log.Error("GameRepository Delete failed id=%d err=%v", id, err)
		return err
	}
	log.Info("GameRepository Delete success id=%d", id)
	return nil
}
