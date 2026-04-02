package worker

import (
	"context"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/internal/repository"
	"time"
)

func RebuildLeaderboardsOnStartup(scoreRepo repository.ScoreRepository, playerRepo repository.PlayerRepository, gameRepo repository.GameRepository, cache repository.LeaderboardCache) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	if err := cache.RebuildFromDb(ctx, scoreRepo, gameRepo, playerRepo); err != nil {
		log.Error("Leaderboard rebuild failed: %v", err)
		return
	}

	log.Info("Leaderboard rebuild completed in %s", time.Since(start).String())
}
