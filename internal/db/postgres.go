package db

import (
	"context"
	"os"
	"time"

	"gaming-leaderboard/internal/db/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPostgres() *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(os.Getenv("DB_URI"))
	if err != nil {
		panic(err)
	}
	config.MaxConns = 200
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	var conn *pgxpool.Pool
	var lastErr error
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			lastErr = conn.Ping(ctx)
			if lastErr == nil {
				return conn
			}
			conn.Close()
		} else {
			lastErr = err
		}
		time.Sleep(2 * time.Second)
	}
	panic(lastErr)
}

func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	if err := migrations.Migrate00001(ctx, db); err != nil {
		return err
	}
	if err := migrations.Migrate00002(ctx, db); err != nil {
		return err
	}
	if err := migrations.Migrate00003(ctx, db); err != nil {
		return err
	}
	return nil
}
