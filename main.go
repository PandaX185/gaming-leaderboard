package main

import (
	"context"
	"gaming-leaderboard/internal/db"
	"gaming-leaderboard/internal/handler"
	"gaming-leaderboard/internal/model"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	config "gaming-leaderboard/internal/server"
	"gaming-leaderboard/internal/service"
	"gaming-leaderboard/internal/worker"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	db, err := db.Init(os.Getenv("DB_URI"))
	if err != nil {
		panic("Error initializing database")
	}
	defer db.Disconnect(context.Background())

	dbName := db.Database(os.Getenv("DB_NAME"))
	model.CreateIndexes(context.Background(), dbName)
	srv := config.NewServer(os.Getenv("PORT"))
	apiPrefix := srv.Srv.Group("/api/v1")

	playerRepo := repository.NewMongoPlayerRepository(dbName)
	playerQueue := queue.NewPlayerQueue(playerRepo)

	maxWorkers, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		maxWorkers = 3
	}
	log.Printf("Starting %d player workers...", maxWorkers)

	for range maxWorkers {
		worker := worker.NewPlayerWorker(playerQueue).SetMaxRetries(3)
		go worker.Start(context.Background())
	}
	playerService := service.NewPlayerService(playerRepo, playerQueue)
	playerHandler := handler.NewPlayerHandler(playerService, apiPrefix)
	playerHandler.RegisterRoutes()

	gameRepo := repository.NewMongoGameRepository(dbName)
	gameService := service.NewGameService(gameRepo)
	gameHandler := handler.NewGameHandler(gameService, apiPrefix)
	gameHandler.RegisterRoutes()

	srv.Run()
}
