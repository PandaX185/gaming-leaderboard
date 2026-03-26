package worker

import (
	"context"
	"gaming-leaderboard/internal/queue"
	"log"
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
}
