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

	"github.com/gin-contrib/pprof"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	dbInstance, err := db.Init(os.Getenv("DB_URI"))
	if err != nil {
		panic("Error initializing database")
	}
	defer dbInstance.Disconnect(context.Background())

	dbName := dbInstance.Database(os.Getenv("DB_NAME"))
	model.CreateIndexes(context.Background(), dbName)
	srv := config.NewServer(os.Getenv("PORT"))
	pprof.Register(srv.Srv)
	apiPrefix := srv.Srv.Group("/api/v1")

	playerRepo := repository.NewMongoPlayerRepository(dbName)
	redisURI := os.Getenv("REDIS_URI")
	redisClient, err := db.InitRedis(redisURI)
	if err != nil {
		log.Printf("Error connecting to Redis: %v. Continuing without cache.\n", err)
	} else {
		log.Println("Connected to Redis successfully")
		playerRepo = repository.NewCachedPlayerRepository(playerRepo, redisClient)
		log.Println("Redis cache enabled for player repository")
	}
	defer func() {
		if redisClient != nil {
			redisClient.Close()
		}
	}()

	playerQueue := queue.NewPlayerQueue(playerRepo)
	log.Println("Starting player worker pool...")
	playerWorker := worker.NewPlayerWorker(playerQueue).SetMaxRetries(0)
	go playerWorker.Start(context.Background())
	
	scoreQueue := queue.NewPlayerQueue(playerRepo)
	log.Println("Starting score worker pool...")
	scoreWorker := worker.NewPlayerWorker(scoreQueue).SetMaxRetries(0)
	go scoreWorker.Start(context.Background())


	playerService := service.NewPlayerService(playerRepo, playerQueue)
	playerHandler := handler.NewPlayerHandler(playerService, apiPrefix)
	playerHandler.RegisterRoutes()

	gameRepo := repository.NewMongoGameRepository(dbName)
	gameService := service.NewGameService(gameRepo)
	gameHandler := handler.NewGameHandler(gameService, apiPrefix)
	gameHandler.RegisterRoutes()

	srv.Run()
}
