package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate00001(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";
	CREATE TABLE IF NOT EXISTS players (
	    id UUID PRIMARY KEY default gen_random_uuid(),
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
