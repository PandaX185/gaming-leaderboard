package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate00003(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS scores (
	    player_id UUID NOT NULL,
	    game_id INT NOT NULL,
	    score INT NOT NULL,
	    created_at TIMESTAMP DEFAULT now(),
	    updated_at TIMESTAMP DEFAULT now(),

	    PRIMARY KEY (player_id, game_id),
	    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
	    FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
		CONSTRAINT check_score CHECK (score >= 0)
	);
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
	CREATE INDEX IF NOT EXISTS idx_scores_score ON scores(score DESC);
	CREATE INDEX IF NOT EXISTS idx_scores_game_id ON scores(game_id);
	`)
	return err
}
