package worker

import (
	"context"
	"gaming-leaderboard/internal/repository"
	"log"
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
		log.Printf("Leaderboard rebuild failed: %v", err)
		return
	}

	log.Printf("Leaderboard rebuild completed in %s", time.Since(start).String())
}
