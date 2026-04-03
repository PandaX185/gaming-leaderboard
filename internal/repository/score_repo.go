package repository

import (
	"context"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/log"
	"iter"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ScoreRepository interface {
	UpdateScore(context.Context, string, string, int) error
	GetAllLeaderboards(ctx context.Context) (iter.Seq[dto.ScoreResponse], error)
}

type scoreUpdate struct {	
	playerID string
	gameID   string
	delta    int
}

type postgresScoreRepository struct {
	db            *pgxpool.Pool
	scoreUpdates  chan scoreUpdate
	flushInterval time.Duration
}

func NewPostgresScoreRepository(db *pgxpool.Pool) ScoreRepository {
	r := &postgresScoreRepository{
		db:            db,
		scoreUpdates:  make(chan scoreUpdate, 2000),
		flushInterval: 100 * time.Millisecond,
	}
	go r.batchScoreLoop()
	return r
}

func (r *postgresScoreRepository) UpdateScore(ctx context.Context, gameID string, playerID string, delta int) error {
	select {
	case r.scoreUpdates <- scoreUpdate{playerID: playerID, gameID: gameID, delta: delta}:
		return nil
	default:
		_, err := r.db.Exec(ctx, `
			insert into scores (player_id, game_id, score)
			values ($1, $2, $3)
			on conflict (player_id, game_id) do update set score = scores.score + $3
		`, playerID, gameID, delta)
		if err != nil {
			log.Error("ScoreRepository UpdateScore fallback failed gameID=%s playerID=%s err=%v", gameID, playerID, err)
		}
		return err
	}
}

func (r *postgresScoreRepository) batchScoreLoop() {
	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()

	var queue []scoreUpdate
	for {
		select {
		case upd := <-r.scoreUpdates:
			queue = append(queue, upd)
			if len(queue) >= 1000 {
				r.flushScoreUpdates(queue)
				queue = queue[:0]
			}
		case <-ticker.C:
			if len(queue) > 0 {
				r.flushScoreUpdates(queue)
				queue = queue[:0]
			}
		}
	}
}

func (r *postgresScoreRepository) flushScoreUpdates(queue []scoreUpdate) {
	if len(queue) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("ScoreRepository batch begin tx failed: %v", err)
		return
	}

	for _, u := range queue {
		_, err := tx.Exec(ctx, `
			insert into scores (player_id, game_id, score)
			values ($1, $2, $3)
			on conflict (player_id, game_id) do update set score = scores.score + $3
		`, u.playerID, u.gameID, u.delta)
		if err != nil {
			log.Error("ScoreRepository batch update failed playerID=%s gameID=%s err=%v", u.playerID, u.gameID, err)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error("ScoreRepository batch rollback failed: %v", rbErr)
			}
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("ScoreRepository batch commit failed: %v", err)
		return
	}
}

func (r *postgresScoreRepository) GetAllLeaderboards(ctx context.Context) (iter.Seq[dto.ScoreResponse], error) {
	rows, err := r.db.Query(ctx, "select player_id, game_id, score from scores")
	if err != nil {
		log.Error("ScoreRepository GetAllLeaderboards query failed: %v", err)
		return nil, err
	}

	return func(yield func(dto.ScoreResponse) bool) {
		defer rows.Close()
		for rows.Next() {
			var score dto.ScoreResponse
			if err := rows.Scan(&score.PlayerID, &score.GameID, &score.Score); err != nil {
				log.Error("ScoreRepository GetAllLeaderboards scan failed: %v", err)
				return
			}
			if !yield(score) {
				return
			}
		}
	}, nil
}
