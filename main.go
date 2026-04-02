package main

import (
	"context"
	"gaming-leaderboard/internal/db"
	"gaming-leaderboard/internal/handler"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/realtime"
	"gaming-leaderboard/internal/repository"
	config "gaming-leaderboard/internal/server"
	"gaming-leaderboard/internal/service"
	"gaming-leaderboard/internal/worker"
	"gaming-leaderboard/metrics"
	"log"
	"os"

	"github.com/gin-contrib/pprof"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	metrics.Init()

	srv := config.NewServer(os.Getenv("PORT"))
	pprof.Register(srv.Srv)
	apiPrefix := srv.Srv.Group("/api/v1")

	dbInstance := db.InitPostgres()
	if err := db.Migrate(context.Background(), dbInstance); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	playerRepo := repository.NewPostgresPlayerRepository(dbInstance)

	scoreRepo := repository.NewPostgresScoreRepository(dbInstance)

	redisURI := os.Getenv("REDIS_URI")
	var leaderboardCache repository.LeaderboardCache
	redisClient, err := db.InitRedis(redisURI)
	if err != nil {
		log.Printf("Error connecting to Redis: %v. Continuing without leaderboard cache.\n", err)
	} else {
		log.Println("Connected to Redis successfully")
		leaderboardCache = repository.NewRedisLeaderboardCache(redisClient)
		worker.RebuildLeaderboardsOnStartup(scoreRepo, leaderboardCache)

		leaderboardHub := realtime.NewLeaderboardHub(redisClient)
		leaderboardWSHandler := handler.NewLeaderboardWSHandler(leaderboardHub, apiPrefix)
		leaderboardWSHandler.RegisterRoutes()
	}
	defer func() {
		if redisClient != nil {
			redisClient.Close()
		}
	}()

	playerQueue := queue.NewQueue(os.Getenv("QUEUE_TYPE"), playerRepo, redisClient, leaderboardCache)
	log.Println("Starting player worker pool...")
	playerWorker := worker.NewWorker(playerQueue).SetMaxRetries(5)
	go playerWorker.Start(context.Background())

	scoreQueue := queue.NewQueue(os.Getenv("QUEUE_TYPE"), scoreRepo, redisClient, leaderboardCache)
	log.Println("Starting score worker pool...")
	scoreWorker := worker.NewWorker(scoreQueue).SetMaxRetries(5)
	go scoreWorker.Start(context.Background())

	playerService := service.NewPlayerService(playerRepo, playerQueue)
	playerHandler := handler.NewPlayerHandler(playerService, apiPrefix)
	playerHandler.RegisterRoutes()

	scoreService := service.NewScoreService(scoreQueue)
	scoreHandler := handler.NewScoreHandler(scoreService, apiPrefix)
	scoreHandler.RegisterRoutes()

	gameRepo := repository.NewPostgresGameRepository(dbInstance)
	gameService := service.NewGameService(gameRepo)
	gameHandler := handler.NewGameHandler(gameService, apiPrefix)
	gameHandler.RegisterRoutes()

	srv.Run()
}
