package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate00002(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS games (
	    id SERIAL PRIMARY KEY,
	    name VARCHAR(255) NOT NULL,
	    created_at TIMESTAMP DEFAULT now(),
	    updated_at TIMESTAMP DEFAULT now()
	);
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
	CREATE INDEX IF NOT EXISTS idx_games_updated_at ON games(updated_at DESC);
	`)
	return err
}
