package main

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/db"
	"gaming-leaderboard/internal/handler"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/realtime"
	"gaming-leaderboard/internal/repository"
	config "gaming-leaderboard/internal/server"
	"gaming-leaderboard/internal/service"
	"gaming-leaderboard/internal/worker"
	"gaming-leaderboard/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn("godotenv load failed, falling back to system environment variables")
	}

	metrics.Init()

	srv := config.NewServer(os.Getenv("PORT"))
	pprof.Register(srv.Srv)
	apiPrefix := srv.Srv.Group("/api/v1")

	dbInstance := db.InitPostgres()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	log.Info("Applying database migrations")
	if err := db.Migrate(ctx, dbInstance); err != nil {
		log.Error("Failed to apply migrations: %v", err)
		panic(err)
	}
	log.Info("Database migrations applied successfully")

	redisURI := os.Getenv("REDIS_URI")
	redisClient := db.InitRedis(redisURI)
	log.Info("Connected to Redis successfully")

	playerRepo := repository.NewPostgresPlayerRepository(dbInstance)
	gameRepo := repository.NewPostgresGameRepository(dbInstance)
	scoreRepo := repository.NewPostgresScoreRepository(dbInstance)

	leaderboardCache := repository.NewRedisLeaderboardCache(redisClient)
	worker.RebuildLeaderboardsOnStartup(scoreRepo, playerRepo, gameRepo, leaderboardCache)

	leaderboardHub := realtime.NewLeaderboardHub(redisClient)
	leaderboardWSHandler := handler.NewLeaderboardWSHandler(leaderboardHub, apiPrefix)
	leaderboardWSHandler.RegisterRoutes()

	defer func() {
		dbInstance.Close()
		redisClient.Close()
		log.Info("Database connections closed")
	}()

	playerQueue := queue.NewRedisQueue(redisClient, consts.PlayerEvents, consts.PlayerConsumerGroup)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	playerWorker := worker.NewWorker(playerQueue, playerRepo, leaderboardCache).SetMaxRetries(5)
	go playerWorker.Start(workerCtx)

	scoreQueue := queue.NewRedisQueue(redisClient, consts.ScoreEvents, consts.ScoreConsumerGroup)
	scoreWorker := worker.NewWorker(scoreQueue, scoreRepo, leaderboardCache).SetMaxRetries(5)
	go scoreWorker.Start(workerCtx)

	playerService := service.NewPlayerService(playerRepo, playerQueue, leaderboardCache)
	playerHandler := handler.NewPlayerHandler(playerService, apiPrefix)
	playerHandler.RegisterRoutes()

	scoreService := service.NewScoreService(scoreRepo, scoreQueue)
	scoreHandler := handler.NewScoreHandler(scoreService, apiPrefix)
	scoreHandler.RegisterRoutes()

	gameService := service.NewGameService(gameRepo, leaderboardCache)
	gameHandler := handler.NewGameHandler(gameService, apiPrefix)
	gameHandler.RegisterRoutes()

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	} else if addr[0] != ':' {
		addr = ":" + addr
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Srv,
	}

	go func() {
		log.Info("Server listening on port %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("ListenAndServe err: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Canceling workers...")
	workerCancel()

	time.Sleep(2 * time.Second)

	log.Info("Server exiting gracefully")
}
