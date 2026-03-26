package worker

import (
	"context"
	"gaming-leaderboard/internal/queue"
	"log"
	"os"
	"strconv"
)

type PlayerWorker struct {
	playerQ    queue.IPlayerQueue
	maxRetries int
}

func NewPlayerWorker(playerQ queue.IPlayerQueue) *PlayerWorker {
	return &PlayerWorker{
		playerQ: playerQ,
	}
}

func (w *PlayerWorker) SetMaxRetries(retries int) *PlayerWorker {
	w.maxRetries = retries
	return w
}

func (w *PlayerWorker) Start(ctx context.Context) {
	poolSize, err := strconv.Atoi(os.Getenv("WORKER_COUNT"))
	if err != nil {
		poolSize = 100
	}
	for i := 0; i < poolSize; i++ {
		go func() {
			for {
				select {
				case event := <-w.playerQ.GetEvents():
					if err := event.Handler(ctx, event.Payload); err != nil {
						if event.Attempt < w.maxRetries {
							log.Printf("Attempt %d failed, retrying...\n", event.Attempt+1)
							event.Attempt++
							w.playerQ.GetEvents() <- event
						} else {
							log.Printf("Max retries reached for event %s, discarding", event.Type)
						}
					}

				case <-ctx.Done():
					log.Println("Player worker shutting down...")
					return
				}
			}
		}()
	}
	<-ctx.Done()
}
