package worker

import (
	"context"
	"gaming-leaderboard/internal/repository"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RebuildLeaderboardsOnStartup(db *mongo.Database, cache repository.LeaderboardCache) {
	if cache == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	if err := cache.RebuildFromMongo(ctx, db); err != nil {
		log.Printf("Leaderboard rebuild failed: %v", err)
		return
	}

	log.Printf("Leaderboard rebuild completed in %s", time.Since(start).String())
}
