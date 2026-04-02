package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate00001(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS players (
	    id SERIAL PRIMARY KEY,
	    username VARCHAR(255) NOT NULL,
		password TEXT NOT NULL,
	    created_at TIMESTAMP DEFAULT now(),
	    updated_at TIMESTAMP DEFAULT now(),

	    CONSTRAINT unique_username UNIQUE (username)
		);
	`)
	if err != nil {
		return err
	}
	return nil
}
