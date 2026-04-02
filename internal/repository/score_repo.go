package repository

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"iter"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository interface {
	UpdateScore(context.Context, string, string, int) error
	GetAllLeaderboards(ctx context.Context) (iter.Seq[dto.ScoreResponse], error)
}

type postgresScoreRepository struct {
	db *pgxpool.Pool
}

func NewPostgresScoreRepository(db *pgxpool.Pool) ScoreRepository {
	return &postgresScoreRepository{db: db}
}

func (r *postgresScoreRepository) UpdateScore(ctx context.Context, gameID string, playerID string, delta int) error {
	_, err := r.db.Exec(ctx, `
		insert into scores (player_id, game_id, score)
		values ($1, $2, $3)
		on conflict (player_id, game_id) do update set score = scores.score + $3
	`, playerID, gameID, delta)
	return err
}

func (r *postgresScoreRepository) GetAllLeaderboards(ctx context.Context) (iter.Seq[dto.ScoreResponse], error) {
	rows, err := r.db.Query(ctx, "select player_id, game_id, score from scores")
	if err != nil {
		return nil, err
	}

	return func(yield func(dto.ScoreResponse) bool) {
		defer rows.Close()
		for rows.Next() {
			var score dto.ScoreResponse
			if err := rows.Scan(&score.PlayerID, &score.GameID, &score.Score); err != nil {
				return
			}
			if !yield(score) {
				return
			}
		}
	}, nil
}
