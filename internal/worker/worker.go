package worker

import (
	"context"
	"gaming-leaderboard/internal/consts"
	"gaming-leaderboard/internal/dto"
	"gaming-leaderboard/internal/log"
	"gaming-leaderboard/internal/queue"
	"gaming-leaderboard/internal/repository"
	"gaming-leaderboard/metrics"
	"os"
	"strconv"
)

type Worker struct {
	qu         queue.IQueue
	maxRetries int
	playerRepo repository.PlayerRepository
	scoreRepo  repository.ScoreRepository
	cache      repository.LeaderboardCache
}

func NewWorker(qu queue.IQueue) *Worker {
	return &Worker{
		qu: qu,
	}
}

func (w *Worker) SetMaxRetries(retries int) *Worker {
	w.maxRetries = retries
	return w
}

func (w *Worker) Start(ctx context.Context) {
	poolSize, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		poolSize = 100
	}

	eventsChan := w.qu.GetEvents()
	for i := 0; i < poolSize; i++ {
		go func() {
			for {
				select {
				case event := <-eventsChan:
					event.Handler = w.getEventHandler(event)
					metrics.WorkerInFlight.Inc()
					if err := event.Handler(ctx, event.Payload); err != nil {
						log.Error("Failed to handle the event %v with error %v", event, err)
						metrics.WorkerErrors.Inc()
						if event.Attempt < w.maxRetries {
							log.Info("Attempt %d failed, retrying...", event.Attempt+1)
							event.Attempt++
							metrics.WorkerRetriesTotal.WithLabelValues(event.Type).Inc()
						} else {
							log.Warn("Max retries reached for event %s, discarding", event.Type)
							if event.Ack != nil {
								if ackErr := event.Ack(ctx); ackErr != nil {
									metrics.QueueAckTotal.WithLabelValues(event.Type, "error").Inc()
								} else {
									metrics.QueueAckTotal.WithLabelValues(event.Type, "success").Inc()
								}
							}
						}
					} else {
						log.Info("Event %v handled successfully", event)
						if event.Ack != nil {
							if ackErr := event.Ack(ctx); ackErr != nil {
								metrics.QueueAckTotal.WithLabelValues(event.Type, "error").Inc()
							} else {
								metrics.QueueAckTotal.WithLabelValues(event.Type, "success").Inc()
							}
						}
						metrics.WorkerProcessed.Inc()
					}
					metrics.WorkerInFlight.Dec()

				case <-ctx.Done():
					log.Info("Worker shutting down...")
					return
				}
			}
		}()
	}
	<-ctx.Done()
}

func (w *Worker) getEventHandler(event queue.Event) func(ctx context.Context, payload any) error {
	switch event.Type {
	case consts.PlayerCreatedEvent:
		return func(workerCtx context.Context, p any) error {
			if err := w.playerRepo.Insert(workerCtx, p.(*dto.CreatePlayerRequest)); err != nil {
				return err
			}
			return w.cache.IncrementPlayerCount(workerCtx)
		}
	case consts.ScoreUpdatedEvent:
		return func(workerCtx context.Context, p any) error {
			data := p.(*dto.UpdateScoreEvent)
			if err := w.scoreRepo.UpdateScore(workerCtx, data.PlayerID, data.GameID, data.Score); err != nil {
				return err
			}
			return w.cache.IncrementScore(workerCtx, data.PlayerID, data.GameID, data.Score)
		}
	default:
		return func(ctx context.Context, payload any) error {
			log.Warn("No handler for event type: %s", event.Type)
			return nil
		}
	}
}
