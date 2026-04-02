package worker

import (
	"context"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/internal/repository"
	"time"
)

func RebuildLeaderboardsOnStartup(repo repository.ScoreRepository, cache repository.LeaderboardCache) {
	if cache == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	if err := cache.RebuildFromDb(ctx, repo); err != nil {
		log.Error("Leaderboard rebuild failed: %v", err)
		return
	}

	log.Info("Leaderboard rebuild completed in %s", time.Since(start).String())
}
